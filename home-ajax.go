package main

import (
	"database/sql"
	"net/http"
)

func ajaxUserHome(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET

	rows, err := db.Query(`
		SELECT id, owner_type, owner_id, created_at, updated_at, spec_name, spec_desc, is_public
		FROM spec
		WHERE owner_type=$1 AND owner_id=$2
		ORDER BY created_at ASC
		`, OwnerTypeUser, userID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	userSpecs := []Spec{}
	for rows.Next() {
		s := Spec{}
		err = rows.Scan(&s.ID, &s.OwnerType, &s.OwnerID, &s.Created, &s.Updated, &s.Name, &s.Desc, &s.Public)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		userSpecs = append(userSpecs, s)
	}

	rows, err = db.Query(`
		SELECT spec.id, owner_type, owner_id, spec.created_at, spec.updated_at, spec_name, spec_desc,
		user_account.username
		FROM spec
		LEFT JOIN user_account
		ON spec.owner_type=$1 AND user_account.id=owner_id
		WHERE is_public
		ORDER BY created_at ASC
		`, OwnerTypeUser)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	publicSpecs := []Spec{}
	for rows.Next() {
		s := Spec{}
		err = rows.Scan(&s.ID, &s.OwnerType, &s.OwnerID, &s.Created, &s.Updated, &s.Name, &s.Desc,
			&s.Username)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		publicSpecs = append(publicSpecs, s)
	}

	payload := struct {
		UserSpecs   []Spec `json:"userSpecs"`
		PublicSpecs []Spec `json:"publicSpecs"`
	}{
		UserSpecs:   userSpecs,
		PublicSpecs: publicSpecs,
	}

	return payload, http.StatusOK, nil
}
