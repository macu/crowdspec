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

// First user is allowed admin access.
const adminUserID = 1

const sessionTokenCookieName = "session_token"

const sessionTokenCookieExpiry = time.Hour * 24 * 30
const sessionTokenCookieRenewIfExpiresIn = time.Hour * 24 * 29

var loginPageTemplate = template.Must(template.ParseFiles("html/login.html"))

// AuthenticatedRoute is a request handler that also accepts *sql.DB and the authenticated userID.
type AuthenticatedRoute func(*sql.DB, uint, http.ResponseWriter, *http.Request)

func isAjax(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("X-Requested-With")) == "xmlhttprequest"
}

// Returns a function that wraps a handler in an authentication intercept that loads
// the authenticated user ID and occasionally updates the expiry of the session cookie.
// The wrapped handler is not called and 401 is returned if no user is authenticated.
func makeAuthenticator(db *sql.DB) func(handler AuthenticatedRoute) func(http.ResponseWriter, *http.Request) {
	selectUserStmt, err := db.Prepare("SELECT user_id, expires FROM user_session WHERE token=$1 AND expires>$2")
	if err != nil {
		panic(err)
	}

	// Return factory function for wrapping handlers that require authentication
	return func(handler AuthenticatedRoute) func(http.ResponseWriter, *http.Request) {

		// Return standard http.Handler which calls the authenticated handler passing db and userID
		return func(w http.ResponseWriter, r *http.Request) {

			// Read auth cookie
			sessionTokenCookie, err := r.Cookie(sessionTokenCookieName)
			if err == http.ErrNoCookie { // r.Cookie returns only ErrNoCookie or nil for error
				if isAjax(r) {
					w.WriteHeader(http.StatusForbidden)
				} else {
					// Redirect to login if no auth cookie
					http.Redirect(w, r, "/login", http.StatusSeeOther)
				}
				return
			}

			// Look up session and read authenticated userID
			now := time.Now()
			var userID uint
			var expires time.Time
			err = selectUserStmt.QueryRow(sessionTokenCookie.Value, now).Scan(&userID, &expires)
			if err == sql.ErrNoRows {
				if isAjax(r) {
					w.WriteHeader(http.StatusForbidden)
				} else {
					// Redirect to login if no valid session
					http.Redirect(w, r, "/login", http.StatusSeeOther)
				}
				return
			} else if err != nil {
				logError(r, 0, fmt.Errorf("loading user from session token: %w", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Refresh session and cookie if old
			if expires.Before(now.Add(sessionTokenCookieRenewIfExpiresIn)) {

				// Update session expires time
				expires := now.Add(sessionTokenCookieExpiry)
				_, err = db.Exec(
					`UPDATE user_session SET expires=$1 WHERE token=$2`,
					expires, sessionTokenCookie.Value)
				if err != nil {
					logError(r, userID, fmt.Errorf("updating session expiry: %w", err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Update cookie expires time
				http.SetCookie(w, &http.Cookie{
					Name:     sessionTokenCookieName,
					Value:    sessionTokenCookie.Value,
					Path:     "/",
					Expires:  expires,
					HttpOnly: true,                    // don't expose cookie to JavaScript
					SameSite: http.SameSiteStrictMode, // send in first-party contexts only
				})
			}

			// Invoke route with authenticated user info
			handler(db, userID, w, r)
		}
	}
}

func makeAdminAuthenticatedRoute(h AuthenticatedRoute) AuthenticatedRoute {
	return func(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) {
		if userID == adminUserID {
			h(db, userID, w, r)
		} else {
			logError(r, userID, fmt.Errorf("forbidden admin access"))
			w.WriteHeader(http.StatusForbidden)
		}
	}
}

func makeLoginHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {

	var executeLoginTemplate = func(w http.ResponseWriter, r *http.Request,
		statusCode int, errMsg string, err error,
	) {
		if statusCode != 0 {
			w.WriteHeader(statusCode)
		}
		if err != nil {
			logError(r, 0, err)
		}
		loginPageTemplate.Execute(w, struct {
			Error        string
			SiteKey      string
			Verify       bool // require reCAPTCHA
			VersionStamp string
		}{
			errMsg,
			recaptchaSiteKey,
			isAppEngine(),
			cacheControlVersionStamp,
		})
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {

			executeLoginTemplate(w, r, http.StatusOK, "", nil)

		} else if r.Method == http.MethodPost {

			username := r.FormValue("username")
			password := r.FormValue("password")

			// Validate reCAPTCHA
			valid, err := verifyRecaptcha(r)
			if !valid {
				if err != nil {
					executeLoginTemplate(w, r,
						http.StatusInternalServerError, "Server error",
						fmt.Errorf("validating recaptcha: %w", err))
					return
				}
				// Use Teapot to indicate reCAPTCHA error
				executeLoginTemplate(w, r,
					http.StatusTeapot, "Invalid reCAPTCHA",
					fmt.Errorf("invalid reCAPTCHA [IP %s]", getUserIP(r)))
				return
			}

			var userID uint
			var authHash string
			err = db.QueryRow(
				`SELECT id, auth_hash FROM user_account WHERE username=$1 OR email=$1`,
				username,
			).Scan(&userID, &authHash)
			if err != nil {
				if err == sql.ErrNoRows {
					// TODO Limit failed attempts
					logNotice(r, struct {
						Event     string
						Username  string
						IPAddress string
					}{
						"InvalidLogin",
						username,
						getUserIP(r),
					})
					executeLoginTemplate(w, r,
						http.StatusForbidden, "Invalid login",
						nil) // don't log
					return
				}
				executeLoginTemplate(w, r,
					http.StatusInternalServerError, "Server error",
					fmt.Errorf("looking up user: %w", err))
				return
			}

			err = bcrypt.CompareHashAndPassword([]byte(authHash), []byte(password))
			if err != nil {
				// TODO Limit failed attempts
				logNotice(r, struct {
					Event     string
					Username  string
					IPAddress string
				}{
					"InvalidLogin",
					username,
					getUserIP(r),
				})
				executeLoginTemplate(w, r,
					http.StatusForbidden, "Invalid login",
					nil) // don't log
				return
			}

			err = authUser(w, r, db, userID)
			if err != nil {
				executeLoginTemplate(w, r,
					http.StatusInternalServerError, "Server error",
					fmt.Errorf("inserting session: %w", err))
				return
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)

			logDefault(r, struct {
				Event     string
				UserID    uint
				IPAddress string
			}{
				"UserLogin",
				userID,
				getUserIP(r),
			})

		} else {

			w.WriteHeader(http.StatusBadRequest)
			executeLoginTemplate(w, r,
				http.StatusBadRequest, "Unrecognized method", nil)

		}
	}
}

const sessionRandLetters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Returns a random session ID that includes current Unix time in nanoseconds.
func makeSessionID() string {
	// 20 digits (current time) + 1 (:) + 9 (random) = 30 digit session ID
	// 20 digits gives until around 5138 (over 3117 years from now as of writing)
	// assuming Earth's orbit and day remains stable
	// https://www.epochconverter.com/
	return fmt.Sprintf("%020d:%s", time.Now().UnixNano(), randomToken(9))
}

func authUser(w http.ResponseWriter, r *http.Request, db DBConn, userID uint) error {

	token := makeSessionID()
	expires := time.Now().Add(sessionTokenCookieExpiry)
	_, err := db.Exec(
		`INSERT INTO user_session (token, user_id, expires) VALUES ($1, $2, $3)`,
		token, userID, expires)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionTokenCookieName,
		Value:    token,
		Path:     "/", // enable AJAX (Info: https://stackoverflow.com/a/22432999/1597274)
		Expires:  expires,
		HttpOnly: true,                    // don't expose cookie to JavaScript
		SameSite: http.SameSiteStrictMode, // send in first-party contexts only
	})

	return nil

}

func logoutHandler(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) {

	sessionTokenCookie, _ := r.Cookie(sessionTokenCookieName)

	_, err := db.Exec("DELETE FROM user_session WHERE token=$1", sessionTokenCookie.Value)
	if err != nil {
		logError(r, userID, fmt.Errorf("deleting session: %w", err))
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionTokenCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,                    // don't expose cookie to JavaScript
		SameSite: http.SameSiteStrictMode, // send in first-party contexts only
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func deleteExpiredSessions(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM user_session WHERE expires <= $1", time.Now())
	if err != nil {
		return fmt.Errorf("deleting expired sessions: %w", err)
	}
	return nil
}
