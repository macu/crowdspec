package main

import (
	"database/sql"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"strings"
	"time"
)

const signupRequestTokenLength = 15

func makeRequestSignupHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {

	var requestSignupPageTemplate *template.Template

	var executeTemplate = func(w http.ResponseWriter, r *http.Request,
		username, email, message string, statusCode int, errMsg string, err error,
	) {
		if err != nil {
			logError(r, 0, err)
		}
		if statusCode != 0 {
			w.WriteHeader(statusCode)
		}
		requestSignupPageTemplate.Execute(w, struct {
			Mode         string
			Error        string
			Username     string
			Email        string
			Message      string
			SiteKey      string
			Verify       bool // reCAPTCHA required
			VersionStamp string
		}{
			"request",
			errMsg,
			username,
			email,
			message,
			recaptchaSiteKey,
			isAppEngine(),
			cacheControlVersionStamp,
		})
	}

	return func(w http.ResponseWriter, r *http.Request) {

		if requestSignupPageTemplate == nil {
			requestSignupPageTemplate = template.Must(template.ParseFiles("html/signup.html"))
		}

		if r.Method == http.MethodGet {

			executeTemplate(w, r, "", "", "", http.StatusOK, "", nil)

		} else if r.Method == http.MethodPost {

			username := strings.TrimSpace(r.FormValue("username"))
			email := strings.TrimSpace(r.FormValue("email"))
			message := r.FormValue("message")

			// Validate reCAPTCHA
			valid, err := verifyRecaptcha(r)
			if !valid {
				if err != nil {
					executeTemplate(w, r,
						username, email, message,
						http.StatusInternalServerError, "Server error",
						fmt.Errorf("validating recaptcha: %w", err))
					return
				}
				// Use Teapot to indicate reCAPTCHA error
				executeTemplate(w, r,
					username, email, message,
					http.StatusTeapot, "Invalid reCAPTCHA",
					// fmt.Errorf("invalid reCAPTCHA [IP %s]", getUserIP(r)))
					fmt.Errorf("invalid reCAPTCHA"))
				return
			}

			if !usernameRegex.MatchString(username) {
				executeTemplate(w, r,
					username, email, message,
					http.StatusBadRequest,
					"Invalid username (a-z A-Z 0-9 _ allowed)",
					nil) // don't log error
				return
			}

			if len(username) > usernameMaxLength {
				executeTemplate(w, r,
					username, email, message,
					http.StatusBadRequest,
					fmt.Sprintf("Maximum username length is %d characters", usernameMaxLength), nil)
				return
			}

			if !emailRegexp.MatchString(email) {
				executeTemplate(w, r,
					username, email, message,
					http.StatusBadRequest,
					"Invalid email address",
					nil) // don't log error
				return
			}

			if len(strings.TrimSpace(email)) > emailAddressMaxLength {
				executeTemplate(w, r,
					username, email, message,
					http.StatusBadRequest,
					fmt.Sprintf(
						"Maximum email address length is %d characters",
						emailAddressMaxLength),
					nil) // dont log error
				return
			}

			var usernameTaken bool
			err = db.QueryRow(
				`SELECT EXISTS(
					SELECT r.id
					FROM user_signup_request r
					WHERE r.username = $1
					UNION
					SELECT u.id
					FROM user_account u
					WHERE u.username = $1
				)`,
				username,
			).Scan(&usernameTaken)
			if err != nil {
				executeTemplate(w, r,
					username, email, message,
					http.StatusInternalServerError,
					"Server error", err)
				return
			}

			if usernameTaken {
				executeTemplate(w, r,
					username, email, message,
					http.StatusBadRequest,
					"Username already claimed",
					nil) // don't log error
				return
			}

			var emailTaken bool
			err = db.QueryRow(
				`SELECT EXISTS(
					SELECT r.id
					FROM user_signup_request r
					WHERE r.email = $1
					UNION
					SELECT u.id
					FROM user_account u
					WHERE u.email = $1
				)`,
				email,
			).Scan(&emailTaken)
			if err != nil {
				executeTemplate(w, r,
					username, email, message,
					http.StatusInternalServerError,
					"Server error", err)
				return
			}

			if emailTaken {
				executeTemplate(w, r,
					username, email, message,
					http.StatusBadRequest,
					"Email address already claimed",
					nil) // don't log error
				return
			}

			logNotice(r, struct {
				Event        string
				Username     string
				EmailAddress string
				// IPAddress    string
			}{
				"SignupRequest",
				username,
				email,
				// getUserIP(r),
			})

			// save the request for review
			var requestID int64
			err = db.QueryRow(
				`INSERT INTO user_signup_request (username, email, created_at)
				VALUES ($1, $2, $3)
				RETURNING id`,
				username, email, time.Now(),
			).Scan(&requestID)
			if err != nil {
				executeTemplate(w, r,
					username, email, message,
					http.StatusInternalServerError,
					"Server error",
					fmt.Errorf("creating signup request: %w", err))
				return
			}

			// send email to admin first to exit early if email service not working
			var reqID = IntToA(int(requestID))
			err = sendEmail("Admin", "crowdspec.dev@gmail.com",
				"CrowdSpec signup request for "+username,
				"Request ID: "+reqID+"\n"+
					"Email address: "+email+"\n"+
					"Username: "+username+"\n"+
					"Message:\n\n"+message+"\n",
				"<table>\n"+
					"<tr><td>Request ID&emsp;</td><td>"+reqID+"</td></tr>\n"+
					"<tr><td>Email address&emsp;</td><td>"+html.EscapeString(email)+"</td></tr>\n"+
					"<tr><td>Username&emsp;</td><td>"+html.EscapeString(username)+"</td></tr>\n"+
					"<tr><td valign=\"top\" style=\"vertical-align:top;\">Message&emsp;</td><td>"+
					strings.ReplaceAll(html.EscapeString(message), "\n", "<br/>")+
					"</td></tr>\n"+
					"</table>\n",
			)
			if err != nil {
				executeTemplate(w, r,
					username, email, message,
					http.StatusInternalServerError,
					"Server error (email service is down)",
					fmt.Errorf("mailjet send signup request admin notification email: %w", err))
				return
			}

			// send email to user next to ensure this step passes for all signup requests
			err = sendEmail(username, email,
				"CrowdSpec signup request submitted for "+username,
				"Username: "+username+"\n"+
					"Message:\n\n"+message+"\n",
				"<table>\n"+
					"<tr><td>Username&emsp;</td><td>"+html.EscapeString(username)+"</td></tr>\n"+
					"<tr><td valign=\"top\" style=\"vertical-align:top;\">Message&emsp;</td><td>"+
					strings.ReplaceAll(html.EscapeString(message), "\n", "<br/>")+
					"</td></tr>\n"+
					"</table>\n"+
					"<br/>\n"+
					"<p>You'll receive another email with an activation link after I approve your username.</p>\n",
			)
			if err != nil {
				executeTemplate(w, r,
					username, email, message,
					http.StatusInternalServerError,
					"Server error",
					fmt.Errorf("mailjet send signup request user notification email: %w", err))
				return
			}

			requestSignupPageTemplate.Execute(w, struct {
				Mode         string
				VersionStamp string
			}{
				"success",
				cacheControlVersionStamp,
			})

		}

	}
}

func makeActivateSignupHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {

	var activateSignupPageTemplate *template.Template

	var executeTemplate = func(w http.ResponseWriter, r *http.Request,
		token string, username string, statusCode int, errMsg string, err error,
	) {
		if err != nil {
			logError(r, 0, err)
		}
		if statusCode != 0 {
			w.WriteHeader(statusCode)
		}
		activateSignupPageTemplate.Execute(w, struct {
			Mode         string
			Error        string
			Token        string
			Username     string
			SiteKey      string
			Verify       bool // reCAPTCHA required
			VersionStamp string
		}{
			"password",
			errMsg,
			token,
			username,
			recaptchaSiteKey,
			isAppEngine(),
			cacheControlVersionStamp,
		})
	}

	return func(w http.ResponseWriter, r *http.Request) {

		if activateSignupPageTemplate == nil {
			activateSignupPageTemplate = template.Must(template.ParseFiles("html/activate-signup.html"))
		}

		token := r.FormValue("t")

		if len(token) != signupRequestTokenLength {
			executeTemplate(w, r, "", "",
				http.StatusNotFound, "Invalid token",
				nil) // don't log error
			return
		}

		var username, email string
		err := db.QueryRow(
			`SELECT r.username, r.email
			FROM user_signup_request r
			WHERE r.token = $1 AND r.reviewed AND r.approved`, token,
		).Scan(&username, &email)
		if err != nil {
			// Token not found or expired
			executeTemplate(w, r, "", "",
				http.StatusNotFound, "Token not found",
				nil) // don't log error
			return
		}

		if r.Method == http.MethodGet {

			executeTemplate(w, r, token, username,
				http.StatusOK, "", nil)
			return

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

			pass := r.FormValue("password")
			pass2 := r.FormValue("password2")

			if len(strings.TrimSpace(pass)) < passwordMinLength {
				executeTemplate(w, r, token, username,
					http.StatusBadRequest,
					fmt.Sprintf("Password minimum length is %d digits", passwordMinLength),
					nil) // don't log error
				return
			}

			if pass != pass2 {
				executeTemplate(w, r, token, username,
					http.StatusForbidden,
					"Please enter the same password twice to confirm",
					nil) // don't log error
				return
			}

			// Create and authenticate user in a transaction
			err = inTransaction(r, db, func(tx *sql.Tx) error {

				userID, err := createUserTx(tx, username, pass, email)
				if err != nil {
					return err
				}

				logNotice(r, struct {
					Event    string
					Username string
					UserID   uint
					// IPAddress string
				}{
					"SignupActivate",
					username,
					userID,
					// getUserIP(r),
				})

				// Clear token and associate request with new user
				_, err = tx.Exec(
					`UPDATE user_signup_request
					SET token = NULL, user_id = $1
					WHERE token = $2`, userID, token)
				if err != nil {
					return err
				}

				err = authUser(w, r, tx, userID)
				if err != nil {
					return err
				}

				return nil
			})

			if err != nil {
				executeTemplate(w, r, token, username,
					http.StatusInternalServerError, "Server error", err)
				return
			}

			// Show success page
			activateSignupPageTemplate.Execute(w, struct {
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
