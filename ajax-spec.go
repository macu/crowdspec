package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Returns the requested spec with immediate blocks.
func ajaxSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// GET
	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing specId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, err := verifyReadSpec(db, userID, specID); !access || err != nil {
		if err != nil {
			logError(r, userID, fmt.Errorf("validating read spec access: %w", err))
			return nil, http.StatusInternalServerError
		}
		logError(r, userID,
			fmt.Errorf("read spec access denied to user %d in spec %d", userID, specID))
		return nil, http.StatusForbidden
	}

	// TODO Finish owner_name, user_is_admin, user_is_contributor
	s := &Spec{
		RenderTime: time.Now(),
	}
	err = db.QueryRow(`
		SELECT spec.id, spec.owner_type, spec.owner_id, user_account.username,
		user_account.user_settings::json#>>'{userProfile,highlightUsername}' AS highlight,
		spec.spec_name, spec.spec_desc, spec.is_public, spec.created_at,
		CASE
			-- when editor
			WHEN spec.owner_type = $2 AND spec.owner_id = $3
				THEN spec.updated_at
			-- when visitor
			ELSE GREATEST(spec.updated_at, spec.blocks_updated_at)
		END AS last_updated
		FROM spec
		LEFT JOIN user_account
		ON spec.owner_type=$2
		AND user_account.id=spec.owner_id
		WHERE spec.id=$1
		`, specID, OwnerTypeUser, userID,
	).Scan(&s.ID, &s.OwnerType, &s.OwnerID, &s.Username, &s.Highlight,
		&s.Name, &s.Desc, &s.Public, &s.Created, &s.Updated)
	if err != nil {
		logError(r, userID, fmt.Errorf("reading spec: %w", err))
		return nil, http.StatusInternalServerError
	}

	if AtoBool(query.Get("loadBlocks")) {
		s.Blocks, err = loadBlocks(db, specID, nil)
		if err != nil {
			logError(r, userID, fmt.Errorf("loading spec blocks: %w", err))
			return nil, http.StatusInternalServerError
		}
	}

	return s, http.StatusOK
}

// Returns the ID of the newly created spec.
func ajaxCreateSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	err := r.ParseForm()
	if err != nil {
		logError(r, userID, err)
		return nil, http.StatusInternalServerError
	}

	// TODO ALlow creating within an org

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		logError(r, userID, fmt.Errorf("spec name required"))
		return nil, http.StatusBadRequest
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	isPublic := AtoBool(r.Form.Get("isPublic"))

	return inTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		var specID int64

		err := tx.QueryRow(`
				INSERT INTO spec (owner_type, owner_id, created_at, updated_at, spec_name, spec_desc, is_public)
				VALUES ($1, $2, $3, $3, $4, $5, $6)
				RETURNING id
				`, OwnerTypeUser, userID, time.Now(), name, desc, isPublic).Scan(&specID)

		if err != nil {
			logError(r, userID, fmt.Errorf("creating spec: %w", err))
			return nil, http.StatusInternalServerError
		}

		return specID, http.StatusCreated
	})
}

func ajaxSaveSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	err := r.ParseForm()
	if err != nil {
		logError(r, userID, err)
		return nil, http.StatusInternalServerError
	}

	specID, err := AtoInt64(r.Form.Get("specId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing specId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, err := verifyWriteSpec(db, userID, specID); !access || err != nil {
		if err != nil {
			logError(r, userID, fmt.Errorf("validating write spec access: %w", err))
			return nil, http.StatusInternalServerError
		}
		logError(r, userID,
			fmt.Errorf("write spec access denied to user %d in spec %d", userID, specID))
		return nil, http.StatusForbidden
	}

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		logError(r, userID, fmt.Errorf("spec name required"))
		return nil, http.StatusBadRequest
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	isPublic := AtoBool(r.Form.Get("isPublic"))

	return inTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		spec := &Spec{
			ID:     specID,
			Public: isPublic,
		}

		err := tx.QueryRow(`
			UPDATE spec
			SET updated_at=$2, spec_name=$3, spec_desc=$4, is_public=$5
			WHERE id=$1
			RETURNING updated_at, spec_name, spec_desc
			`, specID, time.Now(), name, desc, isPublic).Scan(&spec.Updated, &spec.Name, &spec.Desc)

		if err != nil {
			logError(r, userID, fmt.Errorf("updating spec: %w", err))
			return nil, http.StatusInternalServerError
		}

		return spec, http.StatusOK
	})
}

func ajaxDeleteSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	err := r.ParseForm()
	if err != nil {
		logError(r, userID, err)
		return nil, http.StatusInternalServerError
	}

	specID, err := AtoInt64(r.Form.Get("specId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing specId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, err := verifyWriteSpec(db, userID, specID); !access || err != nil {
		if err != nil {
			logError(r, userID, fmt.Errorf("validating write spec access: %w", err))
			return nil, http.StatusInternalServerError
		}
		logError(r, userID,
			fmt.Errorf("write spec access denied to user %d in spec %d", userID, specID))
		return nil, http.StatusForbidden
	}

	return inTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		_, err := tx.Exec(`DELETE FROM spec WHERE id=$1`, specID)

		if err != nil {
			logError(r, userID, fmt.Errorf("deleting spec: %w", err))
			return nil, http.StatusInternalServerError
		}

		return nil, http.StatusOK
	})
}
