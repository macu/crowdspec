package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	mailjet "github.com/mailjet/mailjet-apiv3-go/v3"
	"golang.org/x/crypto/bcrypt"
)

const resetPasswordTokenLength = 15
const resetPasswordTokenExpiry = time.Minute * 10

var requestPasswordResetPageTemplate = template.Must(template.ParseFiles("html/request-password-reset.html"))
var resetPasswordPageTemplate = template.Must(template.ParseFiles("html/reset-password.html"))

func makeRequestPasswordResetHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {

	var executeErrorTemplate = func(w http.ResponseWriter, r *http.Request, errcode int, err error) {
		if err != nil {
			logError(r, 0, err)
		}
		w.WriteHeader(errcode)
		requestPasswordResetPageTemplate.Execute(w, struct {
			Mode    string
			Error   int
			SiteKey string
		}{"request", errcode, recaptchaSiteKey})
	}

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {

			requestPasswordResetPageTemplate.Execute(w, struct {
				Mode    string
				Error   int
				SiteKey string
			}{"request", 0, recaptchaSiteKey})

		} else if r.Method == http.MethodPost {

			email := r.FormValue("email")

			if strings.TrimSpace(email) == "" {
				executeErrorTemplate(w, nil, http.StatusNotFound, nil)
				return
			}

			// Validate reCAPTCHA
			valid, err := verifyRecaptcha(r)
			if !valid {
				if err != nil {
					executeErrorTemplate(w, r,
						http.StatusInternalServerError,
						fmt.Errorf("validating recaptcha: %w", err))
					return
				}
				// Use Teapot to indicate reCAPTCHA error
				executeErrorTemplate(w, nil, http.StatusTeapot, nil)
				return
			}

			var userID uint
			var username string
			err = db.QueryRow(`
				SELECT id, username FROM user_account WHERE email = $1
				`, email).Scan(&userID, &username)
			if err != nil {
				executeErrorTemplate(w, nil, http.StatusNotFound, nil)
				return
			}

			// Create request
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			token := randomToken(resetPasswordTokenLength)
			_, err = db.Exec(`
				INSERT INTO password_reset_request
				(user_id, sent_to_address, token, created_at, requested_at_ip_address)
				VALUES ($1, $2, $3, current_timestamp, $4)
				`, userID, email, token, ip)
			if err != nil {
				executeErrorTemplate(w, r,
					http.StatusInternalServerError,
					fmt.Errorf("creating password_reset_request: %w", err))
				return
			}

			// Send email

			m, err := newMailjetClient()
			if err != nil {
				executeErrorTemplate(w, r,
					http.StatusInternalServerError,
					fmt.Errorf("mailjet init: %w", err))
				return
			}

			var url string
			if isAppEngine() {
				url = fmt.Sprintf("https://%s/reset-password?t=%s", os.Getenv("DOMAIN"), token)
			} else {
				port, err := getLocalPort(r)
				if err != nil {
					executeErrorTemplate(w, r,
						http.StatusInternalServerError, err)

				}
				url = fmt.Sprintf("http://localhost:%s/reset-password?t=%s", port, token)
			}

			messagesInfo := []mailjet.InfoMessagesV31{
				mailjet.InfoMessagesV31{
					From: &mailjet.RecipientV31{
						Email: "matt@crowdspec.dev",
						Name:  "Matt",
					},
					To: &mailjet.RecipientsV31{
						mailjet.RecipientV31{
							Email: email,
							Name:  username,
						},
					},
					Subject:  "CrowdSpec password reset for user " + username,
					TextPart: "Visit the following link to reset your password: " + url,
					HTMLPart: "Visit the following link to reset your password: <a href=\"" + url + "\">" + url + "</a>",
				},
			}
			messages := mailjet.MessagesV31{Info: messagesInfo}
			_, err = m.SendMailV31(&messages)

			if err != nil {
				executeErrorTemplate(w, r,
					http.StatusInternalServerError,
					fmt.Errorf("mailjet send reset password email: %w", err))
				return
			}

			requestPasswordResetPageTemplate.Execute(w, struct {
				Mode  string
				Email string
			}{"sent", email})

		}

	}
}

func makeResetPasswordHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {

	var executeTemplate = func(w http.ResponseWriter, r *http.Request, token string, errcode int, err error) {
		if err != nil {
			logError(r, 0, err)
		}
		if errcode != 0 {
			w.WriteHeader(errcode)
		}
		resetPasswordPageTemplate.Execute(w, struct {
			Mode    string
			Token   string
			Error   int
			SiteKey string
		}{"reset", token, errcode, recaptchaSiteKey})
	}

	return func(w http.ResponseWriter, r *http.Request) {

		token := r.FormValue("t")

		if token == "" {
			executeTemplate(w, r, "", http.StatusNotFound, nil)
			return
		}

		if r.Method == http.MethodGet {

			executeTemplate(w, r, token, 0, nil)

		} else if r.Method == http.MethodPost {

			// Validate reCAPTCHA
			valid, err := verifyRecaptcha(r)
			if !valid {
				if err != nil {
					executeTemplate(w, r, token,
						http.StatusInternalServerError,
						fmt.Errorf("validating recaptcha: %w", err))
					return
				}
				// Use Teapot to indicate reCAPTCHA error
				executeTemplate(w, r, token, http.StatusTeapot, nil)
				return
			}

			newpass := r.FormValue("newpass")
			newpass2 := r.FormValue("newpass2")

			if strings.TrimSpace(newpass) == "" {
				executeTemplate(w, r, token, http.StatusBadRequest, nil)
				return
			} else if newpass != newpass2 {
				executeTemplate(w, r, token, http.StatusForbidden, nil)
				return
			}

			var userID uint
			var username string
			var createdAt time.Time
			err = db.QueryRow(`
				SELECT r.user_id, u.username, r.created_at
				FROM password_reset_request r
				INNER JOIN user_account u
				ON u.id = r.user_id
				WHERE r.token = $1
				AND r.fulfilled_at_ip_address IS NULL
				`, token).Scan(&userID, &username, &createdAt)
			if err != nil {
				executeTemplate(w, r, "",
					http.StatusNotFound,
					fmt.Errorf("fetching password reset request: %w", err))
				return
			}

			// Error if token expired
			if time.Now().After(createdAt.Add(resetPasswordTokenExpiry)) {
				executeTemplate(w, r, "", http.StatusNotFound, nil)
				return
			}

			// Update user password

			authHash, err := bcrypt.GenerateFromPassword([]byte(newpass), BcryptCost)
			if err != nil {
				executeTemplate(w, r, token,
					http.StatusInternalServerError,
					fmt.Errorf("generating new auth hash: %w", err))
				return
			}

			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			_, err = db.Exec(`
				UPDATE password_reset_request SET fulfilled_at_ip_address=$2 WHERE token=$1
				`, token, ip)
			if err != nil {
				executeTemplate(w, r, token,
					http.StatusInternalServerError,
					fmt.Errorf("updating password reset request: %w", err))
				return
			}

			_, err = db.Exec(`
				UPDATE user_account SET auth_hash=$2 WHERE id=$1
				`, userID, authHash)
			if err != nil {
				executeTemplate(w, r, token,
					http.StatusInternalServerError,
					fmt.Errorf("updating user auth_hash: %w", err))
				return
			}

			resetPasswordPageTemplate.Execute(w, struct {
				Mode     string
				Username string
			}{"success", username})

		}

	}
}
