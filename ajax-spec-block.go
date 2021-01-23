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

	if access, status := verifyWriteSpecSubspecBlocks(r, db, userID, specID, subspecID,
		parentID, insertBeforeID); !access {
		return nil, status
	}

	styleType := r.Form.Get("styleType")
	if !isValidListStyleType(styleType) {
		logError(r, userID, fmt.Errorf("invalid styleType: %s", styleType))
		return nil, http.StatusBadRequest
	}

	contentType := AtoPointerNilIfEmpty(r.Form.Get("contentType"))
	if contentType != nil && !isValidTextContentType(*contentType) {
		logError(r, userID, fmt.Errorf("invalid contentType: %s", *contentType))
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

	// TODO Html sanitize title and body

	return inTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

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
		insertAt, code, err := makeInsertAt(tx, specID, subspecID, parentID, insertBeforeID)
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
				style_type, content_type, ref_type, ref_id, block_title, block_body)
			VALUES ($1, $2, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id, created_at, updated_at, block_title, block_body
			`, specID, time.Now(), subspecID, parentID, insertAt,
			styleType, contentType, refType, refID, title, body,
		).Scan(&block.ID, &block.Created, &block.Updated, &block.Title, &block.Body)
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

	contentType := AtoPointerNilIfEmpty(r.Form.Get("contentType"))
	if contentType != nil && !isValidTextContentType(*contentType) {
		logError(r, userID, fmt.Errorf("invalid contentType: %s", *contentType))
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

	// TODO Html sanitize title and body

	return inTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

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
			SET updated_at=$4, style_type=$5, content_type=$6, ref_type=$7, ref_id=$8, block_title=$9, block_body=$10
			WHERE id=$3 AND spec_id=$2
			RETURNING updated_at, subspec_id, block_title, block_body,
			-- select number of unread comments
			(SELECT COUNT(c.id) FROM spec_community_comment AS c
				LEFT JOIN spec_community_read AS r
					ON r.user_id = $1 AND r.target_type = 'comment' AND r.target_id = c.id
				WHERE c.spec_id = spec_block.spec_id
					AND c.target_type = 'block' AND c.target_id = spec_block.id
					AND r.user_id IS NULL
					) AS unread_count
			`, userID, specID, blockID, time.Now(),
			styleType, contentType, refType, refID, title, body,
		).Scan(&block.Updated, &block.SubspecID, &block.Title, &block.Body, &block.UnreadCount)
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

func ajaxSpecMoveBlock(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
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
	var sourceSubspecID *int64
	err = db.QueryRow(`
		SELECT spec_id, subspec_id
		FROM spec_block
		WHERE id = $1
		`, blockID).Scan(&specID, &sourceSubspecID)
	if err != nil {
		logError(r, userID, fmt.Errorf("loading specID for block %d: %w", blockID, err))
		return nil, http.StatusInternalServerError
	}

	subspecID, err := AtoInt64NilIfEmpty(r.Form.Get("subspecId"))
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

	if access, status := verifyWriteSpecSubspecBlocks(r, db, userID, specID, subspecID,
		parentID, insertBeforeID, &blockID); !access {
		return nil, status
	}

	return inTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		insertAt, code, err := makeInsertAt(tx, specID, subspecID, parentID, insertBeforeID)
		if err != nil {
			logError(r, userID, fmt.Errorf("making insert position: %w", err))
			return 0, code
		}

		_, err = tx.Exec(`
			UPDATE spec_block
			SET subspec_id = $2, parent_id = $3, order_number = $4
			WHERE id = $1
			`, blockID, subspecID, parentID, insertAt)
		if err != nil {
			logError(r, userID, fmt.Errorf("moving block: %w", err))
			return nil, http.StatusInternalServerError
		}

		subspecChanged := (sourceSubspecID == nil && subspecID != nil) ||
			(sourceSubspecID != nil && subspecID == nil) ||
			(sourceSubspecID != nil && subspecID != nil && *sourceSubspecID != *subspecID)

		if subspecChanged {
			// Recursively set subspec_id
			_, err = tx.Exec(`
				WITH RECURSIVE block_tree(id) AS (
					-- Anchor
					SELECT id
					FROM spec_block
					WHERE id = $1
					UNION ALL
					-- Recursive Member
					SELECT spec_block.id
					FROM spec_block, block_tree
					WHERE spec_block.parent_id = block_tree.id
				)
				-- Update original table
				UPDATE spec_block
				SET subspec_id = $2
				WHERE spec_block.id IN (SELECT id FROM block_tree)
				`, blockID, subspecID)
			if err != nil {
				logError(r, userID, fmt.Errorf("moving block tree to subspec: %w", err))
				return nil, http.StatusInternalServerError
			}
		}

		if status := recordSpecBlocksUpdated(tx, r, userID, specID); status != http.StatusOK {
			return nil, status
		}

		if subspecID != nil {
			if status := recordSubspecBlocksUpdated(db, r, userID, *subspecID); status != http.StatusOK {
				return nil, status
			}
		}

		if subspecChanged && sourceSubspecID != nil {
			if status := recordSubspecBlocksUpdated(db, r, userID, *sourceSubspecID); status != http.StatusOK {
				return nil, status
			}
		}

		if subspecChanged {
			block, err := loadBlock(tx, userID, blockID)
			if err != nil {
				logError(r, userID, fmt.Errorf("loading blocks: %w", err))
				return nil, http.StatusInternalServerError
			}
			return block, http.StatusOK
		}

		return nil, http.StatusOK
	})
}

// Increments order numbers of blocks starting at the specified block,
// and returns the order number preceeding that block.
// If insertBeforeID is nil, returns the order number at the end of the list.
func makeInsertAt(tx *sql.Tx, specID int64, subspecID *int64, parentID *int64, insertBeforeID *int64) (int, int, error) {
	var insertAt int

	if insertBeforeID == nil {
		// Insert at end - get next order_number

		args := []interface{}{specID}

		query := `
			SELECT COALESCE(MAX(order_number), -1) + 1 AS insert_at FROM spec_block
			WHERE spec_id = $1
			AND ` + eqCond("subspec_id", subspecID, &args) + `
			AND ` + eqCond("parent_id", parentID, &args)

		err := tx.QueryRow(query, args...).Scan(&insertAt)
		if err != nil {
			return 0, http.StatusInternalServerError, fmt.Errorf("selecting next order number: %w", err)
		}

		return insertAt, http.StatusOK, nil
	}

	// Increase order numbers of following blocks

	args := []interface{}{specID}

	query := `UPDATE spec_block
		SET order_number = order_number + 1
		WHERE spec_id = $1
		AND ` + eqCond("subspec_id", subspecID, &args) + `
		AND ` + eqCond("parent_id", parentID, &args) + `
		AND order_number >= (
			SELECT insert_before_block.order_number
			FROM spec_block AS insert_before_block
			WHERE insert_before_block.id = ` + argPlaceholder(*insertBeforeID, &args) + `
		)`

	_, err := tx.Exec(query, args...)
	if err != nil {
		return 0, http.StatusInternalServerError, fmt.Errorf("incrementing order numbers: %w", err)
	}

	// Get order number after preceeding block

	args = []interface{}{specID}

	query = `SELECT COALESCE(MAX(order_number), -1) + 1 FROM spec_block
		WHERE spec_id = $1
		AND ` + eqCond("subspec_id", subspecID, &args) + `
		AND ` + eqCond("parent_id", parentID, &args) + `
		AND order_number < (
			SELECT insert_before_block.order_number
			FROM spec_block AS insert_before_block
			WHERE insert_before_block.id = ` + argPlaceholder(*insertBeforeID, &args) + `
		)`

	err = tx.QueryRow(query, args...).Scan(&insertAt)
	if err != nil {
		return 0, http.StatusInternalServerError, fmt.Errorf("selecting preceeding order number: %w", err)
	}

	return insertAt, http.StatusOK, nil
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
	err = db.QueryRow(`
			SELECT spec_id, subspec_id
			FROM spec_block
			WHERE id = $1
			`, blockID).Scan(&specID, &subspecID)
	if err != nil {
		logError(r, userID, fmt.Errorf("loading specID for block %d: %w", blockID, err))
		return nil, http.StatusInternalServerError
	}

	if access, status := verifyWriteSpecSubspecBlocks(r, db, userID, specID, subspecID, &blockID); !access {
		return nil, status
	}

	return inTransaction(r, db, userID, func(tx *sql.Tx) (interface{}, int) {

		// Delete block row (delete is cascade; subblocks will also be deleted)
		_, err := tx.Exec(`
			DELETE FROM spec_block
			WHERE id=$1
			`, blockID)
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
