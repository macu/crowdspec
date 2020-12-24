package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	UserProfile  UserProfileSettings  `json:"userProfile"`
	BlockEditing BlockEditingSettings `json:"blockEditing"`
}

// UserProfileSettings holds user settings regarding the user's own profile.
type UserProfileSettings struct {
	HighlightUsername *string `json:"highlightUsername"`
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
		return 0, errors.New("username required")
	}
	if len(username) > 25 {
		return 0, errors.New("username must be 25 characters or less")
	}
	if email == "" {
		return 0, errors.New("email required")
	}
	if len(email) > 50 {
		return 0, errors.New("email must be 50 characters or less")
	}
	if !emailRegexp.MatchString(email) {
		return 0, errors.New("invalid email address")
	}
	if strings.TrimSpace(password) == "" {
		return 0, errors.New("password empty")
	}

	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS(SELECT * FROM user_account WHERE username = $1)`,
		username).Scan(&exists)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, errors.New("username already exists")
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
		`INSERT INTO user_account (username, email, auth_hash, created_at)
		VALUES ($1, $2, $3, $4) RETURNING id`,
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

	err := c.QueryRow(
		`SELECT user_settings
		FROM user_account
		WHERE id = $1`, userID).Scan(&settingsString)
	if err != nil {
		return nil, fmt.Errorf("reading user settings: %w", err)
	}

	var settings UserSettings

	if settingsString != nil {
		err = json.Unmarshal([]byte(*settingsString), &settings)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling user settings: %w", err)
		}
	}

	sanitizeSettings(&settings)

	return &settings, nil
}

func saveUserSettings(c DBConn, userID uint, settings *UserSettings) error {

	if settings == nil {

		_, err := c.Exec(
			`UPDATE user_account
			SET user_settings = NULL
			WHERE id = $1`, userID)
		if err != nil {
			return fmt.Errorf("updating user_account: %w", err)
		}

	} else {

		sanitizeSettings(settings)

		settingsBytes, err := json.Marshal(settings)
		if err != nil {
			return fmt.Errorf("marshalling settings: %w", err)
		}

		_, err = c.Exec(
			`UPDATE user_account
			SET user_settings = $2
			WHERE id = $1`, userID, string(settingsBytes))
		if err != nil {
			return fmt.Errorf("updating user_account: %w", err)
		}

	}

	return nil
}

func sanitizeSettings(settings *UserSettings) {

	if settings == nil {
		return
	}

	// Sanitize values
	if settings.UserProfile.HighlightUsername != nil &&
		!isValidColour(*settings.UserProfile.HighlightUsername) {
		settings.UserProfile.HighlightUsername = nil
	}

	// Apply defaults
	if settings.BlockEditing.DeleteButton == "" {
		settings.BlockEditing.DeleteButton = "all"
	}

}
