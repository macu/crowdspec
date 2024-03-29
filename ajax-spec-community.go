package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const commentsPageSize = 5

// TODO pass owner info so community modal can determine readability of private content

type specCommunity struct {
	// context owner
	OwnerType string `json:"ownerType"`
	OwnerID   uint   `json:"ownerId"`

	Spec struct {
		ID      int64     `json:"id"`
		Created time.Time `json:"created"`
		Updated time.Time `json:"updated"`
		Public  bool      `json:"public"`
		Name    string    `json:"name"`
		Desc    *string   `json:"desc"`
	} `json:"spec"`

	Tags          []*Tag     `json:"tags,omitempty"`
	Comments      []*Comment `json:"comments"`
	UnreadCount   uint       `json:"unreadCount"`
	CommentsCount uint       `json:"commentsCount"`

	RenderTime time.Time `json:"renderTime"`
}

type subspecCommunity struct {
	// context owner
	OwnerType  string `json:"ownerType"`
	OwnerID    uint   `json:"ownerId"`
	SpecPublic bool   `json:"specPublic"`

	Subspec struct {
		ID      int64     `json:"id"`
		Created time.Time `json:"created"`
		Updated time.Time `json:"updated"`
		Private bool      `json:"private"`
		Name    string    `json:"name"`
		Desc    *string   `json:"desc"`
	} `json:"subspec"`

	Tags          []*Tag     `json:"tags,omitempty"`
	Comments      []*Comment `json:"comments"`
	UnreadCount   uint       `json:"unreadCount"`
	CommentsCount uint       `json:"commentsCount"`

	RenderTime time.Time `json:"renderTime"`
}

type blockCommunity struct {
	// context owner
	OwnerType      string `json:"ownerType"`
	OwnerID        uint   `json:"ownerId"`
	SpecPublic     bool   `json:"specPublic"`
	SubspecPrivate *bool  `json:"subspecPrivate"`

	Block struct {
		ID          int64     `json:"id"`
		Created     time.Time `json:"created"`
		Updated     time.Time `json:"updated"`
		StyleType   string    `json:"styleType"`
		Number      *uint     `json:"number"` // calculated in query in style type is numbered
		ContentType *string   `json:"contentType"`
		RefType     *string   `json:"refType"`
		RefID       *int64    `json:"refId"`
		Title       *string   `json:"title"`
		Body        *string   `json:"body"`
		HTML        *string   `json:"html"`

		RefItem interface{} `json:"refItem,omitempty"`
	} `json:"block"`

	Tags          []*Tag     `json:"tags,omitempty"`
	Comments      []*Comment `json:"comments"`
	UnreadCount   uint       `json:"unreadCount"`
	CommentsCount uint       `json:"commentsCount"`

	RenderTime time.Time `json:"renderTime"`
}

type commentCommunity struct {
	// context owner
	OwnerType      string `json:"ownerType"`
	OwnerID        uint   `json:"ownerId"`
	SpecPublic     bool   `json:"specPublic"`
	SubspecPrivate *bool  `json:"subspecPrivate"`

	Comment struct {
		ID         int64     `json:"id"`
		Created    time.Time `json:"created"`
		Updated    time.Time `json:"updated"`
		UserID     uint      `json:"userId"`
		TargetType string    `json:"targetType"`
		TargetID   int64     `json:"targetId"`
		Body       *string   `json:"body"`

		Username  string  `json:"username"`  // joined
		Highlight *string `json:"highlight"` // joined
		UserRead  bool    `json:"userRead"`  // joined
	} `json:"comment"`

	Tags          []*Tag     `json:"tags,omitempty"`
	Comments      []*Comment `json:"comments"`
	UnreadCount   uint       `json:"unreadCount"`
	CommentsCount uint       `json:"commentsCount"`

	// Context stack
	Stack []*communityStackElement `json:"stack"`

	RenderTime time.Time `json:"renderTime"`
}

// mirror structure of payload delivered to community-context-stack.vue
type communityStackElement struct {
	ElementType string `json:"type"`

	Element struct { // comment, block, spec, or subspec
		ID             int64    `json:"id"`
		Name           *string  `json:"name"` // spec or subspec
		SpecPublic     *bool    `json:"public"`
		SubspecPrivate *bool    `json:"private"`
		BlockRefType   *string  `json:"refType"` // block
		BlockTitle     *string  `json:"title"`   // block
		Body           *string  `json:"body"`    // block or comment
		BlockRefItem   struct { // ref item of block elements
			// ref subspec
			SubspecName    *string `json:"name"`
			SubspecPrivate *bool   `json:"private"`
			// ref url
			URLTitle *string `json:"title"`
			URL      *string `json:"url"`
		} `json:"refItem"`
		CommentUserID            *uint   `json:"userId"`
		CommentUsername          *string `json:"username"`
		CommentUsernameHighlight *string `json:"highlight"`
	} `json:"element"`

	// Include for null value client-side
	OnAdjustUnread   *struct{} `json:"onAdjustUnread"`
	OnAdjustComments *struct{} `json:"onAdjustComments"`
}

type communityCommentsPage struct {
	Comments      []*Comment `json:"comments"`
	HasMore       bool       `json:"hasMore"`
	UnreadCount   uint       `json:"unreadCount"`
	CommentsCount uint       `json:"commentsCount"`

	RenderTime time.Time `json:"renderTime"`
}

func extractValidCommunityTarget(r *http.Request, userID *uint) (int64, string, int) {
	targetID, err := AtoInt64(r.FormValue("targetId"))
	if err != nil {
		logError(r, userID, err)
		return 0, "", http.StatusBadRequest
	}

	targetType := r.FormValue("targetType")
	switch targetType {
	case "spec", "subspec", "block", "comment":
		return targetID, targetType, http.StatusOK
	default:
		logError(r, userID, fmt.Errorf("unrecognized targetType: %s", targetType))
		return 0, "", http.StatusBadRequest
	}
}

func ajaxSpecLoadCommunity(db *sql.DB, userID *uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// GET

	_, err := AtoInt64(r.FormValue("specId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("invalid specId: %w", err))
		return nil, http.StatusBadRequest
	}

	// validates the target type
	targetID, targetType, status := extractValidCommunityTarget(r, userID)
	if status != http.StatusOK {
		return nil, status
	}

	if access, status := verifyReadTarget(r, db, userID, targetType, targetID); !access {
		return nil, status
	}

	unreadOnly := AtoBool(r.FormValue("unreadOnly")) // only load unread comments

	// load tags
	// TODO

	comments, _, unreadCount, commentsCount, status := loadCommentsPage(r, db, userID,
		targetType, targetID, commentsPageSize, nil, unreadOnly)
	if status != http.StatusOK {
		return nil, status
	}

	switch targetType {

	case CommunityTargetSpec:
		sc := &specCommunity{
			Comments:      comments,
			UnreadCount:   unreadCount,
			CommentsCount: commentsCount,
			RenderTime:    time.Now(),
		}

		err = db.QueryRow(`
			SELECT owner_type, owner_id, is_public,
				id, created_at, spec_name, spec_desc,
				GREATEST(updated_at, blocks_updated_at) AS last_updated
			FROM spec
			WHERE id=$1`,
			targetID,
		).Scan(&sc.OwnerType, &sc.OwnerID, &sc.Spec.Public,
			&sc.Spec.ID, &sc.Spec.Created, &sc.Spec.Name, &sc.Spec.Desc,
			&sc.Spec.Updated)
		if err != nil {
			logError(r, userID, fmt.Errorf("reading spec: %w", err))
			return nil, http.StatusInternalServerError
		}

		// TODO load policy

		return sc, http.StatusOK

	case CommunityTargetSubspec:
		sc := &subspecCommunity{
			Comments:      comments,
			UnreadCount:   unreadCount,
			CommentsCount: commentsCount,
			RenderTime:    time.Now(),
		}

		err = db.QueryRow(`
			SELECT spec.owner_type, spec.owner_id, spec.is_public, spec_subspec.is_private,
			 	spec_subspec.id, spec_subspec.created_at, subspec_name, subspec_desc,
				GREATEST(spec_subspec.updated_at, spec_subspec.blocks_updated_at) AS last_updated
			FROM spec_subspec
			INNER JOIN spec ON spec.id = spec_subspec.spec_id
			WHERE spec_subspec.id=$1`,
			targetID,
		).Scan(&sc.OwnerType, &sc.OwnerID, &sc.SpecPublic, &sc.Subspec.Private,
			&sc.Subspec.ID, &sc.Subspec.Created, &sc.Subspec.Name, &sc.Subspec.Desc,
			&sc.Subspec.Updated)
		if err != nil {
			logError(r, userID, fmt.Errorf("reading subspec: %w", err))
			return nil, http.StatusInternalServerError
		}

		// TODO load policy

		return sc, http.StatusOK

	case CommunityTargetBlock:
		bc := &blockCommunity{
			/*
				Block struct {
					ID          int64   `json:"id"`
					StyleType   string  `json:"styleType"`
					Number      *uint   `json:"number"` // calculated in query in style type is numbered
					ContentType *string `json:"contentType"`
					RefType     *string `json:"refType"`
					RefID       *int64  `json:"refId"`
					Title       *string `json:"title"`
					Body        *string `json:"body"`

					RefItem interface{} `json:"refItem,omitempty"`
				}
			*/

			Comments:      comments,
			UnreadCount:   unreadCount,
			CommentsCount: commentsCount,
			RenderTime:    time.Now(),
		}

		err = db.QueryRow(`SELECT spec.owner_type, spec.owner_id,
				spec.is_public, spec_subspec.is_private,
				spec_block.id, spec_block.created_at, spec_block.updated_at,
				style_type, content_type, ref_type, ref_id, block_title,
				CASE WHEN content_type = 'plaintext' THEN block_body ELSE NULL END AS block_body,
				CASE WHEN content_type = 'markdown' THEN rendered_html ELSE NULL END AS rendered_html,
				(CASE WHEN style_type = 'numbered' THEN
					-- find position among numbered blocks in same immediate list
					array_position(ARRAY(
						SELECT sibs.id
						FROM spec_block AS sibs
						WHERE sibs.spec_id = spec_block.spec_id
							AND (
								(sibs.subspec_id IS NULL AND spec_block.subspec_id IS NULL)
								OR sibs.subspec_id = spec_block.subspec_id
							)
							AND (
								(sibs.parent_id IS NULL AND spec_block.parent_id IS NULL)
								OR sibs.parent_id = spec_block.parent_id
							)
							AND sibs.style_type = 'numbered'
						ORDER BY sibs.order_number
					), spec_block.id)
					END) AS number
			FROM spec_block
			INNER JOIN spec ON spec.id = spec_block.spec_id
			LEFT JOIN spec_subspec ON spec_subspec.id = spec_block.subspec_id
			WHERE spec_block.id=$1`,
			targetID,
		).Scan(&bc.OwnerType, &bc.OwnerID, &bc.SpecPublic, &bc.SubspecPrivate,
			&bc.Block.ID, &bc.Block.Created, &bc.Block.Updated,
			&bc.Block.StyleType, &bc.Block.ContentType, &bc.Block.RefType, &bc.Block.RefID,
			&bc.Block.Title, &bc.Block.Body, &bc.Block.HTML, &bc.Block.Number)
		if err != nil {
			logError(r, userID, fmt.Errorf("reading block: %w", err))
			return nil, http.StatusInternalServerError
		}

		// load refItem
		if bc.Block.RefType != nil && bc.Block.RefID != nil {
			bc.Block.RefItem, err = loadRefItem(db, *bc.Block.RefType, *bc.Block.RefID)
			if err != nil {
				logError(r, userID, fmt.Errorf("reading ref item: %w", err))
				return nil, http.StatusInternalServerError
			}
		}

		// TODO load policy

		return bc, http.StatusOK

	case CommunityTargetComment:
		cc := &commentCommunity{
			/*
				Comment struct {
					ID         int64     `json:"id"`
					Created    time.Time `json:"created"`
					Updated    time.Time `json:"updated"`
					UserID     int64     `json:"userId"`
					TargetType string    `json:"targetType"`
					TargetID   int64     `json:"targetId"`
					Body       *string   `json:"body"`

					Username  int64 `json:"username"`  // joined
					Highlight int64 `json:"highlight"` // joined
				}
			*/

			Comments:      comments,
			UnreadCount:   unreadCount,
			CommentsCount: commentsCount,
			RenderTime:    time.Now(),
		}

		var commentArgs []interface{}

		var userReadField string
		if userID == nil {
			userReadField = `FALSE AS user_read`
		} else {
			userReadField = `(SELECT EXISTS(
					SELECT r.user_id FROM spec_community_read AS r
					WHERE r.user_id = ` + argPlaceholder(*userID, &commentArgs) + `
					AND r.target_type = ` + argPlaceholder(CommunityTargetComment, &commentArgs) + `
					AND r.target_id = ` + argPlaceholder(targetID, &commentArgs) + `
				)) AS user_read`
		}

		err = db.QueryRow(`
			SELECT spec.owner_type, spec.owner_id, spec.is_public,
				c.id, c.created_at, c.updated_at, c.user_id, u.username,
				u.user_settings::json#>>'{userProfile,highlightUsername}' AS highlight,
				c.target_type, c.target_id, c.comment_body,
				`+userReadField+`
			FROM spec_community_comment AS c
			INNER JOIN spec ON spec.id = c.spec_id
			INNER JOIN user_account AS u
				ON u.id = c.user_id
			WHERE c.id = `+argPlaceholder(targetID, &commentArgs),
			commentArgs...,
		).Scan(&cc.OwnerType, &cc.OwnerID, &cc.SpecPublic,
			&cc.Comment.ID, &cc.Comment.Created, &cc.Comment.Updated,
			&cc.Comment.UserID, &cc.Comment.Username, &cc.Comment.Highlight,
			&cc.Comment.TargetType, &cc.Comment.TargetID, &cc.Comment.Body,
			&cc.Comment.UserRead)
		if err != nil {
			logError(r, userID, fmt.Errorf("reading comment: %w", err))
			return nil, http.StatusInternalServerError
		}

		// TODO load policy

		if AtoBool(r.FormValue("loadStack")) {
			// stack of comments up to a spec item
			// is requested only for comment communities

			var targetType = cc.Comment.TargetType
			var targetID = cc.Comment.TargetID

			if targetType == CommunityTargetComment {
				var rows, err = db.Query(
					`WITH RECURSIVE comment_stack(id, target_type, target_id, level) AS (
						-- Anchor
						SELECT id, target_type, target_id, 0
						FROM spec_community_comment
						WHERE id = $1
						-- Recursive Member
						UNION ALL
						SELECT cc.id, cc.target_type, cc.target_id, cs.level + 1
						FROM spec_community_comment cc, comment_stack cs
						WHERE cs.target_type = $2
							AND cs.target_id = cc.id
					)
					SELECT cc.id, cc.comment_body, cc.user_id, u.username,
						u.user_settings::json#>>'{userProfile,highlightUsername}' AS highlight,
						cc.target_type, cc.target_id
					FROM spec_community_comment cc
					INNER JOIN user_account AS u
						ON u.id = cc.user_id
					INNER JOIN comment_stack cs
					ON cs.id = cc.id
					ORDER BY cs.level ASC`,
					targetID, CommunityTargetComment,
				)
				if err != nil {
					logError(r, userID, fmt.Errorf("querying stack: %w", err))
					return nil, http.StatusInternalServerError
				}

				for rows.Next() {
					var context = communityStackElement{ElementType: CommunityTargetComment}
					var nextTargetType string
					var nextTargetID int64
					err = rows.Scan(&context.Element.ID, &context.Element.Body,
						&context.Element.CommentUserID, &context.Element.CommentUsername,
						&context.Element.CommentUsernameHighlight,
						&nextTargetType, &nextTargetID)
					if err != nil {
						if err2 := rows.Close(); err2 != nil {
							logError(r, userID, fmt.Errorf("closing rows: %s; on scanning stack: %w", err2, err))
							return nil, http.StatusInternalServerError
						}
						logError(r, userID, fmt.Errorf("scanning stack row: %w", err))
						return nil, http.StatusInternalServerError
					}
					cc.Stack = append(cc.Stack, &context)
					targetType = nextTargetType
					targetID = nextTargetID
				}
			}

			if targetType == CommunityTargetSpec {

				var context = communityStackElement{ElementType: CommunityTargetSpec}
				err = db.QueryRow(
					`SELECT id, spec_name, is_public
					FROM spec WHERE id = $1`, targetID,
				).Scan(&context.Element.ID, &context.Element.Name, &context.Element.SpecPublic)
				if err != nil {
					logError(r, userID, fmt.Errorf("scanning stack spec: %w", err))
					return nil, http.StatusInternalServerError
				}
				cc.Stack = append(cc.Stack, &context)

				// SpecPubic is already known

			} else if targetType == CommunityTargetSubspec {

				var context = communityStackElement{ElementType: CommunityTargetSubspec}
				err = db.QueryRow(
					`SELECT spec_subspec.id, subspec_name, is_private
					FROM spec_subspec WHERE id = $1`, targetID,
				).Scan(&context.Element.ID, &context.Element.Name, &context.Element.SubspecPrivate)
				if err != nil {
					logError(r, userID, fmt.Errorf("scanning stack subspec: %w", err))
					return nil, http.StatusInternalServerError
				}
				cc.Stack = append(cc.Stack, &context)

				// SpecPubic is already known
				// Copy value of SubspecPrivate to outer context
				cc.SubspecPrivate = context.Element.SubspecPrivate

			} else if targetType == CommunityTargetBlock {

				// SpecPubic is already known
				var context = communityStackElement{ElementType: CommunityTargetBlock}
				err = db.QueryRow(
					`SELECT spec_subspec.is_private,
						b.id, b.ref_type,
						-- only take first 100 characters for single-line stack
						substr(b.block_title, 0, 100) AS block_title,
						substr(b.block_body, 0, 100) AS block_body,
						COALESCE(ref_subspec.subspec_name, '') AS subspec_name,
						COALESCE(ref_subspec.is_private, FALSE) AS subspec_private,
						ref_url.url_title AS url_title,
						COALESCE(ref_url.url, '') AS url
					FROM spec_block b
					LEFT JOIN spec_subspec ON spec_subspec.id = b.subspec_id
					LEFT JOIN spec_subspec AS ref_subspec
						ON b.ref_type = $2
						AND ref_subspec.id = b.ref_id
						AND ref_subspec.spec_id = b.spec_id
					LEFT JOIN spec_url AS ref_url
						ON b.ref_type = $3
						AND ref_url.id = b.ref_id
						AND ref_url.spec_id = b.spec_id
					WHERE b.id = $1`,
					targetID, BlockRefSubspec, BlockRefURL,
				).Scan(&cc.SubspecPrivate, // Set value of SubspecPrivate on outer context
					&context.Element.ID, &context.Element.BlockRefType,
					&context.Element.BlockTitle, &context.Element.Body,
					&context.Element.BlockRefItem.SubspecName,
					&context.Element.BlockRefItem.SubspecPrivate,
					&context.Element.BlockRefItem.URLTitle, &context.Element.BlockRefItem.URL)
				if err != nil {
					logError(r, userID, fmt.Errorf("scanning stack subspec: %w", err))
					return nil, http.StatusInternalServerError
				}
				cc.Stack = append(cc.Stack, &context)

			}

			// Reverse order of stack
			for i, j := 0, len(cc.Stack)-1; i < j; i, j = i+1, j-1 {
				cc.Stack[i], cc.Stack[j] = cc.Stack[j], cc.Stack[i]
			}
		}

		return cc, http.StatusOK

	default:
		return nil, http.StatusInternalServerError
	}
}

func ajaxSpecCommunityLoadCommentsPage(db *sql.DB, userID *uint,
	w http.ResponseWriter, r *http.Request) (interface{}, int) {

	_, err := AtoInt64(r.FormValue("specId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("invalid specId: %w", err))
		return nil, http.StatusBadRequest
	}

	// validates the target type
	targetID, targetType, status := extractValidCommunityTarget(r, userID)
	if status != http.StatusOK {
		return nil, status
	}

	if access, status := verifyReadTarget(r, db, userID, targetType, targetID); !access {
		return nil, status
	}

	updatedBefore, err := AtoTimeNilIfEmpty(r.FormValue("updatedBefore"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing updatedBefore: %w", err))
		return nil, http.StatusBadRequest
	}

	unreadOnly := AtoBool(r.FormValue("unreadOnly")) // only load unread comments

	comments, hasMore, unreadCount, commentsCount, status := loadCommentsPage(r, db, userID,
		targetType, targetID, commentsPageSize, updatedBefore, unreadOnly)
	if status != http.StatusOK {
		return nil, status
	}

	results := &communityCommentsPage{
		Comments:      comments,
		HasMore:       hasMore,
		UnreadCount:   unreadCount,
		CommentsCount: commentsCount,
		RenderTime:    time.Now(),
	}

	return results, http.StatusOK
}

func ajaxSpecCommunityAddComment(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	specID, err := AtoInt64(r.FormValue("specId"))
	if err != nil {
		logError(r, &userID, fmt.Errorf("invalid specId: %w", err))
		return nil, http.StatusBadRequest
	}

	targetID, targetType, status := extractValidCommunityTarget(r, &userID)
	if status != http.StatusOK {
		return nil, status
	}

	// allow adding comment under any readable target
	if access, status := verifyReadTarget(r, db, &userID, targetType, targetID); !access {
		return nil, status
	}

	body := strings.TrimSpace(r.FormValue("body"))
	if body == "" {
		logError(r, &userID, fmt.Errorf("empty comment body"))
		return nil, http.StatusBadRequest
	}

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		var now = time.Now()
		var c = Comment{
			UserID:   userID,
			Created:  now,
			Updated:  now,
			UserRead: true,
		}

		err := tx.QueryRow(
			`INSERT INTO spec_community_comment (
				spec_id, target_type, target_id, user_id, created_at, updated_at, comment_body
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7
			) RETURNING id, comment_body, (SELECT username FROM user_account WHERE id = $4)
			`, specID, targetType, targetID, userID, now, now, body).Scan(&c.ID, &c.Body, &c.Username)
		if err != nil {
			logError(r, &userID, fmt.Errorf("adding comment: %w", err))
			return nil, http.StatusInternalServerError
		}

		// Mark new comment read by author
		_, err = tx.Exec(
			`INSERT INTO spec_community_read (
				user_id, target_type, target_id, updated_at
			) VALUES (
				$1, $2, $3, $4
			)`, userID, CommunityTargetComment, c.ID, time.Now(),
		)
		if err != nil {
			logError(r, &userID, fmt.Errorf("marking new comment read: %w", err))
			return nil, http.StatusInternalServerError
		}

		// TODO Add initial comment tags

		return c, http.StatusOK
	})
}

func ajaxSpecCommunityUpdateComment(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST
	// TODO Retain history of edits

	_, err := AtoInt64(r.FormValue("specId"))
	if err != nil {
		logError(r, &userID, fmt.Errorf("invalid specId: %w", err))
		return nil, http.StatusBadRequest
	}

	commentID, err := AtoInt64(r.FormValue("commentId"))
	if err != nil {
		logError(r, &userID, fmt.Errorf("invalid commentId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyWriteTarget(r, db, userID, CommunityTargetComment, commentID); !access {
		return nil, status
	}

	body := strings.TrimSpace(r.FormValue("body"))
	if body == "" {
		logError(r, &userID, fmt.Errorf("empty comment body"))
		return nil, http.StatusBadRequest
	}

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		var comment = struct {
			ID      int64     `json:"id"`
			UserID  uint      `json:"userId"`
			Updated time.Time `json:"updated"`
			Body    string    `json:"body"`
		}{
			ID:      commentID,
			Updated: time.Now(),
		}

		err := tx.QueryRow(
			`UPDATE spec_community_comment
			SET comment_body = $2, updated_at = $3
			WHERE id = $1
			RETURNING user_id, comment_body
			`, commentID, body, comment.Updated).Scan(&comment.UserID, &comment.Body)
		if err != nil {
			logError(r, &userID, fmt.Errorf("updating comment: %w", err))
			return nil, http.StatusInternalServerError
		}

		return comment, http.StatusOK
	})
}

func ajaxSpecCommunityDeleteComment(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	_, err := AtoInt64(r.FormValue("specId"))
	if err != nil {
		logError(r, &userID, fmt.Errorf("invalid specId: %w", err))
		return nil, http.StatusBadRequest
	}

	commentID, err := AtoInt64(r.FormValue("commentId"))
	if err != nil {
		logError(r, &userID, fmt.Errorf("invalid commentId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyDeleteTarget(r, db, userID, CommunityTargetComment, commentID); !access {
		return nil, status
	}

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		_, err := tx.Exec(`DELETE FROM spec_community_comment WHERE id = $1`, commentID)
		if err != nil {
			logError(r, &userID, fmt.Errorf("deleting comment: %w", err))
			return nil, http.StatusInternalServerError
		}

		return nil, http.StatusOK
	})
}

func ajaxSpecCommunityMarkRead(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	targetID, targetType, status := extractValidCommunityTarget(r, &userID)
	if status != http.StatusOK {
		return nil, status
	}

	read := AtoBool(r.FormValue("read"))

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		var existing bool

		err := tx.QueryRow(
			`SELECT EXISTS(
				SELECT * FROM spec_community_read
				WHERE user_id = $1
					AND target_type = $2 AND target_id = $3
			)`, userID, targetType, targetID,
		).Scan(&existing)
		if err != nil {
			logError(r, &userID, fmt.Errorf("checking whether user read: %w", err))
			return nil, http.StatusInternalServerError
		}

		if existing {
			if read {
				// Do nothing - read record already present
			} else {
				// Delete read record
				_, err = db.Exec(
					`DELETE FROM spec_community_read
						WHERE user_id = $1
							AND target_type = $2 AND target_id = $3`,
					userID, targetType, targetID,
				)
				if err != nil {
					logError(r, &userID, fmt.Errorf("marking unread: %w", err))
					return nil, http.StatusInternalServerError
				}
			}
		} else if read {
			// Create read record
			_, err = db.Exec(
				`INSERT INTO spec_community_read (
					user_id, target_type, target_id, updated_at
				) VALUES ($1, $2, $3, $4)`,
				userID, targetType, targetID, time.Now(),
			)
			if err != nil {
				logError(r, &userID, fmt.Errorf("marking read: %w", err))
				return nil, http.StatusInternalServerError
			}
		}

		return nil, http.StatusOK
	})
}
