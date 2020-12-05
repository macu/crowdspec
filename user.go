package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserAccount represents a user_account record.
type UserAccount struct {
	ID       int64     `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Created  time.Time `json:"created"`
}

// UserSettings represents a user's configurable settings.
type UserSettings struct {
	BlockEditing BlockEditingSettings `json:"blockEditing"`
}

// BlockEditingSettings holds user settings regarding blocks they can edit.
type BlockEditingSettings struct {
	DeleteButton string `json:"deleteButton"`
}

// Regex adapted from https://www.w3.org/TR/html5/forms.html#valid-e-mail-address
var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Returns the user ID if the user was successfully created.
func createUser(db *sql.DB, username, password, email string) (int64, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	if username == "" {
		return 0, errors.New("Username required")
	}
	if len(username) > 25 {
		return 0, errors.New("Username must be 25 characters or less")
	}
	if email == "" {
		return 0, errors.New("Email required")
	}
	if len(email) > 50 {
		return 0, errors.New("Email must be 50 characters or less")
	}
	if !emailRegexp.MatchString(email) {
		return 0, errors.New("Invalid email address")
	}
	if password == "" {
		return 0, errors.New("Password must not be empty")
	}

	existing := db.QueryRow("SELECT EXISTS(SELECT * FROM user_account WHERE username = $1)", username)
	var exists bool
	err := existing.Scan(&exists)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, errors.New("Username already exists")
	}

	authHash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return 0, err
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}

	var userID int64
	err = tx.QueryRow(
		"INSERT INTO user_account (username, email, auth_hash, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		username, email, authHash, time.Now()).Scan(&userID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	// TODO Create any initial user-related records in same transaction

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func loadUserSettings(c DBConn, userID uint) (*UserSettings, error) {

	var settingsString *string

	err := c.QueryRow(`SELECT user_settings
		FROM user_account
		WHERE id = $1`, userID).Scan(&settingsString)
	if err != nil {
		return nil, err
	}

	var settings UserSettings

	if settingsString != nil {
		err = json.Unmarshal([]byte(*settingsString), &settings)
		if err != nil {
			return nil, err
		}
	}

	// Apply defaults
	if settings.BlockEditing.DeleteButton == "" {
		settings.BlockEditing.DeleteButton = "all"
	}

	return &settings, nil

}

func saveUserSettings(c DBConn, userID uint, settings *UserSettings) error {

	if settings == nil {

		_, err := c.Exec(`UPDATE user_account
			SET user_settings = NULL
			WHERE id = $1`, userID)
		if err != nil {
			return err
		}

	} else {

		settingsBytes, err := json.Marshal(settings)
		if err != nil {
			return err
		}

		_, err = c.Exec(`UPDATE user_account
			SET user_settings = $2
			WHERE id = $1`, userID, string(settingsBytes))
		if err != nil {
			return err
		}

	}

	return nil

}
