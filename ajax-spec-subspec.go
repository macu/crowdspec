package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func ajaxSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// GET
	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing specId: %w", err))
		return nil, http.StatusBadRequest
	}

	subspecID, err := AtoInt64(query.Get("subspecId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing subspecId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, err := verifyReadSubspec(db, userID, subspecID); !access || err != nil {
		if err != nil {
			logError(r, userID, fmt.Errorf("validating read subspe accessc: %w", err))
			return nil, http.StatusInternalServerError
		}
		logError(r, userID,
			fmt.Errorf("read subspec access denied to user %d in subspec %d", userID, subspecID))
		return nil, http.StatusForbidden
	}

	s := &SpecSubspec{
		ID:         subspecID,
		SpecID:     specID,
		RenderTime: time.Now(),
	}

	err = db.QueryRow(`SELECT spec_subspec.created_at,
		spec_subspec.subspec_name, spec_subspec.subspec_desc,
		-- spec.spec_name, spec.owner_type, spec.owner_id,
		CASE
			-- when editor
			WHEN spec.owner_type = $3 AND spec.owner_id = $4
				THEN spec_subspec.updated_at
			-- when visitor
			ELSE GREATEST(spec_subspec.updated_at, spec_subspec.blocks_updated_at)
		END AS last_updated
		FROM spec_subspec
		INNER JOIN spec ON spec.id = $1
		WHERE spec_subspec.id = $2
		AND spec_subspec.spec_id = $1`,
		specID, subspecID, OwnerTypeUser, userID,
	).Scan(&s.Created, &s.Name, &s.Desc, &s.Updated)
	if err != nil {
		logError(r, userID, fmt.Errorf("reading subspec: %w", err))
		return nil, http.StatusInternalServerError
	}

	if AtoBool(query.Get("loadBlocks")) {
		s.Blocks, err = loadBlocks(db, specID, &subspecID)
		if err != nil {
			logError(r, userID, fmt.Errorf("loading subspec blocks: %w", err))
			return nil, http.StatusInternalServerError
		}
	}

	return s, http.StatusOK
}

func ajaxSubspecs(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
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

	rows, err := db.Query(`
		SELECT id, created_at, updated_at, subspec_name, subspec_desc
		FROM spec_subspec
		WHERE spec_id = $1
		ORDER BY subspec_name`, specID)
	if err != nil {
		logError(r, userID, fmt.Errorf("querying subspecs: %w", err))
		return nil, http.StatusInternalServerError
	}

	subspecs := []*SpecSubspec{}

	for rows.Next() {
		s := &SpecSubspec{
			SpecID: specID,
		}
		err = rows.Scan(&s.ID, &s.Created, &s.Updated, &s.Name, &s.Desc)
		if err != nil {
			if err2 := rows.Close(); err2 != nil {
				logError(r, userID, fmt.Errorf("closing rows: %s; on scan error: %w", err2, err))
				return nil, http.StatusInternalServerError
			}
			logError(r, userID, fmt.Errorf("scanning subspec: %w", err))
			return nil, http.StatusInternalServerError
		}
		subspecs = append(subspecs, s)
	}

	return subspecs, http.StatusOK
}

func ajaxSpecCreateSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
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
		logError(r, userID, fmt.Errorf("write spec access denied to user %d in spec %d", userID, specID))
		return nil, http.StatusForbidden
	}

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		logError(r, userID, fmt.Errorf("subspec name required"))
		return nil, http.StatusBadRequest
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	return inTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		var subspecID int64

		err = tx.QueryRow(`INSERT INTO spec_subspec (spec_id, created_at, updated_at, subspec_name, subspec_desc)
			VALUES ($1, $2, $2, $3, $4) RETURNING id`,
			specID, time.Now(), name, desc).Scan(&subspecID)

		if err != nil {
			logError(r, userID, fmt.Errorf("creating subspec: %w", err))
			return nil, http.StatusInternalServerError
		}

		return subspecID, http.StatusCreated
	})
}

func ajaxSpecSaveSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	err := r.ParseForm()
	if err != nil {
		logError(r, userID, err)
		return nil, http.StatusInternalServerError
	}

	subspecID, err := AtoInt64(r.Form.Get("subspecId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing subspecId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, err := verifyWriteSubspec(db, userID, subspecID); !access || err != nil {
		if err != nil {
			logError(r, userID, fmt.Errorf("validating write subspec access: %w", err))
			return nil, http.StatusInternalServerError
		}
		logError(r, userID,
			fmt.Errorf("write subspec access denied to user %d in subspec %d", userID, subspecID))
		return nil, http.StatusForbidden
	}

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		logError(r, userID, fmt.Errorf("subspec name required"))
		return nil, http.StatusBadRequest
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	return inTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		s := &SpecSubspec{
			ID: subspecID,
		}

		// Scan new values as represented in DB for return
		err = tx.QueryRow(`UPDATE spec_subspec SET updated_at = $2,
			subspec_name = $3, subspec_desc = $4
			WHERE id = $1
			RETURNING spec_id, updated_at, subspec_name, subspec_desc`,
			subspecID, time.Now(), name, desc).Scan(&s.SpecID, &s.Updated, &s.Name, &s.Desc)

		if err != nil {
			logError(r, userID, fmt.Errorf("updating subspec: %w", err))
			return nil, http.StatusInternalServerError
		}

		return s, http.StatusOK
	})
}

func ajaxSpecDeleteSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	err := r.ParseForm()
	if err != nil {
		logError(r, userID, err)
		return nil, http.StatusInternalServerError
	}

	subspecID, err := AtoInt64(r.Form.Get("subspecId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing subspecId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, err := verifyWriteSubspec(db, userID, subspecID); !access || err != nil {
		if err != nil {
			logError(r, userID, fmt.Errorf("validating write subspec access: %w", err))
			return nil, http.StatusInternalServerError
		}
		logError(r, userID,
			fmt.Errorf("write subspec access denied to user %d in subspec %d", userID, subspecID))
		return nil, http.StatusForbidden
	}

	return inTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		// Blocks in subspec are deleted on cascade

		// Don't clear references from blocks - display "content unavailable" message

		_, err := tx.Exec(`
			DELETE FROM spec_subspec
			WHERE id=$1
			`, subspecID)

		if err != nil {
			logError(r, userID, fmt.Errorf("deleting subspec: %w", err))
			return nil, http.StatusInternalServerError
		}

		return nil, http.StatusOK
	})
}
