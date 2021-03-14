package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Creates a block within a spec.
func ajaxSpecCreateBlock(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
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

	subspecID, err := AtoInt64NilIfEmpty(r.Form.Get("subspecId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing subspecId: %w", err))
		return nil, http.StatusBadRequest
	}

	parentID, err := AtoInt64NilIfEmpty(r.Form.Get("parentId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing parentId: %w", err))
		return nil, http.StatusBadRequest
	}

	insertBeforeID, err := AtoInt64NilIfEmpty(r.Form.Get("insertBeforeId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing insertBeforeId: %w", err))
		return nil, http.StatusBadRequest
	}

	var verifyBlockIds = []int64{}
	if parentID != nil {
		verifyBlockIds = append(verifyBlockIds, *parentID)
	}
	if insertBeforeID != nil {
		verifyBlockIds = append(verifyBlockIds, *insertBeforeID)
	}
	if access, status := verifyWriteSpecSubspecBlocks(r, db, userID, specID, subspecID,
		verifyBlockIds...); !access {
		return nil, status
	}

	styleType := r.Form.Get("styleType")
	if !isValidListStyleType(styleType) {
		logError(r, userID, fmt.Errorf("invalid styleType: %s", styleType))
		return nil, http.StatusBadRequest
	}

	contentType := r.Form.Get("contentType")
	if !isValidTextContentType(contentType) {
		logError(r, userID, fmt.Errorf("invalid contentType: %s", contentType))
		return nil, http.StatusBadRequest
	}

	refType, refID, err := validateCreateRefItemFields(r.Form)
	if err != nil {
		logError(r, userID, fmt.Errorf("invalid ref fields: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyRefAccess(r, db, userID, specID, refType, refID); !access {
		return nil, status
	}

	title := AtoPointerNilIfEmpty(Substr(strings.TrimSpace(r.Form.Get("title")), blockTitleMaxLen))

	body := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("body")))

	if refType == nil && title == nil && body == nil {
		logError(r, userID, fmt.Errorf("empty blocks are not currently allowed"))
		return nil, http.StatusBadRequest
	}

	var renderedHTML *string
	if body != nil && contentType == TextContentMarkdown {
		html, err := renderMarkdown(*body)
		if err != nil {
			// don't log error
			return nil, http.StatusBadRequest
		}
		renderedHTML = &html
	}

	return handleInTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		var err error

		// Create or load ref item
		var refItem interface{}
		if refType != nil {
			if refID == nil {
				refID, refItem, err = handleCreateRefItem(tx, specID, r.Form)
				if err != nil {
					logError(r, userID, fmt.Errorf("creating ref item: %w", err))
					return nil, http.StatusInternalServerError
				}
			} else {
				refItem, err = loadRefItem(tx, *refType, *refID)
				if err != nil {
					logError(r, userID, fmt.Errorf("loading ref item: %w", err))
					return nil, http.StatusInternalServerError
				}
			}
		}

		// Prepare insert position
		insertAt, code, err := makeInsertAt(tx, specID, subspecID, parentID, insertBeforeID, 1)
		if err != nil {
			logError(r, userID, fmt.Errorf("making insert position: %w", err))
			return nil, code
		}

		block := &SpecBlock{
			SpecID:      specID,
			SubspecID:   subspecID,
			ParentID:    parentID,
			OrderNumber: insertAt,
			StyleType:   styleType,
			ContentType: contentType,
			RefType:     refType,
			RefID:       refID,
			RefItem:     refItem,
		}

		// Create block row
		err = tx.QueryRow(`
			INSERT INTO spec_block
			(spec_id, created_at, updated_at, subspec_id, parent_id, order_number,
				style_type, content_type, ref_type, ref_id, block_title, block_body, rendered_html)
			VALUES ($1, $2, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id, created_at, updated_at, block_title, block_body, rendered_html
			`, specID, time.Now(), subspecID, parentID, insertAt,
			styleType, contentType, refType, refID, title, body, renderedHTML,
		).Scan(&block.ID, &block.Created, &block.Updated, &block.Title, &block.Body, &block.HTML)
		if err != nil {
			logError(r, userID, fmt.Errorf("creating block: %w", err))
			return nil, http.StatusInternalServerError
		}

		if status := recordSpecBlocksUpdated(tx, r, userID, specID); status != http.StatusOK {
			return nil, status
		}
		if subspecID != nil {
			if status := recordSubspecBlocksUpdated(db, r, userID, *subspecID); status != http.StatusOK {
				return nil, status
			}
		}

		return block, http.StatusOK
	})
}

func ajaxLoadBlockForEditing(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// GET

	specID, err := AtoInt64(r.FormValue("specId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing specId: %w", err))
		return nil, http.StatusBadRequest
	}

	blockID, err := AtoInt64(r.FormValue("blockId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing blockId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyWriteSpecBlock(r, db, userID, specID, blockID); !access {
		return nil, status
	}

	block, err := loadBlockForEditing(db, userID, specID, blockID)
	if err != nil {
		logError(r, userID, fmt.Errorf("loading block: %w", err))
		return nil, http.StatusInternalServerError
	}

	return block, http.StatusOK
}

func ajaxRenderMarkdown(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	// trim now as blackfriday trims the output anyway
	var markdown = strings.TrimSpace(r.FormValue("markdown"))

	if markdown == "" {
		return struct {
			HTML string `json:"html"`
		}{""}, http.StatusOK
	}

	// renderMarkdown will report errors in XML with line numbers,
	// which won't be accurate as blackfriday drops leading and trailing whitespace
	// and produces other transformations from source lines to html
	html, err := renderMarkdown(markdown)
	if err != nil {
		// don't log error;
		// return validation error to client
		return struct {
			Error string `json:"error"`
		}{err.Error()}, http.StatusOK
	}
	return struct {
		HTML string `json:"html"`
	}{html}, http.StatusOK
}

func ajaxSpecSaveBlock(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
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

	blockID, err := AtoInt64(r.Form.Get("blockId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing blockId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyWriteSpecBlock(r, db, userID, specID, blockID); !access {
		return nil, status
	}

	styleType := r.Form.Get("styleType")
	if !isValidListStyleType(styleType) {
		logError(r, userID, fmt.Errorf("invalid styleType: %s", styleType))
		return nil, http.StatusBadRequest
	}

	contentType := r.Form.Get("contentType")
	if !isValidTextContentType(contentType) {
		logError(r, userID, fmt.Errorf("invalid contentType: %s", contentType))
		return nil, http.StatusBadRequest
	}

	refType, refID, err := validateCreateRefItemFields(r.Form)
	if err != nil {
		logError(r, userID, fmt.Errorf("invalid ref fields: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyRefAccess(r, db, userID, specID, refType, refID); !access {
		return nil, status
	}

	title := AtoPointerNilIfEmpty(Substr(strings.TrimSpace(r.Form.Get("title")), blockTitleMaxLen))

	body := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("body")))

	if refType == nil && title == nil && body == nil {
		logError(r, userID, fmt.Errorf("empty blocks are not currently allowed"))
		return nil, http.StatusBadRequest
	}

	var renderedHTML *string
	if body != nil && contentType == TextContentMarkdown {
		html, err := renderMarkdown(*body)
		if err != nil {
			// don't log error
			return nil, http.StatusBadRequest
		}
		renderedHTML = &html
	}

	return handleInTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		var err error

		// Create or load ref item
		var refItem interface{}
		if refType != nil {
			if refID == nil {
				refID, refItem, err = handleCreateRefItem(tx, specID, r.Form)
				if err != nil {
					logError(r, userID, fmt.Errorf("creating ref item: %w", err))
					return nil, http.StatusInternalServerError
				}
			} else {
				refItem, err = loadRefItem(tx, *refType, *refID)
				if err != nil {
					logError(r, userID, fmt.Errorf("loading ref item: %w", err))
					return nil, http.StatusInternalServerError
				}
			}
		}

		block := &SpecBlock{
			ID:          blockID,
			StyleType:   styleType,
			ContentType: contentType,
			RefType:     refType,
			RefID:       refID,
			RefItem:     refItem,
		}

		// Update block row
		err = tx.QueryRow(`
			UPDATE spec_block
			SET updated_at=$4, style_type=$5, content_type=$6,
				ref_type=$7, ref_id=$8, block_title=$9, block_body=$10, rendered_html=$11
			WHERE id=$3 AND spec_id=$2
			RETURNING updated_at, subspec_id, block_title, block_body, rendered_html,
			-- select number of unread comments
			(SELECT COUNT(*) FROM spec_community_comment AS c
				LEFT JOIN spec_community_read AS r
					ON r.user_id = $1 AND r.target_type = 'comment' AND r.target_id = c.id
				WHERE c.target_type = 'block' AND c.target_id = spec_block.id
					AND r.user_id IS NULL
			) AS unread_count,
			-- select total number of comments
			(SELECT COUNT(*)
				FROM spec_community_comment AS c
				WHERE c.target_type = 'block' AND c.target_id = spec_block.id
			) AS comments_count`,
			userID, specID, blockID, time.Now(),
			styleType, contentType, refType, refID, title, body, renderedHTML,
		).Scan(&block.Updated, &block.SubspecID, &block.Title, &block.Body, &block.HTML,
			&block.UnreadCount, &block.CommentsCount)
		if err != nil {
			logError(r, userID, fmt.Errorf("updating block: %w", err))
			return nil, http.StatusInternalServerError
		}

		if status := recordSpecBlocksUpdated(tx, r, userID, specID); status != http.StatusOK {
			return nil, status
		}
		if block.SubspecID != nil {
			if status := recordSubspecBlocksUpdated(tx, r, userID, *block.SubspecID); status != http.StatusOK {
				return nil, status
			}
		}

		return block, http.StatusOK
	})
}

func ajaxSpecMoveBlocks(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	err := r.ParseForm()
	if err != nil {
		logError(r, userID, err)
		return nil, http.StatusInternalServerError
	}

	blockIDs, err := AtoInt64Array(r.Form.Get("blockIds"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing blockIds: %w", err))
		return nil, http.StatusBadRequest
	}

	if len(blockIDs) == 0 {
		logError(r, userID, fmt.Errorf("blank blockIds: %w", err))
		return nil, http.StatusBadRequest
	}

	// all the blocks should come from the same spec and subspec.
	// currnently multiselect move is supported from only a single source context.
	// spec membership is verified by verifyWriteSpecSubspecBlocks.
	var specID int64
	var sourceSubspecID *int64
	err = db.QueryRow(
		`SELECT spec_id, subspec_id FROM spec_block WHERE id = $1`,
		blockIDs[0]).Scan(&specID, &sourceSubspecID)
	if err != nil {
		logError(r, userID, fmt.Errorf("loading specID for block %d: %w", blockIDs[0], err))
		return nil, http.StatusInternalServerError
	}

	targetSubspecID, err := AtoInt64NilIfEmpty(r.Form.Get("subspecId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing subsapceId: %w", err))
		return nil, http.StatusBadRequest
	}

	parentID, err := AtoInt64NilIfEmpty(r.Form.Get("parentId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing parentId: %w", err))
		return nil, http.StatusBadRequest
	}

	insertBeforeID, err := AtoInt64NilIfEmpty(r.Form.Get("insertBeforeId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing insertBeforeId: %w", err))
		return nil, http.StatusBadRequest
	}

	var verifyBlockIds = blockIDs[:]
	if parentID != nil {
		verifyBlockIds = append(verifyBlockIds, *parentID)
	}
	if insertBeforeID != nil {
		verifyBlockIds = append(verifyBlockIds, *insertBeforeID)
	}
	if access, status := verifyWriteSpecSubspecBlocks(r, db, userID, specID, targetSubspecID,
		verifyBlockIds...); !access {
		return nil, status
	}

	return handleInTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		insertAt, code, err := makeInsertAt(tx, specID, targetSubspecID, parentID, insertBeforeID, len(blockIDs))
		if err != nil {
			logError(r, userID, fmt.Errorf("making insert position: %w", err))
			return 0, code
		}

		var args = []interface{}{parentID}

		var orderNumbers []interface{}
		for i := 0; i < len(blockIDs); i++ {
			orderNumbers = append(orderNumbers, insertAt+i)
		}

		var values = createIDsValuesMap(&args, blockIDs, orderNumbers)

		_, err = tx.Query(
			`UPDATE spec_block
			SET parent_id = $1, order_number = v.order_number
			FROM (`+values+`) AS v(id, order_number)
			WHERE spec_block.id = v.id`,
			args...)
		if err != nil {
			logError(r, userID, fmt.Errorf("moving blocks: %w", err))
			return nil, http.StatusInternalServerError
		}

		subspecChanged := (sourceSubspecID == nil && targetSubspecID != nil) ||
			(sourceSubspecID != nil && targetSubspecID == nil) ||
			(sourceSubspecID != nil && targetSubspecID != nil && *sourceSubspecID != *targetSubspecID)

		if subspecChanged {
			// Recursively set subspec_id
			var args = []interface{}{targetSubspecID}
			var placeholders = createArgsListInt64s(&args, blockIDs...)
			_, err = tx.Exec(
				`WITH RECURSIVE block_tree(id) AS (
					-- Anchor
					SELECT id
					FROM spec_block
					WHERE id IN (`+placeholders+`)
					UNION ALL
					-- Recursive Member
					SELECT spec_block.id
					FROM spec_block, block_tree
					WHERE spec_block.parent_id = block_tree.id
				)
				-- Update original table
				UPDATE spec_block
				SET subspec_id = $1
				WHERE spec_block.id IN (SELECT id FROM block_tree)`,
				args...)
			if err != nil {
				logError(r, userID, fmt.Errorf("moving block tree to subspec: %w", err))
				return nil, http.StatusInternalServerError
			}
		}

		if status := recordSpecBlocksUpdated(tx, r, userID, specID); status != http.StatusOK {
			return nil, status
		}

		if targetSubspecID != nil {
			if status := recordSubspecBlocksUpdated(db, r, userID, *targetSubspecID); status != http.StatusOK {
				return nil, status
			}
		}

		if subspecChanged && sourceSubspecID != nil {
			if status := recordSubspecBlocksUpdated(db, r, userID, *sourceSubspecID); status != http.StatusOK {
				return nil, status
			}
		}

		if subspecChanged {
			blocks, err := loadBlocksByID(tx, userID, specID, blockIDs...)
			if err != nil {
				logError(r, userID, fmt.Errorf("loading blocks: %w", err))
				return nil, http.StatusInternalServerError
			}
			return blocks, http.StatusOK
		}

		return nil, http.StatusOK
	})
}

func ajaxSpecDeleteBlock(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	err := r.ParseForm()
	if err != nil {
		logError(r, userID, err)
		return nil, http.StatusInternalServerError
	}

	blockID, err := AtoInt64(r.Form.Get("blockId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing blockId: %w", err))
		return nil, http.StatusBadRequest
	}

	var specID int64
	var subspecID *int64
	err = db.QueryRow(
		`SELECT spec_id, subspec_id
		FROM spec_block
		WHERE id = $1`,
		blockID,
	).Scan(&specID, &subspecID)

	if err != nil {
		logError(r, userID, fmt.Errorf("loading specID for block %d: %w", blockID, err))
		return nil, http.StatusInternalServerError
	}

	if access, status := verifyWriteSpecSubspecBlocks(r, db, userID, specID, subspecID, blockID); !access {
		return nil, status
	}

	return handleInTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		// Delete block row (delete is cascade; subblocks will also be deleted)
		_, err := tx.Exec(
			`DELETE FROM spec_block
			WHERE id=$1`,
			blockID)
		if err != nil {
			logError(r, userID, fmt.Errorf("deleting block: %w", err))
			return nil, http.StatusInternalServerError
		}

		if status := recordSpecBlocksUpdated(tx, r, userID, specID); status != http.StatusOK {
			return nil, status
		}
		if subspecID != nil {
			if status := recordSubspecBlocksUpdated(db, r, userID, *subspecID); status != http.StatusOK {
				return nil, status
			}
		}

		return nil, http.StatusOK
	})
}
