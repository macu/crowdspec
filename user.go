package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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
	UserProfile  UserProfileSettings   `json:"userProfile"`
	Homepage     UserHomepageSettings  `json:"homepage"`
	BlockEditing BlockEditingSettings  `json:"blockEditing"`
	Community    UserCommunitySettings `json:"community"`
}

// UserProfileSettings holds user settings regarding the user's own profile.
type UserProfileSettings struct {
	HighlightUsername *string `json:"highlightUsername"`
}

// UserHomepageSettings holds user settings regarding the homepage.
type UserHomepageSettings struct {
	SpecsLayout string `json:"specsLayout"`
}

// BlockEditingSettings holds user settings regarding blocks they can edit.
type BlockEditingSettings struct {
	DeleteButton string `json:"deleteButton"`
}

// UserCommunitySettings holds user-specific community-related settings.
type UserCommunitySettings struct {
	UnreadOnly bool `json:"unreadOnly"`
}

// Username pattern
var usernameRegex = regexp.MustCompile("^[a-zA-Z0-9_]+$")

// Regex adapted from https://www.w3.org/TR/html5/forms.html#valid-e-mail-address
var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

const usernameMaxLength = 25
const emailAddressMaxLength = 50
const passwordMinLength = 5

// Used for error logging userID when there is no user logged in.
const nullUserID = 0

// Returns the user ID if the user was successfully created.
func createUser(r *http.Request, db *sql.DB,
	username, password, email string) (uint, error) {

	var newUserID uint
	var err error

	err = inTransaction(r, db, func(tx *sql.Tx) error {
		// inTransaction may return same or different error
		newUserID, err = createUserTx(tx, username, password, email)
		return err
	})

	if err != nil {
		return 0, err
	}

	return newUserID, nil
}

// Creates a user within an existing transaction.
// Error is returned on failure; cancel the transaction in the parent context.
// Returns the user ID if the user was successfully created.
func createUserTx(tx *sql.Tx, username, password, email string) (uint, error) {

	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	if username == "" {
		return 0, fmt.Errorf("username required")
	}
	if len(username) > usernameMaxLength {
		return 0, fmt.Errorf("username must be %d characters or less", usernameMaxLength)
	}
	if email == "" {
		return 0, fmt.Errorf("email required")
	}
	if len(email) > emailAddressMaxLength {
		return 0, fmt.Errorf("email must be %d characters or less", emailAddressMaxLength)
	}
	if !emailRegexp.MatchString(email) {
		return 0, fmt.Errorf("invalid email address")
	}
	if len(strings.TrimSpace(password)) < passwordMinLength {
		return 0, fmt.Errorf("password must be %d characters or more", passwordMinLength)
	}

	var exists bool
	err := tx.QueryRow(
		`SELECT EXISTS(SELECT * FROM user_account WHERE username = $1)`, username,
	).Scan(&exists)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, fmt.Errorf("username already exists")
	}

	authHash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return 0, err
	}

	var userID uint
	err = tx.QueryRow(
		`INSERT INTO user_account (username, email, auth_hash, created_at)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		username, email, authHash, time.Now()).Scan(&userID)
	if err != nil {
		return 0, err
	}

	// TODO Create any initial user-related records in same transaction

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
	if settings.UserProfile.HighlightUsername != nil {
		if !isValidColour(*settings.UserProfile.HighlightUsername) {
			settings.UserProfile.HighlightUsername = nil
		}
	}
	if settings.Homepage.SpecsLayout == "" ||
		!(settings.Homepage.SpecsLayout == "list" ||
			settings.Homepage.SpecsLayout == "grid") {
		settings.Homepage.SpecsLayout = "list"
	}
	if settings.BlockEditing.DeleteButton == "" ||
		!(settings.BlockEditing.DeleteButton == "modal" ||
			settings.BlockEditing.DeleteButton == "recent" ||
			settings.BlockEditing.DeleteButton == "all") {
		settings.BlockEditing.DeleteButton = "all"
	}

}
