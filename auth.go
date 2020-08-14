package main

import (
	"bytes"
	crypto_rand "crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sessionTokenCookieName = "session_token"

const sessionTokenCookieExpiry = time.Hour * 24 * 30
const sessionTokenCookieRenewIfExpiresIn = time.Hour * 24 * 29

var loginPageTemplate = template.Must(template.ParseFiles("login.html"))

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
	updateSessionStmt, err := db.Prepare("UPDATE user_session SET expires=$1 WHERE token=$2")
	if err != nil {
		panic(err)
	}

	// Return factory function for wrapping handlers that require authentication
	return func(handler AuthenticatedRoute) func(http.ResponseWriter, *http.Request) {

		// Return standard http.Handler which calls the authenticated handler passing db and userID
		return func(w http.ResponseWriter, r *http.Request) {

			// Read auth cookie
			sessionTokenCookie, err := r.Cookie(sessionTokenCookieName)
			if err == http.ErrNoCookie {
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
				_, err = updateSessionStmt.Exec(expires, sessionTokenCookie.Value)
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
					HttpOnly: true,
				})
			}

			// Invoke route with authenticated user info
			handler(db, userID, w, r)
		}
	}
}

func makeLoginHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	selectUserStmt, err := db.Prepare("SELECT id, auth_hash FROM user_account WHERE username=$1")
	if err != nil {
		panic(err)
	}
	insertSessionStmt, err := db.Prepare("INSERT INTO user_session (token, user_id, expires) VALUES ($1, $2, $3)")
	if err != nil {
		panic(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			loginPageTemplate.Execute(w, struct{ Error int }{})
		} else if r.Method == http.MethodPost {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			username := r.FormValue("username")
			password := r.FormValue("password")
			var userID uint
			var authHash string
			err := selectUserStmt.QueryRow(username).Scan(&userID, &authHash)
			if err == sql.ErrNoRows {
				// TODO Limit failed attempts
				log.Printf("invalid username: %s [%s]", username, ip) // no error report
				w.WriteHeader(http.StatusForbidden)
				loginPageTemplate.Execute(w, struct{ Error int }{http.StatusForbidden})
				return
			} else if err != nil {
				logError(r, 0, fmt.Errorf("looking up user: %w", err))
				w.WriteHeader(http.StatusInternalServerError)
				loginPageTemplate.Execute(w, struct{ Error int }{http.StatusInternalServerError})
				return
			}
			err = bcrypt.CompareHashAndPassword([]byte(authHash), []byte(password))
			if err != nil {
				// TODO Limit failed attempts
				log.Printf("invalid password for user: %s [%s]", username, ip) // no error report
				w.WriteHeader(http.StatusUnauthorized)
				loginPageTemplate.Execute(w, struct{ Error int }{http.StatusForbidden})
				return
			}
			token := makeSessionID()
			expires := time.Now().Add(sessionTokenCookieExpiry)
			_, err = insertSessionStmt.Exec(token, userID, expires)
			if err != nil {
				logError(r, userID, fmt.Errorf("inserting session: %w", err))
				w.WriteHeader(http.StatusInternalServerError)
				loginPageTemplate.Execute(w, struct{ Error int }{http.StatusInternalServerError})
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     sessionTokenCookieName,
				Value:    token,
				Path:     "/", // Info: https://stackoverflow.com/a/22432999/1597274
				Expires:  expires,
				HttpOnly: true,
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			log.Printf("user login: %s [%s]", username, ip)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			loginPageTemplate.Execute(w, struct{ Error int }{http.StatusBadRequest})
		}
	}
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
		HttpOnly: true,
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

const sessionRandLetters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Returns a random session ID that includes current Unix time in nanoseconds.
func makeSessionID() string {
	var seed int64
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		logError(nil, 0, fmt.Errorf("reading session ID random generator seed bytes with crypto/rand: %w", err))
		seed = time.Now().UnixNano()
	} else {
		buf := bytes.NewBuffer(b[:])
		err = binary.Read(buf, binary.LittleEndian, &seed)
		if err != nil {
			logError(nil, 0, fmt.Errorf("reading session ID random generator seed with binary.Read: %w", err))
			seed = time.Now().UnixNano()
		}
	}

	randGen := rand.New(rand.NewSource(seed))

	bytes := make([]byte, 9)

	for i := range bytes {
		bytes[i] = sessionRandLetters[randGen.Intn(len(sessionRandLetters))]
	}

	// 20 digits (current time) + 1 (:) + 9 (random) = 30 digit session ID
	return fmt.Sprintf("%020d:%s", time.Now().UnixNano(), string(bytes))
}
