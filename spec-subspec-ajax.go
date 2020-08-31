package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func ajaxSubspecs(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET
	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid specId: %w", err)
	}

	if access, err := verifyReadSpec(db, userID, specID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying read spec: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("read spec access denied to user %d in spec %d", userID, specID)
	}

	rows, err := db.Query(`
		SELECT id, created_at, updated_at, subspec_name, subspec_desc
		FROM spec_subspec
		WHERE spec_id = $1
		ORDER BY subspec_name`, specID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("querying blocks: %w", err)
	}

	subspecs := []*SpecSubspec{}

	for rows.Next() {
		s := &SpecSubspec{
			SpecID: specID,
		}
		err = rows.Scan(&s.ID, &s.Created, &s.Updated, &s.Name, &s.Desc)
		if err != nil {
			if err2 := rows.Close(); err2 != nil { // TODO Add everywhere
				return nil, http.StatusInternalServerError, fmt.Errorf("error closing rows: %s; on scan error: %w", err2, err)
			}
			return nil, http.StatusInternalServerError, fmt.Errorf("error scanning subspec: %w", err)
		}
		subspecs = append(subspecs, s)
	}

	return subspecs, http.StatusOK, nil
}

func ajaxSpecCreateSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
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
		return nil, http.StatusBadRequest, fmt.Errorf("subspec name required")
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		var subspecID int64

		err = tx.QueryRow(`INSERT INTO spec_subspec (spec_id, created_at, updated_at, subspec_name, subspec_desc)
			VALUES ($1, $2, $2, $3, $4) RETURNING id`,
			specID, time.Now(), name, desc).Scan(&subspecID)

		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("inserting new subspec: %w", err)
		}

		return subspecID, http.StatusCreated, nil
	})
}

func ajaxSpecSaveSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	subspecID, err := AtoInt64(r.Form.Get("subspecId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing subspecId: %w", err)
	}

	if access, err := verifyWriteSubspec(db, userID, subspecID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying write subspec: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("write subspec access denied to user %d in subspec %d", userID, subspecID)
	}

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("subspec name required")
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		s := &SpecSubspec{
			ID: subspecID,
		}

		// Scan values as represented in DB for return
		err = tx.QueryRow(`UPDATE spec_subspec SET updated_at = $2,
			subspec_name = $3, subspec_desc = $4
			WHERE id = $1
			RETURNING spec_id, updated_at, subspec_name, subspec_desc`,
			subspecID, time.Now(), name, desc).Scan(&s.SpecID, &s.Updated, &s.Name, &s.Desc)

		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("updating subspec: %w", err)
		}

		return s, http.StatusOK, nil
	})
}

func ajaxSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET
	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid specId: %w", err)
	}

	subspecID, err := AtoInt64(query.Get("subspecId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid subspecId: %w", err)
	}

	if access, err := verifyReadSubspec(db, userID, subspecID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying read subspec: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("read subspec access denied to user %d in subspec %d", userID, subspecID)
	}

	s := &SpecSubspec{
		ID:     subspecID,
		SpecID: specID,
	}

	err = db.QueryRow(`SELECT spec_subspec.created_at,
		spec_subspec.subspec_name, spec_subspec.subspec_desc,
		spec.spec_name, spec.owner_type, spec.owner_id,
		CASE
			-- when editor
			WHEN spec.owner_type = $3 AND spec.owner_id = $4
				THEN spec_subspec.updated_at
			-- when visitor
			ELSE GREATEST(spec_subspec.updated_at, (
				SELECT updated_at FROM spec_block
				WHERE spec_block.spec_id = spec.id
				AND spec_block.subspec_id = spec_subspec.id
				ORDER BY updated_at DESC
				LIMIT 1
			))
		END AS last_updated
		FROM spec_subspec
		INNER JOIN spec ON spec.id = $1
		WHERE spec_subspec.id = $2
		AND spec_subspec.spec_id = $1`,
		specID, subspecID, OwnerTypeUser, userID,
	).Scan(&s.Created, &s.Name, &s.Desc,
		&s.SpecName, &s.OwnerType, &s.OwnerID, &s.Updated)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("loading subspec: %w", err)
	}

	s.Blocks, err = loadBlocks(db, specID, &subspecID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("loading subspec block: %w", err)
	}

	return s, http.StatusOK, nil
}

func ajaxSpecDeleteSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	subspecID, err := AtoInt64(r.Form.Get("subspecId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing subspecId: %w", err)
	}

	if access, err := verifyWriteSubspec(db, userID, subspecID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying write subspec: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("write subspec access denied to user %d in subspec %d", userID, subspecID)
	}

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		// Leave block references

		_, err := tx.Exec(`
			DELETE FROM spec_subspec
			WHERE id=$1
			`, subspecID)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("deleting subspec: %w", err)
		}

		return nil, http.StatusOK, nil
	})
}
