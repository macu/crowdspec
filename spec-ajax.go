package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Returns a list of the current user's specs.
// func ajaxUserSpecs(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
// 	// GET
// 	rows, err := db.Query(`
// 		SELECT id, owner_type, owner_id, created_at, spec_name, spec_desc, is_public
// 		FROM spec
// 		WHERE owner_type='user' AND owner_id=$1
// 		ORDER BY created_at DESC
// 		`, userID)
// 	if err != nil {
// 		return nil, http.StatusInternalServerError, err
// 	}
//
// 	specs := []Spec{}
// 	for rows.Next() {
// 		s := Spec{}
// 		err = rows.Scan(&s.ID, &s.OwnerType, &s.OwnerID, &s.Created, &s.Name, &s.Desc, &s.Public)
// 		if err != nil {
// 			return nil, http.StatusInternalServerError, err
// 		}
// 		specs = append(specs, s)
// 	}
//
// 	return specs, http.StatusOK, nil
// }

// Returns the ID of the newly created spec.
func ajaxCreateSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// TODO ALlow creating within an org

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("Blank spec name")
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	isPublic := AtoBool(r.Form.Get("isPublic"))

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		var specID int64

		err := tx.QueryRow(`
				INSERT INTO spec (owner_type, owner_id, created_at, updated_at, spec_name, spec_desc, is_public)
				VALUES ($1, $2, $3, $3, $4, $5, $6)
				RETURNING id
				`, OwnerTypeUser, userID, time.Now(), name, desc, isPublic).Scan(&specID)

		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		return specID, http.StatusCreated, nil
	})
}

func ajaxSaveSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specID, err := AtoInt64(r.Form.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing specId: %w", err)
	}

	if access, err := verifyWriteSpec(db, userID, specID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying write spec: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("write spec access denied to user %d in spec %d", userID, specID)
	}

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("Blank spec name")
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	isPublic := AtoBool(r.Form.Get("isPublic"))

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

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
			return nil, http.StatusInternalServerError, fmt.Errorf("updating spec: %w", err)
		}

		return spec, http.StatusOK, nil
	})
}

func ajaxDeleteSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specID, err := AtoInt64(r.Form.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing specId: %w", err)
	}

	if access, err := verifyWriteSpec(db, userID, specID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying write spec: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("write spec access denied to user %d in spec %d", userID, specID)
	}

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		_, err := tx.Exec(`DELETE FROM spec WHERE id=$1`, specID)

		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("deleting spec: %w", err)
		}

		return nil, http.StatusOK, nil
	})
}

// Returns the requested spec with immediate blocks.
func ajaxSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET
	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("Invalid specId: %w", err)
	}

	if access, err := verifyReadSpec(db, userID, specID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying read spec: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("read spec access denied to user %d in spec %d", userID, specID)
	}

	// TODO Finish owner_name, user_is_admin, user_is_contributor
	s := &Spec{}
	err = db.QueryRow(`
		SELECT spec.id, spec.created_at, spec.updated_at,
		spec.owner_type, spec.owner_id, user_account.username,
		spec.spec_name, spec.spec_desc, spec.is_public
		FROM spec
		LEFT JOIN user_account
		ON spec.owner_type=$2
		AND user_account.id=spec.owner_id
		WHERE spec.id=$1
		`, specID, OwnerTypeUser).Scan(&s.ID, &s.Created, &s.Updated,
		&s.OwnerType, &s.OwnerID, &s.Username,
		&s.Name, &s.Desc, &s.Public)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	s.Blocks, err = loadBlocks(db, specID, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return s, http.StatusOK, nil
}
