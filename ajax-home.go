package main

import (
	"database/sql"
	"fmt"
	"net/http"
)

func ajaxUserHome(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// GET

	rows, err := db.Query(
		`SELECT id, owner_type, owner_id, spec_name, spec_desc, is_public,
			GREATEST(spec.updated_at, spec.blocks_updated_at) AS last_updated
		FROM spec
		WHERE owner_type=$1 AND owner_id=$2
		ORDER BY created_at ASC
		`, OwnerTypeUser, userID)
	if err != nil {
		logError(r, userID, fmt.Errorf("querying user specs: %w", err))
		return nil, http.StatusInternalServerError
	}

	userSpecs := []Spec{}
	for rows.Next() {
		s := Spec{}
		err = rows.Scan(&s.ID, &s.OwnerType, &s.OwnerID, &s.Name, &s.Desc, &s.Public, &s.Updated)
		if err != nil {
			if err2 := rows.Close(); err2 != nil { // TODO Add everywhere
				logError(r, userID, fmt.Errorf("closing rows: %s; on scan error: %w", err2, err))
				return nil, http.StatusInternalServerError
			}
			logError(r, userID, fmt.Errorf("scanning spec: %w", err))
			return nil, http.StatusInternalServerError
		}
		userSpecs = append(userSpecs, s)
	}

	rows, err = db.Query(
		`SELECT spec.id, owner_type, owner_id, spec_name, spec_desc,
			user_account.username,
			user_account.user_settings::json#>>'{userProfile,highlightUsername}' AS highlight,
		GREATEST(spec.updated_at, spec.blocks_updated_at) AS last_updated
		FROM spec
		LEFT JOIN user_account
		ON spec.owner_type=$1 AND user_account.id=owner_id
		WHERE is_public
		ORDER BY spec.created_at ASC
		`, OwnerTypeUser)
	if err != nil {
		logError(r, userID, fmt.Errorf("querying public specs: %w", err))
		return nil, http.StatusInternalServerError
	}

	publicSpecs := []Spec{}
	for rows.Next() {
		s := Spec{}
		err = rows.Scan(&s.ID, &s.OwnerType, &s.OwnerID, &s.Name, &s.Desc,
			&s.Username, &s.Highlight, &s.Updated)
		if err != nil {
			if err2 := rows.Close(); err2 != nil { // TODO Add everywhere
				logError(r, userID, fmt.Errorf("closing rows: %s; on scan error: %w", err2, err))
				return nil, http.StatusInternalServerError
			}
			logError(r, userID, fmt.Errorf("scanning spec: %w", err))
			return nil, http.StatusInternalServerError
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

	return payload, http.StatusOK
}
