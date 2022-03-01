package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func ajaxSubspec(db *sql.DB, userID *uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
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

	if access, status := verifyReadTarget(r, db, userID, CommunityTargetSubspec, subspecID); !access {
		return nil, status
	}

	s := &SpecSubspec{
		ID:         subspecID,
		SpecID:     specID,
		RenderTime: time.Now(),
	}

	var args = []interface{}{specID, subspecID}

	var lastUpdatedField string
	if userID == nil {
		lastUpdatedField = `GREATEST(spec_subspec.updated_at, spec_subspec.blocks_updated_at) AS last_updated`
	} else {
		lastUpdatedField = `CASE
			-- when editor
			WHEN spec.owner_type = ` + argPlaceholder(OwnerTypeUser, &args) + `
				AND spec.owner_id = ` + argPlaceholder(*userID, &args) + `
			THEN spec_subspec.updated_at
			-- when visitor
			ELSE GREATEST(spec_subspec.updated_at, spec_subspec.blocks_updated_at)
		END AS last_updated`
	}

	var unreadCountField string
	if userID == nil {
		unreadCountField = `0 AS unread_count`
	} else {
		unreadCountField = `(SELECT COUNT(*) FROM spec_community_comment AS c
			LEFT JOIN spec_community_read AS r
				ON r.user_id = ` + argPlaceholder(*userID, &args) + `
				AND r.target_type = ` + argPlaceholder(CommunityTargetComment, &args) + `
				AND r.target_id = c.id
			WHERE c.target_type = ` + argPlaceholder(CommunityTargetSubspec, &args) + `
				 AND c.target_id = spec_subspec.id
				AND r.user_id IS NULL
		) AS unread_count`
	}

	err = db.QueryRow(`SELECT spec_subspec.created_at,
		spec_subspec.subspec_name, spec_subspec.subspec_desc, spec_subspec.is_private,
		-- spec.spec_name, spec.owner_type, spec.owner_id,
		-- select last updated
		`+lastUpdatedField+`,
		-- select number of unread comments
		`+unreadCountField+`,
		-- select total number of comments
		(SELECT COUNT(*) FROM spec_community_comment AS c
			WHERE c.target_type = `+argPlaceholder(CommunityTargetSubspec, &args)+`
			AND c.target_id = spec_subspec.id
		) AS comments_count
		FROM spec_subspec
		INNER JOIN spec ON spec.id = $1
		WHERE spec_subspec.id = $2
			AND spec_subspec.spec_id = $1`,
		args...,
	).Scan(&s.Created, &s.Name, &s.Desc, &s.Private,
		&s.Updated, &s.UnreadCount, &s.CommentsCount)
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
func ajaxSubspecs(db *sql.DB, userID *uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// GET
	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing specId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyReadTarget(r, db, userID, CommunityTargetSpec, specID); !access {
		return nil, status
	}

	var args []interface{}

	var publicSubspecCond string
	if userID == nil {
		// The public may not be allowed to view this subspec.
		publicSubspecCond = `(NOT spec_subspec.is_private)`
	} else {
		// Only the owner may view private subspecs
		publicSubspecCond = `((NOT spec_subspec.is_private) OR (
			spec.owner_type = ` + argPlaceholder(OwnerTypeUser, &args) + `
			AND spec.owner_id = ` + argPlaceholder(*userID, &args) + `
		))`
	}

	rows, err := db.Query(
		`SELECT spec_subspec.id, spec_subspec.created_at, spec_subspec.updated_at,
			spec_subspec.subspec_name, spec_subspec.subspec_desc, spec_subspec.is_private
		FROM spec_subspec
		INNER JOIN spec ON spec.id = spec_subspec.spec_id
		WHERE spec.id = `+argPlaceholder(specID, &args)+`
			AND `+publicSubspecCond+`
		ORDER BY spec_subspec.is_private, spec_subspec.subspec_name`,
		args...,
	)
	if err != nil {
		logError(r, userID, fmt.Errorf("querying subspecs: %w", err))
		return nil, http.StatusInternalServerError
	}

	subspecs := []*SpecSubspec{}

	for rows.Next() {
		s := &SpecSubspec{
			SpecID: specID,
		}
		err = rows.Scan(&s.ID, &s.Created, &s.Updated, &s.Name, &s.Desc, &s.Private)
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
		logError(r, &userID, err)
		return nil, http.StatusInternalServerError
	}

	specID, err := AtoInt64(r.Form.Get("specId"))
	if err != nil {
		logError(r, &userID, fmt.Errorf("parsing specId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyCreateSubspec(r, db, userID, specID); !access {
		return nil, status
	}

	name := Substr(strings.TrimSpace(r.Form.Get("name")), subspecNameMaxLen)
	if name == "" {
		logError(r, &userID, fmt.Errorf("subspec name required"))
		return nil, http.StatusBadRequest
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	private := AtoBool(strings.TrimSpace(r.Form.Get("private")))

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		var subspec = &SpecSubspec{}

		err = tx.QueryRow(
			`INSERT INTO spec_subspec (
				spec_id, created_at, updated_at, subspec_name, subspec_desc, is_private
			) VALUES (
				$1, $2, $2, $3, $4, $5
			) RETURNING id, spec_id, created_at, updated_at,
				subspec_name, subspec_desc, is_private`,
			specID, time.Now(), name, desc, private,
		).Scan(&subspec.ID, &subspec.SpecID, &subspec.Created, &subspec.Updated,
			&subspec.Name, &subspec.Desc, &subspec.Private)

		if err != nil {
			logError(r, &userID, fmt.Errorf("creating subspec: %w", err))
			return nil, http.StatusInternalServerError
		}

		return subspec, http.StatusCreated
	})
}

func ajaxSpecSaveSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	err := r.ParseForm()
	if err != nil {
		logError(r, &userID, err)
		return nil, http.StatusInternalServerError
	}

	subspecID, err := AtoInt64(r.Form.Get("subspecId"))
	if err != nil {
		logError(r, &userID, fmt.Errorf("parsing subspecId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyWriteTarget(r, db, userID, CommunityTargetSubspec, subspecID); !access {
		return nil, status
	}

	name := Substr(strings.TrimSpace(r.Form.Get("name")), subspecNameMaxLen)
	if name == "" {
		logError(r, &userID, fmt.Errorf("subspec name required"))
		return nil, http.StatusBadRequest
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	private := AtoBool(strings.TrimSpace(r.Form.Get("private")))

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		s := &SpecSubspec{
			ID: subspecID,
		}

		var args = []interface{}{}

		// Scan new values as represented in DB for return
		err = tx.QueryRow(`UPDATE spec_subspec
			SET updated_at = `+argPlaceholder(time.Now(), &args)+`,
				subspec_name = `+argPlaceholder(name, &args)+`,
				subspec_desc = `+argPlaceholder(desc, &args)+`,
				is_private = `+argPlaceholder(private, &args)+`
			WHERE id = `+argPlaceholder(subspecID, &args)+`
			RETURNING spec_id, created_at, updated_at,
				subspec_name, subspec_desc, is_private,
			-- select number of unread comments
			(SELECT COUNT(*) FROM spec_community_comment AS c
				LEFT JOIN spec_community_read AS r
					ON r.user_id = `+argPlaceholder(userID, &args)+`
					AND r.target_type = `+argPlaceholder(CommunityTargetComment, &args)+`
					AND r.target_id = c.id
				WHERE c.target_type = `+argPlaceholder(CommunityTargetSubspec, &args)+`
					AND c.target_id = spec_subspec.id
					AND r.user_id IS NULL
			) AS unread_count,
			-- select total number of comments
			(SELECT COUNT(*)
				FROM spec_community_comment AS c
				WHERE c.target_type = `+argPlaceholder(CommunityTargetSubspec, &args)+`
					AND c.target_id = spec_subspec.id
			) AS comments_count`,
			args...,
		).Scan(&s.SpecID, &s.Created, &s.Updated, &s.Name, &s.Desc, &s.Private,
			&s.UnreadCount, &s.CommentsCount)

		if err != nil {
			logError(r, &userID, fmt.Errorf("updating subspec: %w", err))
			return nil, http.StatusInternalServerError
		}

		return s, http.StatusOK
	})
}

func ajaxSpecDeleteSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	err := r.ParseForm()
	if err != nil {
		logError(r, &userID, err)
		return nil, http.StatusInternalServerError
	}

	subspecID, err := AtoInt64(r.Form.Get("subspecId"))
	if err != nil {
		logError(r, &userID, fmt.Errorf("parsing subspecId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyDeleteTarget(r, db, userID, CommunityTargetSubspec, subspecID); !access {
		return nil, status
	}

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		// Blocks in subspec are deleted on cascade

		// Don't clear references from blocks - display "content unavailable" message

		_, err := tx.Exec(`DELETE FROM spec_subspec WHERE id=$1`, subspecID)

		if err != nil {
			logError(r, &userID, fmt.Errorf("deleting subspec: %w", err))
			return nil, http.StatusInternalServerError
		}

		return nil, http.StatusOK
	})
}
