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

func ajaxUserChangePassword(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	oldPassword := r.Form.Get("old")
	newPassword := r.Form.Get("new")
	newPasswordConfirm := r.Form.Get("new2")

	if len(strings.TrimSpace(newPassword)) < 5 || newPassword != newPasswordConfirm {
		return nil, http.StatusBadRequest, fmt.Errorf("new password invalid")
	}

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		var authHash string
		err := tx.QueryRow(`SELECT auth_hash FROM user_account WHERE id=$1`, userID).Scan(&authHash)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("looking up user: %w", err)
		}

		err = bcrypt.CompareHashAndPassword([]byte(authHash), []byte(oldPassword))
		if err != nil {
			return nil, http.StatusForbidden, fmt.Errorf("old password invalid: %w", err)
		}

		newAuthHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("hashing password: %w", err)
		}

		_, err = tx.Exec(`
			UPDATE user_account SET auth_hash=$3 WHERE id=$1 AND auth_hash=$2
			`, userID, authHash, newAuthHash)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("updating user: %w", err)
		}

		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		log.Printf("password updated for user ID %d [%s]", userID, ip)

		return nil, http.StatusOK, nil
	})
}

func ajaxUserSettings(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {

	settings, err := loadUserSettings(db, userID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return settings, http.StatusOK, nil

}

func ajaxUserSaveSettings(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var settings = UserSettings{}
	err = json.Unmarshal([]byte(r.Form.Get("settings")), &settings)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	err = saveUserSettings(db, userID, &settings)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return settings, http.StatusOK, nil

}
