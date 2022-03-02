package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const resetPasswordTokenLength = 15
const resetPasswordTokenExpiry = time.Minute * 10

func makeRequestPasswordResetHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {

	var requestPasswordResetPageTemplate *template.Template

	var executeTemplate = func(w http.ResponseWriter, r *http.Request,
		statusCode int, errMsg string, err error,
	) {
		if err != nil {
			logError(r, nil, err)
		}
		if statusCode != 0 {
			w.WriteHeader(statusCode)
		}
		requestPasswordResetPageTemplate.Execute(w, struct {
			Mode         string
			Error        string
			SiteKey      string
			Verify       bool // reCAPTCHA required
			VersionStamp string
		}{
			"request",
			errMsg,
			recaptchaSiteKey,
			isAppEngine(),
			cacheControlVersionStamp,
		})
	}

	return func(w http.ResponseWriter, r *http.Request) {

		if requestPasswordResetPageTemplate == nil {
			requestPasswordResetPageTemplate = template.Must(template.ParseFiles("html/request-password-reset.html"))
		}

		if r.Method == http.MethodGet {

			executeTemplate(w, r, http.StatusOK, "", nil)

		} else if r.Method == http.MethodPost {

			email := strings.TrimSpace(r.FormValue("email"))

			if !emailRegexp.MatchString(email) {
				executeTemplate(w, r,
					http.StatusNotFound, "Invalid email address",
					nil) // don't log error
				return
			}

			// Validate reCAPTCHA
			valid, err := verifyRecaptcha(r)
			if !valid {
				if err != nil {
					executeTemplate(w, r,
						http.StatusInternalServerError, "Server error",
						fmt.Errorf("validating recaptcha: %w", err))
					return
				}
				// Use Teapot to indicate reCAPTCHA error
				executeTemplate(w, r,
					http.StatusTeapot, "Invalid reCAPTCHA",
					// fmt.Errorf("invalid reCAPTCHA [IP %s]", getUserIP(r)))
					fmt.Errorf("invalid reCAPTCHA"))
				return
			}

			var userID uint
			var username string
			err = db.QueryRow(
				`SELECT id, username FROM user_account WHERE email = $1`, email,
			).Scan(&userID, &username)
			if err != nil {
				if err == sql.ErrNoRows {
					executeTemplate(w, nil,
						http.StatusNotFound, "User not found",
						nil) // don't log error
					return
				}
				executeTemplate(w, nil,
					http.StatusInternalServerError, "Server error", err)
				return
			}

			// Create request

			logNotice(r, struct {
				Event  string
				UserID uint
				// IPAddress string
			}{
				"RequestResetPassword",
				userID,
				// getUserIP(r),
			})

			token := randomToken(resetPasswordTokenLength)
			_, err = db.Exec(
				`INSERT INTO password_reset_request
				(user_id, sent_to_address, token, created_at)
				VALUES ($1, $2, $3, current_timestamp)`,
				userID, email, token)
			if err != nil {
				executeTemplate(w, r,
					http.StatusInternalServerError, "Server error",
					fmt.Errorf("creating password_reset_request: %w", err))
				return
			}

			// Send email

			url, err := buildAbsoluteURL(r, "reset-password?t="+token)
			if err != nil {
				executeTemplate(w, r,
					http.StatusInternalServerError,
					"Server error",
					fmt.Errorf("building URL: %w", err))
				return
			}

			err = sendEmail(username, email,
				"CrowdSpec password reset for user "+username,
				"Visit the following link to reset your password: "+url,
				"<p>Visit the following link to reset your password: "+
					`<a href="`+url+`">`+url+`</a></p>`,
				false,
			)
			if err != nil {
				executeTemplate(w, r,
					http.StatusInternalServerError, "Server error",
					fmt.Errorf("mailjet send reset password email: %w", err))
				return
			}

			requestPasswordResetPageTemplate.Execute(w, struct {
				Mode         string
				Email        string
				VersionStamp string
			}{
				"sent",
				email,
				cacheControlVersionStamp,
			})

		}

	}
}

func makeResetPasswordHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {

	var resetPasswordPageTemplate *template.Template

	var executeTemplate = func(w http.ResponseWriter, r *http.Request,
		token, username string, errCode int, errMsg string, err error,
	) {
		if err != nil {
			logError(r, nil, err)
		}
		if errCode != 0 {
			w.WriteHeader(errCode)
		}
		resetPasswordPageTemplate.Execute(w, struct {
			Mode         string
			Token        string
			Username     string
			Error        string
			SiteKey      string
			Verify       bool // reCAPTCHA required
			VersionStamp string
		}{
			"reset",
			token,
			username,
			errMsg,
			recaptchaSiteKey,
			isAppEngine(),
			cacheControlVersionStamp,
		})
	}

	return func(w http.ResponseWriter, r *http.Request) {

		if resetPasswordPageTemplate == nil {
			resetPasswordPageTemplate = template.Must(template.ParseFiles("html/reset-password.html"))
		}

		token := r.FormValue("t")

		if len(token) != resetPasswordTokenLength {
			executeTemplate(w, r, "", "",
				http.StatusNotFound, "Invalid token",
				nil) // don't log error
			return
		}

		var createdAt time.Time
		var userID uint
		var username string
		err := db.QueryRow(
			`SELECT r.created_at, u.id, u.username
				FROM password_reset_request r
				INNER JOIN user_account u
					ON u.id = r.user_id
				WHERE r.token = $1`, token,
		).Scan(&createdAt, &userID, &username)
		if err != nil {
			if err == sql.ErrNoRows {
				// Token not found or expired
				executeTemplate(w, r, "", "",
					http.StatusNotFound, "Token not found",
					nil) // don't log error
				return
			}
			executeTemplate(w, r, "", "",
				http.StatusInternalServerError, "Server error", err)
			return
		}

		// Error if token expired
		if time.Now().After(createdAt.Add(resetPasswordTokenExpiry)) {
			executeTemplate(w, r, "", "",
				http.StatusNotFound,
				fmt.Sprintf(
					"Token expired - please reset password within %d minutes of request",
					resetPasswordTokenExpiry/time.Minute,
				),
				nil) // don't log error
			return
		}

		if r.Method == http.MethodGet {

			executeTemplate(w, r, token, username, http.StatusOK, "", nil)

		} else if r.Method == http.MethodPost {

			// Validate reCAPTCHA
			valid, err := verifyRecaptcha(r)
			if !valid {
				if err != nil {
					executeTemplate(w, r, token, username,
						http.StatusInternalServerError, "Server error",
						fmt.Errorf("validating recaptcha: %w", err))
					return
				}
				// Use Teapot to indicate reCAPTCHA error
				executeTemplate(w, r, token, username,
					http.StatusTeapot, "Invalid reCAPTCHA",
					// fmt.Errorf("invalid reCAPTCHA [IP %s]", getUserIP(r)))
					fmt.Errorf("invalid reCAPTCHA"))
				return
			}

			newpass := r.FormValue("newpass")
			newpass2 := r.FormValue("newpass2")

			if len(strings.TrimSpace(newpass)) < passwordMinLength {
				executeTemplate(w, r, token, username,
					http.StatusBadRequest,
					fmt.Sprintf("Password minimum length is %d digits", passwordMinLength),
					nil) // don't log error
				return
			} else if newpass != newpass2 {
				executeTemplate(w, r, token, username,
					http.StatusForbidden,
					"Please enter the same password twice to confirm",
					nil) // don't log error
				return
			}

			// Update user password

			logNotice(r, struct {
				Event  string
				UserID uint
				// IPAddress string
			}{
				"ResetPassword",
				userID,
				// getUserIP(r),
			})

			authHash, err := bcrypt.GenerateFromPassword([]byte(newpass), BcryptCost)
			if err != nil {
				executeTemplate(w, r, token, username,
					http.StatusInternalServerError,
					"Server error",
					fmt.Errorf("generating new auth hash: %w", err))
				return
			}

			// Clear token so request cannot be redeemed again
			_, err = db.Exec(
				`UPDATE password_reset_request SET token=NULL WHERE token=$1`, token)
			if err != nil {
				executeTemplate(w, r, token, username,
					http.StatusInternalServerError,
					"Server error",
					fmt.Errorf("updating password reset request: %w", err))
				return
			}

			_, err = db.Exec(
				`UPDATE user_account SET auth_hash=$2 WHERE id=$1`,
				userID, authHash)
			if err != nil {
				executeTemplate(w, r, token, username,
					http.StatusInternalServerError,
					"Server error",
					fmt.Errorf("updating user auth_hash: %w", err))
				return
			}

			// Authenticate user
			authUser(w, r, db, userID)

			resetPasswordPageTemplate.Execute(w, struct {
				Mode         string
				Token        string
				VersionStamp string
			}{
				"success",
				"",
				cacheControlVersionStamp,
			})

		}

	}
}
