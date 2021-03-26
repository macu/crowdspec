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
		ORDER BY is_public DESC, GREATEST(updated_at, blocks_updated_at) DESC
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
			if err2 := rows.Close(); err2 != nil {
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
		ORDER BY GREATEST(spec.updated_at, spec.blocks_updated_at) DESC
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
			if err2 := rows.Close(); err2 != nil {
				logError(r, userID, fmt.Errorf("closing rows: %s; on scan error: %w", err2, err))
				return nil, http.StatusInternalServerError
			}
			logError(r, userID, fmt.Errorf("scanning spec: %w", err))
			return nil, http.StatusInternalServerError
		}
		publicSpecs = append(publicSpecs, s)
	}

	var unreadComments uint
	err = db.QueryRow(
		`SELECT (
			-- count unread comments on user's own specs
			SELECT COUNT(*)
			FROM spec
			INNER JOIN spec_community_comment cc
				ON cc.spec_id = spec.id
			WHERE spec.owner_type = $1 AND spec.owner_id = $2
				AND cc.target_type != 'comment' -- counted separately
				AND NOT EXISTS(
					SELECT * FROM spec_community_read cr
					WHERE cr.user_id = $2
						AND cr.target_type = 'comment'
						AND cr.target_id = cc.id
				)
		) + (
			-- count comments by current user with unread replies
			SELECT COUNT(*)
			FROM spec_community_comment cc_user
			INNER JOIN spec_community_comment cc_reply
				ON cc_reply.target_type = 'comment' AND cc_reply.target_id = cc_user.id
			WHERE cc_user.user_id = $2
				AND NOT EXISTS(
					SELECT * FROM spec_community_read cr
					WHERE cr.user_id = $2
						AND cr.target_type = 'comment'
						AND cr.target_id = cc_reply.id
				)
		)`,
		OwnerTypeUser, userID,
	).Scan(&unreadComments)
	if err != nil {
		logError(r, userID, fmt.Errorf("counting unread comments: %w", err))
		return nil, http.StatusInternalServerError
	}

	payload := struct {
		UnreadCommentCount uint   `json:"unread"`
		UserSpecs          []Spec `json:"userSpecs"`
		PublicSpecs        []Spec `json:"publicSpecs"`
	}{
		UnreadCommentCount: unreadComments,
		UserSpecs:          userSpecs,
		PublicSpecs:        publicSpecs,
	}

	return payload, http.StatusOK
}
