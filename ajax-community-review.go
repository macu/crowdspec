package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

const commentsReviewPageSize = 7

type communityReviewPage struct {
	Specs    *[]*communityReviewSpec    `json:"specs,omitempty"`
	Subspecs *[]*communityReviewSubspec `json:"subspecs,omitempty"`
	Comments *[]*communityReviewComment `json:"comments,omitempty"`

	// Comments properties
	TotalComments   *uint `json:"totalComments,omitempty"`
	HasMoreComments *bool `json:"hasMoreComments,omitempty"`

	RenderTime time.Time `json:"renderTime"`
}

type communityReviewSpec struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Updated time.Time `json:"updated"`

	UnreadComments uint `json:"unread"`
	TotalComments  uint `json:"total"`

	BlockUnreadComments uint `json:"blockUnread"`
	BlockTotalComments  uint `json:"blockTotal"`

	HasSubspecs              bool `json:"hasSubspecs"`
	HasUnreadSubspecComments bool `json:"hasUnreadSubspec"`
}

type communityReviewSubspec struct {
	SpecID  int64     `json:"specId"`
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Updated time.Time `json:"updated"`

	UnreadComments uint `json:"unread"`
	TotalComments  uint `json:"total"`

	BlockUnreadComments uint `json:"blockUnread"`
	BlockTotalComments  uint `json:"blockTotal"`
}

type communityReviewComment struct {
	SpecID  int64     `json:"specId"`
	ID      int64     `json:"id"`
	Body    string    `json:"body"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`

	UnreadComments uint `json:"unread"`
	TotalComments  uint `json:"total"`
}

func ajaxLoadCommuntyReviewPage(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// GET

	var request = r.FormValue("request")
	var all = request == "all"

	var response = &communityReviewPage{
		RenderTime: time.Now(),
	}

	if all {
		// Load specs review

		response.Specs = &[]*communityReviewSpec{}

		rows, err := db.Query(
			`SELECT spec.id, spec.spec_name, spec.updated_at,
				(
					SELECT COUNT(*)
					FROM spec_community_comment cc
					WHERE cc.target_type = 'spec'
						AND cc.target_id = spec.id
						AND NOT EXISTS(
							SELECT * FROM spec_community_read cr
							WHERE cr.user_id = $2
								AND cr.target_type = 'comment'
								AND cr.target_id = cc.id
						)
				) AS unread_comments,
				(
					SELECT COUNT(*)
					FROM spec_community_comment cc
					WHERE cc.target_type = 'spec'
						AND cc.target_id = spec.id
				) AS total_comments,
				(
					SELECT COUNT(*)
					FROM spec_community_comment cc
					INNER JOIN spec_block
						ON cc.target_type = 'block'
						AND spec_block.id = cc.target_id
						AND spec_block.spec_id = spec.id
						AND spec_block.subspec_id IS NULL
					WHERE cc.spec_id = spec.id
						AND NOT EXISTS(
							SELECT * FROM spec_community_read cr
							WHERE cr.user_id = $2
								AND cr.target_type = 'comment'
								AND cr.target_id = cc.id
						)
				) AS block_unread_comments,
				(
					SELECT COUNT(*)
					FROM spec_community_comment cc
					INNER JOIN spec_block
						ON cc.target_type = 'block'
						AND spec_block.id = cc.target_id
						AND spec_block.spec_id = spec.id
						AND spec_block.subspec_id IS NULL
					WHERE cc.spec_id = spec.id
				) AS block_total_comments,
				(
					SELECT EXISTS(
						SELECT * FROM spec_subspec
						WHERE spec_id = spec.id
					)
				) AS has_subspecs,
				(
					SELECT EXISTS(
						SELECT ss.*
						FROM spec_subspec ss
						INNER JOIN spec_community_comment cc
							ON cc.target_type = 'subspec'
							AND cc.target_id = ss.id
						WHERE ss.spec_id = spec.id
							AND NOT EXISTS(
								SELECT *
								FROM spec_community_read cr
								WHERE cr.user_id = $2
									AND cr.target_type = 'comment'
									AND cr.target_id = cc.id
							)
					) OR EXISTS(
						SELECT sb.*
						FROM spec_block sb
						INNER JOIN spec_community_comment cc
							ON cc.target_type = 'block'
							AND cc.target_id = sb.id
						WHERE sb.spec_id = spec.id
							AND sb.subspec_id IS NOT NULL
							AND NOT EXISTS(
								SELECT *
								FROM spec_community_read cr
								WHERE cr.user_id = $2
									AND cr.target_type = 'comment'
									AND cr.target_id = cc.id
							)
					)
				) AS has_subspec_unread_comments
			FROM spec
			WHERE spec.owner_type = $1 AND spec.owner_id = $2
			ORDER BY spec.spec_name ASC`,
			OwnerTypeUser, userID,
		)
		if err != nil {
			logError(r, &userID, fmt.Errorf("querying specs review: %w", err))
			return nil, http.StatusInternalServerError
		}

		for rows.Next() {
			var review = communityReviewSpec{}
			err = rows.Scan(&review.ID, &review.Name, &review.Updated,
				&review.UnreadComments, &review.TotalComments,
				&review.BlockUnreadComments, &review.BlockTotalComments,
				&review.HasSubspecs, &review.HasUnreadSubspecComments)
			if err != nil {
				logError(r, &userID, fmt.Errorf("scanning specs review row: %w", err))
				return nil, http.StatusInternalServerError
			}
			(*response.Specs) = append((*response.Specs), &review)
		}
	}

	if request == "subspecs" {
		// Load subspecs review within a spec

		var specID, err = AtoInt64(r.FormValue("specId"))
		if err != nil {
			logError(r, &userID, fmt.Errorf("parsing specId: %w", err))
			return nil, http.StatusBadRequest
		}

		response.Subspecs = &[]*communityReviewSubspec{}

		rows, err := db.Query(
			`SELECT spec_subspec.spec_id,
				spec_subspec.id, spec_subspec.subspec_name, spec_subspec.updated_at,
				(
					SELECT COUNT(*)
					FROM spec_community_comment cc
					WHERE cc.target_type = 'subspec'
						AND cc.target_id = spec_subspec.id
						AND NOT EXISTS(
							SELECT * FROM spec_community_read cr
							WHERE cr.user_id = $2
								AND cr.target_type = 'comment'
								AND cr.target_id = cc.id
						)
				) AS unread_comments,
				(
					SELECT COUNT(*)
					FROM spec_community_comment cc
					WHERE cc.target_type = 'subspec'
						AND cc.target_id = spec_subspec.id
				) AS total_comments,
				(
					SELECT COUNT(*)
					FROM spec_community_comment cc
					INNER JOIN spec_block
						ON cc.target_type = 'block'
						AND spec_block.id = cc.target_id
						AND spec_block.spec_id = spec_subspec.spec_id
						AND spec_block.subspec_id = spec_subspec.id
					WHERE cc.spec_id = spec_subspec.spec_id
						AND NOT EXISTS(
							SELECT * FROM spec_community_read cr
							WHERE cr.user_id = $2
								AND cr.target_type = 'comment'
								AND cr.target_id = cc.id
						)
				) AS block_unread_comments,
				(
					SELECT COUNT(*)
					FROM spec_community_comment cc
					INNER JOIN spec_block
						ON cc.target_type = 'block'
						AND spec_block.id = cc.target_id
						AND spec_block.spec_id = spec_subspec.spec_id
						AND spec_block.subspec_id = spec_subspec.id
					WHERE cc.spec_id = spec_subspec.spec_id
				) AS block_total_comments
			FROM spec_subspec
			INNER JOIN spec
				ON spec.id = spec_subspec.spec_id
			WHERE spec.owner_type = $1 AND spec.owner_id = $2
				AND spec.id = $3
			ORDER BY spec_subspec.subspec_name ASC`,
			OwnerTypeUser, userID, specID,
		)
		if err != nil {
			logError(r, &userID, fmt.Errorf("querying subspecs review: %w", err))
			return nil, http.StatusInternalServerError
		}

		for rows.Next() {
			var review = communityReviewSubspec{}
			err = rows.Scan(&review.SpecID, &review.ID, &review.Name, &review.Updated,
				&review.UnreadComments, &review.TotalComments,
				&review.BlockUnreadComments, &review.BlockTotalComments)
			if err != nil {
				logError(r, &userID, fmt.Errorf("scanning subspecs review row: %w", err))
				return nil, http.StatusInternalServerError
			}
			(*response.Subspecs) = append((*response.Subspecs), &review)
		}
	}

	if all || request == "comments" {
		// Load comments user has authored review

		updatedBefore, err := AtoTimeNilIfEmpty(r.FormValue("updatedBefore"))
		if err != nil {
			logError(r, &userID, fmt.Errorf("parsing updatedBefore: %w", err))
			return nil, http.StatusBadRequest
		}

		var unreadOnly = AtoBool(r.FormValue("unreadOnly"))

		response.Comments = &[]*communityReviewComment{}

		var args = []interface{}{userID, commentsReviewPageSize}

		var unreadOnlyCond string
		if unreadOnly {
			unreadOnlyCond = `AND EXISTS(
				SELECT *
				FROM spec_community_comment cc_sub
					WHERE cc_sub.target_type = 'comment'
						AND cc_sub.target_id = cc.id
						AND NOT EXISTS(
							SELECT * FROM spec_community_read cr
							WHERE cr.user_id = $1
								AND cr.target_type = 'comment'
								AND cr.target_id = cc_sub.id
						)
			)`
		}

		var updatedBeforeCond string
		if updatedBefore != nil {
			updatedBeforeCond = `AND cc.updated_at < ` + argPlaceholder(updatedBefore, &args)
		}

		rows, err := db.Query(
			`SELECT cc.spec_id, cc.id, cc.comment_body, cc.created_at, cc.updated_at,
				(
					SELECT COUNT(*)
					FROM spec_community_comment cc_sub
					WHERE cc_sub.target_type = 'comment'
						AND cc_sub.target_id = cc.id
						AND NOT EXISTS(
							SELECT * FROM spec_community_read cr
							WHERE cr.user_id = $1
								AND cr.target_type = 'comment'
								AND cr.target_id = cc_sub.id
						)
				) AS unread_comments,
				(
					SELECT COUNT(*)
					FROM spec_community_comment cc_sub
					WHERE cc_sub.target_type = 'comment'
						AND cc_sub.target_id = cc.id
				) AS total_comments
			FROM spec_community_comment cc
			WHERE cc.user_id = $1
				`+unreadOnlyCond+`
				`+updatedBeforeCond+`
			ORDER BY cc.updated_at DESC
			LIMIT $2`,
			args...,
		)
		if err != nil {
			logError(r, &userID, fmt.Errorf("querying comments review: %w", err))
			return nil, http.StatusInternalServerError
		}

		for rows.Next() {
			var review = communityReviewComment{}
			err = rows.Scan(&review.SpecID, &review.ID, &review.Body,
				&review.Created, &review.Updated,
				&review.UnreadComments, &review.TotalComments)
			if err != nil {
				logError(r, &userID, fmt.Errorf("scanning comments review row: %w", err))
				return nil, http.StatusInternalServerError
			}
			(*response.Comments) = append((*response.Comments), &review)
		}

		if len(*response.Comments) > 0 {
			// Check whether the query has more results following this page
			var lastResultUpdatedAt = (*response.Comments)[len(*response.Comments)-1].Updated
			err = db.QueryRow(
				`SELECT EXISTS(
					SELECT *
					FROM spec_community_comment AS cc
					WHERE cc.user_id = $1
						AND cc.updated_at < $2
						`+unreadOnlyCond+`
					LIMIT 1
				)`,
				userID, lastResultUpdatedAt,
			).Scan(&response.HasMoreComments)
			if err != nil {
				logError(r, &userID, fmt.Errorf("reading has more comments: %w", err))
				return nil, http.StatusInternalServerError
			}
		} else {
			var f = false
			response.HasMoreComments = &f
		}

		if updatedBefore == nil {
			// Count total number of comments on first page
			err = db.QueryRow(
				`SELECT COUNT(*)
				FROM spec_community_comment AS cc
				WHERE cc.user_id = $1
					`+unreadOnlyCond,
				userID,
			).Scan(&response.TotalComments)
			if err != nil {
				logError(r, &userID, fmt.Errorf("counting comments: %w", err))
				return nil, http.StatusInternalServerError
			}
		}
	}

	return response, http.StatusOK
}
