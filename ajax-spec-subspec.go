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

	if access, status := verifyReadSubspec(r, db, userID, subspecID); !access {
		return nil, status
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
		END AS last_updated,
		-- select number of unread comments
		(SELECT COUNT(*) FROM spec_community_comment AS c
			LEFT JOIN spec_community_read AS r
				ON r.user_id = $4 AND r.target_type = 'comment' AND r.target_id = c.id
			WHERE c.target_type = 'subspec' AND c.target_id = spec_subspec.id
				AND r.user_id IS NULL
		) AS unread_count,
		-- select total number of comments
		(SELECT COUNT(*)
			FROM spec_community_comment AS c
			WHERE c.target_type = 'subspec' AND c.target_id = spec_subspec.id
		) AS comments_count
		FROM spec_subspec
		INNER JOIN spec ON spec.id = $1
		WHERE spec_subspec.id = $2
			AND spec_subspec.spec_id = $1`,
		specID, subspecID, OwnerTypeUser, userID,
	).Scan(&s.Created, &s.Name, &s.Desc, &s.Updated, &s.UnreadCount, &s.CommentsCount)
	if err != nil {
		logError(r, userID, fmt.Errorf("reading subspec: %w", err))
		return nil, http.StatusInternalServerError
	}

	if AtoBool(query.Get("loadBlocks")) {
		s.Blocks, err = loadContextBlocks(db, userID, specID, &subspecID)
		if err != nil {
			logError(r, userID, fmt.Errorf("loading subspec blocks: %w", err))
			return nil, http.StatusInternalServerError
		}
	}

	return s, http.StatusOK
}

// load a list of subspecs for nav or block ref editing
func ajaxSubspecs(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// GET
	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing specId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyReadSpec(r, db, userID, specID); !access {
		return nil, status
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

	if access, status := verifyWriteSpec(r, db, userID, specID); !access {
		return nil, status
	}

	name := Substr(strings.TrimSpace(r.Form.Get("name")), subspecNameMaxLen)
	if name == "" {
		logError(r, userID, fmt.Errorf("subspec name required"))
		return nil, http.StatusBadRequest
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	return handleInTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		var subspecID int64

		err = tx.QueryRow(
			`INSERT INTO spec_subspec (
				spec_id, created_at, updated_at, subspec_name, subspec_desc
			) VALUES (
				$1, $2, $2, $3, $4
			) RETURNING id`,
			specID, time.Now(), name, desc,
		).Scan(&subspecID)

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

	if access, status := verifyWriteSubspec(r, db, userID, subspecID); !access {
		return nil, status
	}

	name := Substr(strings.TrimSpace(r.Form.Get("name")), subspecNameMaxLen)
	if name == "" {
		logError(r, userID, fmt.Errorf("subspec name required"))
		return nil, http.StatusBadRequest
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	return handleInTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		s := &SpecSubspec{
			ID: subspecID,
		}

		// Scan new values as represented in DB for return
		err = tx.QueryRow(
			`UPDATE spec_subspec
			SET updated_at = $2, subspec_name = $3, subspec_desc = $4
			WHERE id = $1
			RETURNING spec_id, created_at, updated_at, subspec_name, subspec_desc,
			-- select number of unread comments
			(SELECT COUNT(*) FROM spec_community_comment AS c
				LEFT JOIN spec_community_read AS r
					ON r.user_id = $1 AND r.target_type = 'comment' AND r.target_id = c.id
				WHERE c.target_type = 'subspec' AND c.target_id = spec_subspec.id
					AND r.user_id IS NULL
			) AS unread_count,
			-- select total number of comments
			(SELECT COUNT(*)
				FROM spec_community_comment AS c
				WHERE c.target_type = 'subspec' AND c.target_id = spec_subspec.id
			) AS comments_count`,
			subspecID, time.Now(), name, desc,
		).Scan(&s.SpecID, &s.Created, &s.Updated, &s.Name, &s.Desc, &s.UnreadCount, &s.CommentsCount)

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

	if access, status := verifyWriteSubspec(r, db, userID, subspecID); !access {
		return nil, status
	}

	return handleInTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

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
