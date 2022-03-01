package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Returns the requested spec with immediate blocks.
func ajaxSpec(db *sql.DB, userID *uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
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

	var args = []interface{}{specID}

	var lastUpdatedField string
	if userID == nil {
		lastUpdatedField = `GREATEST(spec.updated_at, spec.blocks_updated_at) AS last_updated`
	} else {
		lastUpdatedField = `CASE
			-- when editor
			WHEN spec.owner_type = ` + argPlaceholder(OwnerTypeUser, &args) + `
				AND spec.owner_id = ` + argPlaceholder(*userID, &args) + `
			THEN spec.updated_at
			-- when visitor
			ELSE GREATEST(spec.updated_at, spec.blocks_updated_at)
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
			WHERE c.target_type = 'spec' AND c.target_id = spec.id
				AND r.user_id IS NULL
		) AS unread_count`
	}

	// TODO Finish owner_name, user_is_admin, user_is_contributor
	s := &Spec{
		RenderTime: time.Now(),
	}
	err = db.QueryRow(`
		SELECT spec.id, spec.owner_type, spec.owner_id, user_account.username,
		user_account.user_settings::json#>>'{userProfile,highlightUsername}' AS highlight,
		spec.spec_name, spec.spec_desc, spec.is_public, spec.created_at,
		-- select last updated
		`+lastUpdatedField+`,
		-- select number of unread comments
		`+unreadCountField+`,
		-- select total number of comments
		(SELECT COUNT(*) FROM spec_community_comment AS c
			WHERE c.target_type = `+argPlaceholder(CommunityTargetSpec, &args)+`
			AND c.target_id = spec.id
		) AS comments_count
		FROM spec
		LEFT JOIN user_account
			ON spec.owner_type=`+argPlaceholder(OwnerTypeUser, &args)+`
			AND user_account.id=spec.owner_id
		WHERE spec.id=$1`,
		args...,
	).Scan(&s.ID, &s.OwnerType, &s.OwnerID, &s.Username, &s.Highlight,
		&s.Name, &s.Desc, &s.Public, &s.Created, &s.Updated, &s.UnreadCount, &s.CommentsCount)
	if err != nil {
		logError(r, userID, fmt.Errorf("reading spec: %w", err))
		return nil, http.StatusInternalServerError
	}

	if AtoBool(query.Get("loadBlocks")) {
		s.Blocks, err = loadContextBlocks(db, userID, specID, nil)
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
		logError(r, &userID, err)
		return nil, http.StatusInternalServerError
	}

	// TODO ALlow creating within an org

	if access, status := verifyCreateSpec(r, db, userID); !access {
		return nil, status
	}

	name := Substr(strings.TrimSpace(r.Form.Get("name")), spenNameMaxLen)
	if name == "" {
		logError(r, &userID, fmt.Errorf("spec name required"))
		return nil, http.StatusBadRequest
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	isPublic := AtoBool(r.Form.Get("isPublic"))

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		var specID int64

		err := tx.QueryRow(`
				INSERT INTO spec (owner_type, owner_id, created_at, updated_at, spec_name, spec_desc, is_public)
				VALUES ($1, $2, $3, $3, $4, $5, $6)
				RETURNING id
				`, OwnerTypeUser, userID, time.Now(), name, desc, isPublic).Scan(&specID)

		if err != nil {
			logError(r, &userID, fmt.Errorf("creating spec: %w", err))
			return nil, http.StatusInternalServerError
		}

		return specID, http.StatusCreated
	})
}

func ajaxSaveSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
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

	if access, status := verifyWriteTarget(r, db, userID, CommunityTargetSpec, specID); !access {
		return nil, status
	}

	name := Substr(strings.TrimSpace(r.Form.Get("name")), spenNameMaxLen)
	if name == "" {
		logError(r, &userID, fmt.Errorf("spec name required"))
		return nil, http.StatusBadRequest
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	isPublic := AtoBool(r.Form.Get("isPublic"))

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		spec := &Spec{
			ID:     specID,
			Public: isPublic,
		}

		err := tx.QueryRow(
			`UPDATE spec
			SET updated_at=$2, spec_name=$3, spec_desc=$4, is_public=$5
			WHERE id=$1
			RETURNING updated_at, spec_name, spec_desc,
			-- select number of unread comments
			(SELECT COUNT(*) FROM spec_community_comment AS c
				LEFT JOIN spec_community_read AS r
					ON r.user_id = $1 AND r.target_type = 'comment' AND r.target_id = c.id
				WHERE c.target_type = 'spec' AND c.target_id = spec.id
					AND r.user_id IS NULL
			) AS unread_count,
			-- select total number of comments
			(SELECT COUNT(*)
				FROM spec_community_comment AS c
				WHERE c.target_type = 'spec' AND c.target_id = spec.id
			) AS comments_count`,
			specID, time.Now(), name, desc, isPublic,
		).Scan(&spec.Updated, &spec.Name, &spec.Desc, &spec.UnreadCount, &spec.CommentsCount)

		if err != nil {
			logError(r, &userID, fmt.Errorf("updating spec: %w", err))
			return nil, http.StatusInternalServerError
		}

		return spec, http.StatusOK
	})
}

func ajaxDeleteSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
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

	if access, status := verifyDeleteTarget(r, db, userID, CommunityTargetSpec, specID); !access {
		return nil, status
	}

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		_, err := tx.Exec(`DELETE FROM spec WHERE id=$1`, specID)

		if err != nil {
			logError(r, &userID, fmt.Errorf("deleting spec: %w", err))
			return nil, http.StatusInternalServerError
		}

		return nil, http.StatusOK
	})
}
