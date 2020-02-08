package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func inTransaction(c context.Context, db *sql.DB, f func(*sql.Tx) (interface{}, int, error)) (interface{}, int, error) {
	tx, err := db.BeginTx(c, nil)
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			log.Println(rbErr)
		}
		return nil, http.StatusInternalServerError, err
	}

	response, statusCode, err := f(tx)
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			log.Println(rbErr)
		}
		return nil, statusCode, err
	}

	err = tx.Commit()
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			log.Println(rbErr)
		}
		return nil, http.StatusInternalServerError, err
	}

	return response, statusCode, nil
}

// AtoUint converts base 10 string to uint.
func AtoUint(s string) (uint, error) {
	r, err := strconv.ParseUint(s, 10, 64)
	return uint(r), err
}

// Spec represents a db spec row
type Spec struct {
	ID      uint      `json:"id"`
	Created time.Time `json:"created"`

	OwnerType string `json:"ownerType"`
	OwnerID   uint   `json:"ownerId"`
	OwnerName string `json:"ownerName"`

	Name string `json:"name"`
	Desc string `json:"desc"`

	Public bool `json:"public"`
}

func ajaxCreateSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	var ownerType string
	var ownerID uint

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	orgID := r.Form.Get("orgId")
	if strings.TrimSpace(orgID) != "" {
		// TODO Verify org admin
		ownerType = "org"
		ownerID, err = AtoUint(orgID)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
	} else {
		ownerType = "user"
		ownerID = userID
	}

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("Blank spec name")
	}

	desc := strings.TrimSpace(r.Form.Get("desc"))

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {
		res, err := tx.Exec(`
				INSERT INTO spec (owner_type, owner_id, created, name, description) VALUES (?, ?, ?, ?, ?)
				`, ownerType, ownerID, time.Now(), name, desc)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		specID, err := res.LastInsertId()
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return specID, http.StatusOK, nil
	})
}

// SpecUserAccess represents a spec accessed by a user
type SpecUserAccess struct {
	Spec

	UserIsAdmin       bool `json:"userIsAdmin"`
	UserIsContributor bool `json:"userIsContributor"`
}

func ajaxSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	query := r.URL.Query()

	specID := query.Get("specId")
	if specID == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("No specId provided")
	}

	// TODO Verify read access

	// TODO Finish owner_name, user_is_admin, user_is_contributor
	s := &SpecUserAccess{}
	row := db.QueryRow(`
		SELECT spec.id, spec.created, spec.owner_type, spec.owner_id, spec.name, spec.description, spec.public,
			IF(spec.owner_type="user", user.username, "TODO") AS owner_name,
			IF(spec.owner_type="user" AND spec.owner_id=?, 1, 0) AS user_is_admin,
			IF(spec.owner_type="user" AND spec.owner_id=?, 1, 0) AS user_is_contributor
		FROM spec
		LEFT JOIN user
			ON spec.owner_type="user"
			AND user.id=spec.owner_id
		WHERE spec.id=?
		`, userID, userID, specID)
	err := row.Scan(&s.ID, &s.Created, &s.OwnerType, &s.OwnerID, &s.Name, &s.Desc, &s.Public,
		&s.OwnerName, &s.UserIsAdmin, &s.UserIsContributor)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return s, http.StatusOK, nil
}

func ajaxUserSpecs(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	rows, err := db.Query(`
		SELECT spec.id, spec.created, spec.owner_type, spec.owner_id, spec.name, spec.description, spec.public,
			user.username AS owner_name
		FROM spec
		INNER JOIN user
			ON user.id=spec.owner_id
		WHERE spec.owner_type="user" AND spec.owner_id=?
		`, userID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specs := []Spec{}
	for rows.Next() {
		s := Spec{}
		err = rows.Scan(&s.ID, &s.Created, &s.OwnerType, &s.OwnerID, &s.Name, &s.Desc, &s.Public,
			&s.OwnerName)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		specs = append(specs, s)
	}

	return specs, http.StatusOK, nil
}

func ajaxPublicSpecs(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// TODO Finish owner_name, user_is_admin, user_is_contributor
	rows, err := db.Query(`
		SELECT spec.id, spec.created, spec.owner_type, spec.owner_id, spec.name, spec.description, spec.public,
			IF(spec.owner_type="user", user.username, "TODO") AS owner_name,
			IF(user.id=?, 1, 0) AS user_is_admin,
			IF(user.id=?, 1, 0) AS user_is_contributor
		FROM spec
		INNER JOIN user
			ON user.id=spec.owner_id
		WHERE spec.public=1
		`, userID, userID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specs := []SpecUserAccess{}
	for rows.Next() {
		s := SpecUserAccess{}
		err = rows.Scan(&s.ID, &s.Created, &s.OwnerType, &s.OwnerID, &s.Name, &s.Desc, &s.Public,
			&s.OwnerName, &s.UserIsAdmin, &s.UserIsContributor)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		specs = append(specs, s)
	}

	return specs, http.StatusOK, nil
}
