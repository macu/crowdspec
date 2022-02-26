package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func ajaxLoadAuthenticatedUser(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {

	var username, email string
	err := db.QueryRow(
		`SELECT username, email FROM user_account WHERE id=$1`, userID,
	).Scan(&username, &email)
	if err != nil {
		logError(r, &userID, fmt.Errorf("selecting username: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		return nil, http.StatusInternalServerError
	}
	settings, err := loadUserSettings(db, userID)
	if err != nil {
		logError(r, &userID, fmt.Errorf("loading user settings: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		return nil, http.StatusInternalServerError
	}

	return struct {
		ID           uint          `json:"id"`
		Username     string        `json:"username"`
		EmailAddress string        `json:"email"`
		Admin        bool          `json:"admin"`
		Settings     *UserSettings `json:"settings"`
	}{
		userID,
		username,
		email,
		userID == adminUserID,
		settings,
	}, http.StatusOK

}

func ajaxUserChangePassword(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {

	err := r.ParseForm()
	if err != nil {
		logError(r, &userID, err)
		return nil, http.StatusInternalServerError
	}

	oldPassword := r.Form.Get("old")
	newPassword := r.Form.Get("new")
	newPasswordConfirm := r.Form.Get("new2")

	if len(strings.TrimSpace(newPassword)) < passwordMinLength || newPassword != newPasswordConfirm {
		return nil, http.StatusBadRequest
	}

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		var authHash string
		err := tx.QueryRow(
			`SELECT auth_hash FROM user_account WHERE id=$1`, userID,
		).Scan(&authHash)
		if err != nil {
			logError(r, &userID, fmt.Errorf("looking up user: %w", err))
			return nil, http.StatusInternalServerError
		}

		err = bcrypt.CompareHashAndPassword([]byte(authHash), []byte(oldPassword))
		if err != nil {
			// Silently ignore incorrect password errors
			return nil, http.StatusForbidden
		}

		newAuthHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
		if err != nil {
			logError(r, &userID, fmt.Errorf("hashing password: %w", err))
			return nil, http.StatusInternalServerError
		}

		_, err = tx.Exec(
			`UPDATE user_account SET auth_hash=$3 WHERE id=$1 AND auth_hash=$2`,
			userID, authHash, newAuthHash)
		if err != nil {
			logError(r, &userID, fmt.Errorf("updating user: %w", err))
			return nil, http.StatusInternalServerError
		}

		logNotice(r, struct {
			Event  string
			UserID uint
			// IPAddress string
		}{
			"UpdatePassword",
			userID,
			// getUserIP(r),
		})

		return nil, http.StatusOK
	})
}

func ajaxUserSettings(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {

	settings, err := loadUserSettings(db, userID)
	if err != nil {
		logError(r, &userID, fmt.Errorf("loading user settings: %w", err))
		return nil, http.StatusInternalServerError
	}

	return settings, http.StatusOK
}

func ajaxUserSaveSettings(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {

	err := r.ParseForm()
	if err != nil {
		logError(r, &userID, err)
		return nil, http.StatusInternalServerError
	}

	var settings = UserSettings{}
	err = json.Unmarshal([]byte(r.Form.Get("settings")), &settings)
	if err != nil {
		logError(r, &userID, fmt.Errorf("parsing user settings: %w", err))
		return nil, http.StatusBadRequest
	}

	err = saveUserSettings(db, userID, &settings)
	if err != nil {
		logError(r, &userID, fmt.Errorf("saving user settings: %w", err))
		return nil, http.StatusInternalServerError
	}

	return settings, http.StatusOK
}
