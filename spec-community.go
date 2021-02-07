package main

import (
	"fmt"
	"net/http"
	"time"
)

// Comment represents a comment submitted on a spec, subspec, or block.
// specId, targetType, and targetId are omitted as they are known by the caller.
type Comment struct {
	ID      int64     `json:"id"`
	UserID  uint      `json:"userId"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	Body    string    `json:"body"`

	Username  string  `json:"username"`
	Highlight *string `json:"highlight"`

	// Community attributes
	UserRead      bool `json:"userRead"`
	UnreadCount   uint `json:"unreadCount"`
	CommentsCount uint `json:"commentsCount"`
}

// Tag represents a tag associated with an item.
// specId is omitted as it is known by the caller.
type Tag struct {
	TargetType   int64   `json:"targetType"`
	TargetID     int64   `json:"targetId"`
	TagID        int64   `json:"tagId"`
	AssentVotes  uint    `json:"assentVotes"`
	DissentVotes uint    `json:"dissentVotes"`
	AdminPin     *string `json:"adminPin"`

	// The current user's own vote.
	UserVote *string `json:"userVote"`
}

const (
	// CommunityTargetSpec associates community features with a spec target.
	CommunityTargetSpec = "spec"
	// CommunityTargetSubspec associates community features with a subspec target.
	CommunityTargetSubspec = "subspec"
	// CommunityTargetBlock associates community features with a block target.
	CommunityTargetBlock = "block"
	// CommunityTargetComment associates community features with a comment target.
	CommunityTargetComment = "comment"
)

func loadCommentsPage(r *http.Request, db DBConn, userID uint,
	targetType string, targetID int64,
	pageSize uint, updatedBefore *time.Time, unreadOnly bool) ([]*Comment, bool, uint, uint, int) {

	var comments = []*Comment{}
	var unreadCount uint
	var commentsCount uint
	var hasMore bool

	var args = []interface{}{targetType, targetID, userID, pageSize}

	var unreadCountField = `(SELECT COUNT(*)
		FROM spec_community_comment AS subc
		LEFT JOIN spec_community_read AS subr
			ON subr.user_id = $3
			AND subr.target_type = 'comment'
			AND subr.target_id = subc.id
		WHERE subc.target_type = 'comment' AND subc.target_id = c.id
			AND subr.user_id IS NULL
		) AS unread_count`

	var unreadOnlyCond string
	var commentsCountField string

	if unreadOnly {
		// Limit to unread comments
		unreadOnlyCond = `AND NOT EXISTS(
			SELECT *
			FROM spec_community_read AS r
			WHERE r.user_id = $3 AND r.target_type = 'comment' AND r.target_id = c.id
			)`
		// Don't count comments when only vieweing unread
		commentsCountField = `0 AS comments_count`
	} else {
		// Count comments when viewing all
		commentsCountField = `(SELECT COUNT(*)
			FROM spec_community_comment AS subc
			WHERE subc.target_type = 'comment' AND subc.target_id = c.id
			) AS comments_count`
	}

	var userReadField = `(SELECT EXISTS(
		SELECT *
		FROM spec_community_read AS r
		WHERE r.user_id = $3 AND r.target_type = 'comment' AND r.target_id = c.id
		)) AS user_read`

	var unionUsersOwnComments string
	var updatedBeforeCond string

	if updatedBefore == nil {
		// Select all of the current user's own comments
		unionUsersOwnComments =
			`(SELECT c.id, c.user_id, c.created_at, c.updated_at, c.comment_body, u.username,
				'' AS highlight, -- blank because highlight for current user is already known
				` + unreadCountField + `,
				` + commentsCountField + `,
				` + userReadField + `
			FROM spec_community_comment AS c
			INNER JOIN user_account AS u
				ON u.id = c.user_id
			WHERE c.target_type = $1 AND c.target_id = $2
				AND c.user_id = $3
				` + unreadOnlyCond + `
			ORDER BY c.updated_at)
			UNION`
	} else {
		updatedBeforeCond = `AND c.updated_at < ` + argPlaceholder(updatedBefore, &args)
	}

	// Select pageSize community comments (preceeding updatedBefore if given)
	// Comment count is only returned when requesting first page;
	// afterward, only whether there are further pages is returned
	rows, err := db.Query(
		`SELECT * FROM (
			`+unionUsersOwnComments+`
			(SELECT c.id, c.user_id, c.created_at, c.updated_at, c.comment_body, u.username,
				u.user_settings::json#>>'{userProfile,highlightUsername}' AS highlight,
				`+unreadCountField+`,
				`+commentsCountField+`,
				`+userReadField+`
			FROM spec_community_comment AS c
			INNER JOIN user_account AS u
				ON u.id = c.user_id
			WHERE c.target_type = $1 AND c.target_id = $2
				AND c.user_id != $3
				`+updatedBeforeCond+`
				`+unreadOnlyCond+`
			ORDER BY c.updated_at DESC
			LIMIT $4)
		) AS c ORDER BY CASE WHEN c.user_id = $3 THEN 0 ELSE 1 END, c.updated_at DESC`,
		args...)
	if err != nil {
		logError(r, userID, fmt.Errorf("reading comments: %w", err))
		return nil, false, 0, 0, http.StatusInternalServerError
	}

	for rows.Next() {
		c := Comment{}
		err = rows.Scan(&c.ID, &c.UserID, &c.Created, &c.Updated, &c.Body,
			&c.Username, &c.Highlight, &c.UnreadCount, &c.CommentsCount, &c.UserRead)
		if err != nil {
			if err2 := rows.Close(); err2 != nil {
				logError(r, userID, fmt.Errorf("closing rows: %s; on scan error: %w", err2, err))
				return nil, false, 0, 0, http.StatusInternalServerError
			}
			logError(r, userID, fmt.Errorf("scanning spec: %w", err))
			return nil, false, 0, 0, http.StatusInternalServerError
		}
		comments = append(comments, &c)
	}

	if updatedBefore == nil {
		// Count unread and total comments when loading initial page
		err = db.QueryRow(
			`SELECT COUNT(c.id)
			FROM spec_community_comment AS c
			LEFT JOIN spec_community_read AS r
				ON r.user_id = $3 AND r.target_type = 'comment' AND r.target_id = c.id
			WHERE c.target_type = $1 AND c.target_id = $2 AND r.user_id IS NULL`,
			targetType, targetID, userID).Scan(&unreadCount)
		if err != nil {
			logError(r, userID, fmt.Errorf("reading unread count: %w", err))
			return nil, false, 0, 0, http.StatusInternalServerError
		}
		err = db.QueryRow(
			`SELECT COUNT(c.id)
			FROM spec_community_comment AS c
			WHERE c.target_type = $1 AND c.target_id = $2`,
			targetType, targetID).Scan(&commentsCount)
		if err != nil {
			logError(r, userID, fmt.Errorf("reading comment count: %w", err))
			return nil, false, 0, 0, http.StatusInternalServerError
		}
		hasMore = commentsCount > pageSize
	} else if len(comments) > 0 {
		// Check whether the query has more results following this page
		var lastResultUpdatedAt = comments[len(comments)-1].Updated
		err = db.QueryRow(
			`SELECT EXISTS(
				SELECT *
				FROM spec_community_comment AS c
				WHERE c.target_type = $1 AND c.target_id = $2
					AND c.user_id != $3
					AND c.updated_at < $4
					`+unreadOnlyCond+`
				LIMIT 1
			)`,
			targetType, targetID, userID, lastResultUpdatedAt,
		).Scan(&hasMore)
		if err != nil {
			logError(r, userID, fmt.Errorf("reading has more comments: %w", err))
			return nil, false, 0, 0, http.StatusInternalServerError
		}
	}

	return comments, hasMore, unreadCount, commentsCount, http.StatusOK
}
