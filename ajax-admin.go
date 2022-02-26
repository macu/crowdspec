package main

import (
	"database/sql"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"
)

type signupRequest struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	EmailAddress string    `json:"email"`
	CreatedAt    time.Time `json:"created"`
	Reviewed     bool      `json:"reviewed"`
	Approved     bool      `json:"approved"`
	UserID       *uint     `json:"userId"`
}

type adminUserView struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Highlight    string    `json:"highlight"`
	EmailAddress string    `json:"email"`
	CreatedAt    time.Time `json:"created"`
	SpecCount    int       `json:"specs"`
}

func ajaxAdminLoadSignupRequests(db *sql.DB,
	userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {

	var requests = []*signupRequest{}

	var conds string
	var orderby string

	if AtoBool(r.FormValue("all")) {
		// no conds
		orderby = `ORDER BY created_at DESC`
	} else {
		conds = `WHERE NOT reviewed`
		orderby = `ORDER BY created_at ASC`
	}

	rows, err := db.Query(
		`SELECT id, username, email, created_at, reviewed, approved, user_id
		FROM user_signup_request
		` + conds + `
		` + orderby)
	if err != nil {
		logError(r, &userID, fmt.Errorf("loading signup requests: %w", err))
		return nil, http.StatusInternalServerError
	}

	for rows.Next() {
		var sr signupRequest
		err = rows.Scan(&sr.ID, &sr.Username, &sr.EmailAddress, &sr.CreatedAt,
			&sr.Reviewed, &sr.Approved, &sr.UserID)
		if err != nil {
			if err2 := rows.Close(); err2 != nil {
				logError(r, &userID, fmt.Errorf("closing rows: %s; on scan error: %w", err2, err))
				return nil, http.StatusInternalServerError
			}
			logError(r, &userID, fmt.Errorf("scanning signup requests: %w", err))
			return nil, http.StatusInternalServerError
		}
		requests = append(requests, &sr)
	}

	return requests, http.StatusOK
}

func ajaxAdminSubmitSignupRequestReview(db *sql.DB,
	userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {

	requestID, err := AtoInt64(r.FormValue("requestId"))
	if err != nil {
		logError(r, &userID, fmt.Errorf("parsing requestId: %w", err))
		return nil, http.StatusBadRequest
	}

	approved := AtoBool(r.FormValue("approved"))

	message := r.FormValue("message")

	var token *string
	if approved {
		var t = randomToken(signupRequestTokenLength)
		token = &t
	}

	var username, email string

	err = db.QueryRow(
		`UPDATE user_signup_request
		SET reviewed=true, approved=$2, token=$3
		WHERE id=$1
		RETURNING username, email`,
		requestID, approved, token,
	).Scan(&username, &email)

	if err != nil {
		logError(r, &userID, fmt.Errorf("updating signup request: %w", err))
		return nil, http.StatusInternalServerError
	}

	var messagePlain, messageHTML string
	if strings.TrimSpace(message) != "" {
		messagePlain = "\n\nMessage:\n" + message
		messageHTML = "\n<br/>\n" +
			strings.ReplaceAll(html.EscapeString(message), "\n", "<br/>")
	}

	if approved {

		url, err := buildAbsoluteURL(r, "activate-signup?t="+*token)
		if err != nil {
			logError(r, &userID, fmt.Errorf("building URL: %w", err))
			return nil, http.StatusInternalServerError
		}

		err = sendEmail(username, email,
			"CrowdSpec signup request approved for "+username,
			"Visit this link to set a password and log in: "+url+
				messagePlain,
			"<p>Visit this link to set a password and log in: "+
				`<a href="`+url+`">`+url+`</a></p>`+
				messageHTML,
		)
		if err != nil {
			logError(r, &userID, fmt.Errorf("sending signup request approval email: %w", err))
			return nil, http.StatusInternalServerError
		}

	} else {

		err = sendEmail(username, email,
			"CrowdSpec signup request declined for "+username,
			"Sorry, I have decided not to approve your signup request."+
				messagePlain,
			"<p>Sorry, I have decided not to approve your signup request.</p>"+
				messageHTML,
		)
		if err != nil {
			logError(r, &userID, fmt.Errorf("sending signup request approval email: %w", err))
			return nil, http.StatusInternalServerError
		}

	}

	return nil, http.StatusOK
}

func ajaxAdminLoadUsers(db *sql.DB,
	userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {

	var users = []*adminUserView{}

	rows, err := db.Query(
		`SELECT id, username, email, created_at,
			(SELECT COUNT(*) FROM spec WHERE owner_type=$1 AND owner_id=user_account.id) AS spec_count
		FROM user_account
		ORDER BY id`,
		OwnerTypeUser)
	if err != nil {
		logError(r, &userID, fmt.Errorf("loading signup requests: %w", err))
		return nil, http.StatusInternalServerError
	}

	for rows.Next() {
		var u adminUserView
		err = rows.Scan(&u.ID, &u.Username, &u.EmailAddress, &u.CreatedAt,
			&u.SpecCount)
		if err != nil {
			if err2 := rows.Close(); err2 != nil {
				logError(r, &userID, fmt.Errorf("closing rows: %s; on scan error: %w", err2, err))
				return nil, http.StatusInternalServerError
			}
			logError(r, &userID, fmt.Errorf("scanning user: %w", err))
			return nil, http.StatusInternalServerError
		}
		users = append(users, &u)
	}

	return users, http.StatusOK
}
