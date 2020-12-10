package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func ajaxUserChangePassword(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {

	err := r.ParseForm()
	if err != nil {
		logError(r, userID, err)
		return nil, http.StatusInternalServerError
	}

	oldPassword := r.Form.Get("old")
	newPassword := r.Form.Get("new")
	newPasswordConfirm := r.Form.Get("new2")

	if len(strings.TrimSpace(newPassword)) < 5 || newPassword != newPasswordConfirm {
		logError(r, userID, fmt.Errorf("invalid new password"))
		return nil, http.StatusBadRequest
	}

	return inTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		var authHash string
		err := tx.QueryRow(`SELECT auth_hash FROM user_account WHERE id=$1`, userID).Scan(&authHash)
		if err != nil {
			logError(r, userID, fmt.Errorf("looking up user: %w", err))
			return nil, http.StatusInternalServerError
		}

		err = bcrypt.CompareHashAndPassword([]byte(authHash), []byte(oldPassword))
		if err != nil {
			// Silently ignore incorrect password errors
			return nil, http.StatusForbidden
		}

		newAuthHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
		if err != nil {
			logError(r, userID, fmt.Errorf("hashing password: %w", err))
			return nil, http.StatusInternalServerError
		}

		_, err = tx.Exec(`
			UPDATE user_account SET auth_hash=$3 WHERE id=$1 AND auth_hash=$2
			`, userID, authHash, newAuthHash)
		if err != nil {
			logError(r, userID, fmt.Errorf("updating user: %w", err))
			return nil, http.StatusInternalServerError
		}

		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		log.Printf("password updated for user ID %d [%s]", userID, ip)

		return nil, http.StatusOK
	})
}

func ajaxUserSettings(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {

	settings, err := loadUserSettings(db, userID)
	if err != nil {
		logError(r, userID, fmt.Errorf("loading user settings: %w", err))
		return nil, http.StatusInternalServerError
	}

	return settings, http.StatusOK
}

func ajaxUserSaveSettings(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {

	err := r.ParseForm()
	if err != nil {
		logError(r, userID, err)
		return nil, http.StatusInternalServerError
	}

	var settings = UserSettings{}
	err = json.Unmarshal([]byte(r.Form.Get("settings")), &settings)
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing user settings: %w", err))
		return nil, http.StatusBadRequest
	}

	err = saveUserSettings(db, userID, &settings)
	if err != nil {
		logError(r, userID, fmt.Errorf("saving user settings: %w", err))
		return nil, http.StatusInternalServerError
	}

	return settings, http.StatusOK
}
