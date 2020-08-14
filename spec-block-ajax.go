package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

// Creates a block within a spec.
func ajaxSpecCreateBlock(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specID, err := AtoInt64(r.Form.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing specId: %w", err)
	}

	subspecID, err := AtoInt64NilIfEmpty(r.Form.Get("subspecId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing subspecId: %w", err)
	}

	parentID, err := AtoInt64NilIfEmpty(r.Form.Get("parentId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing parentId: %w", err)
	}

	insertBeforeID, err := AtoInt64NilIfEmpty(r.Form.Get("insertBeforeId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing insertBeforeId: %w", err)
	}

	if access, err := verifyWriteSpecSubspecBlocks(db, userID, specID, subspecID,
		parentID, insertBeforeID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying write blocks: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("write blocks access denied to user %d in spec %d", userID, specID)
	}

	styleType := r.Form.Get("styleType")
	if !isValidListStyleType(styleType) {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid styleType: %s", styleType)
	}

	contentType := AtoPointerNilIfEmpty(r.Form.Get("contentType"))
	if contentType != nil && !isValidTextContentType(*contentType) {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid contentType: %s", *contentType)
	}

	refType, refID, err := validateCreateRefItemFields(r.Form)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid ref fields: %w", err)
	}

	if access, err := verifyRefAccess(db, specID, refType, refID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying ref access: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("ref access denied to user %d in spec %d, refType %s refID %d", userID, specID, *refType, *refID)
	}

	title := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("title")))

	body := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("body")))

	if refType == nil && title == nil && body == nil {
		return nil, http.StatusBadRequest, fmt.Errorf("empty blocks are not currently allowed")
	}

	// TODO Html sanitize title and body

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		var err error

		// Create or load ref item
		var refItem interface{}
		if refType != nil {
			if refID == nil {
				refID, refItem, err = handleCreateRefItem(tx, specID, r.Form)
				if err != nil {
					return nil, http.StatusInternalServerError, fmt.Errorf("error creating ref item: %w", err)
				}
			} else {
				refItem, err = loadRefItem(tx, *refType, *refID)
				if err != nil {
					return nil, http.StatusInternalServerError, fmt.Errorf("error loading ref item: %w", err)
				}
			}
		}

		// Prepare insert position
		insertAt, code, err := makeInsertAt(tx, specID, subspecID, parentID, insertBeforeID)
		if err != nil {
			return nil, code, err
		}

		// Create block row
		var blockID int64
		err = tx.QueryRow(`
			INSERT INTO spec_block
			(spec_id, subspec_id, parent_id, order_number,
				style_type, content_type, ref_type, ref_id, block_title, block_body)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id
			`, specID, subspecID, parentID, insertAt,
			styleType, contentType, refType, refID, title, body).Scan(&blockID)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("inserting block: %w", err)
		}

		// Return block
		block := &SpecBlock{
			ID:          blockID,
			SpecID:      specID,
			SubspecID:   subspecID,
			ParentID:    parentID,
			OrderNumber: insertAt,
			StyleType:   styleType,
			ContentType: contentType,
			RefType:     refType,
			RefID:       refID,
			Title:       title,
			Body:        body,
			RefItem:     refItem,
		}

		return block, http.StatusOK, nil
	})
}

func ajaxSpecSaveBlock(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specID, err := AtoInt64(r.Form.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing specId: %w", err)
	}

	blockID, err := AtoInt64(r.Form.Get("blockId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing blockId: %w", err)
	}

	if access, err := verifyWriteSpecBlock(db, userID, specID, blockID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying write block: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("write block access denied to user %d in spec %d", userID, specID)
	}

	styleType := r.Form.Get("styleType")
	if !isValidListStyleType(styleType) {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid styleType: %s", styleType)
	}

	contentType := AtoPointerNilIfEmpty(r.Form.Get("contentType"))
	if contentType != nil && !isValidTextContentType(*contentType) {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid contentType: %s", *contentType)
	}

	refType, refID, err := validateCreateRefItemFields(r.Form)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid ref fields: %w", err)
	}

	if access, err := verifyRefAccess(db, specID, refType, refID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying ref access: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("ref access denied to user %d in spec %d, refType %s refID %d", userID, specID, *refType, *refID)
	}

	title := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("title")))

	body := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("body")))

	if refType == nil && title == nil && body == nil {
		return nil, http.StatusBadRequest, fmt.Errorf("empty blocks are not currently allowed")
	}

	// TODO Html sanitize title and body

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		var err error

		// Create or load ref item
		var refItem interface{}
		if refType != nil {
			if refID == nil {
				refID, refItem, err = handleCreateRefItem(tx, specID, r.Form)
				if err != nil {
					return nil, http.StatusInternalServerError, fmt.Errorf("error creating ref item: %w", err)
				}
			} else {
				refItem, err = loadRefItem(tx, *refType, *refID)
				if err != nil {
					return nil, http.StatusInternalServerError, fmt.Errorf("error loading ref item: %w", err)
				}
			}
		}

		// Update block row
		_, err = tx.Exec(`
				UPDATE spec_block
				SET style_type=$3, content_type=$4, ref_type=$5, ref_id=$6, block_title=$7, block_body=$8
				WHERE id=$1 AND spec_id=$2
				`, blockID, specID, styleType, contentType, refType, refID, title, body)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("updating block: %w", err)
		}

		// Return updated block
		block := &SpecBlock{
			ID:          blockID,
			StyleType:   styleType,
			ContentType: contentType,
			RefType:     refType,
			RefID:       refID,
			Title:       title,
			Body:        body,
			RefItem:     refItem,
		}

		return block, http.StatusOK, nil
	})
}

func ajaxSpecMoveBlock(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	blockID, err := AtoInt64(r.Form.Get("blockId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing blockId: %w", err)
	}

	var specID int64
	err = db.QueryRow(`
		SELECT spec_id
		FROM spec_block
		WHERE id = $1
		`, blockID).Scan(&specID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("loading specID for block %d: %w", blockID, err)
	}

	subspecID, err := AtoInt64NilIfEmpty(r.Form.Get("subspecId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing subsapceId: %w", err)
	}

	parentID, err := AtoInt64NilIfEmpty(r.Form.Get("parentId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing parentId: %w", err)
	}

	insertBeforeID, err := AtoInt64NilIfEmpty(r.Form.Get("insertBeforeId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing insertBeforeId: %w", err)
	}

	if access, err := verifyWriteSpecSubspecBlocks(db, userID, specID, subspecID,
		parentID, insertBeforeID, &blockID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying write blocks: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("write blocks access denied to user %d in spec %d", userID, specID)
	}

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		insertAt, code, err := makeInsertAt(tx, specID, subspecID, parentID, insertBeforeID)
		if err != nil {
			return 0, code, err
		}

		query := `UPDATE spec_block
			SET subspec_id = $1, parent_id = $2, order_number = $3
			WHERE id = $4`

		_, err = tx.Exec(query, subspecID, parentID, insertAt, blockID)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("inserting at end: %w", err)
		}

		// TODO move sub blocks recursively to target subspec if changed
		// See https://stackoverflow.com/a/30274296/1597274

		return nil, http.StatusOK, nil
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
					AND ` + subspecCond(subspecID, &args) + `
					AND ` + parentCond(parentID, &args)

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
		AND ` + subspecCond(subspecID, &args) + `
		AND ` + parentCond(parentID, &args) + `
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
		AND ` + subspecCond(subspecID, &args) + `
		AND ` + parentCond(parentID, &args) + `
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

func ajaxSpecDeleteBlock(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	blockID, err := AtoInt64(r.Form.Get("blockId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing blockId: %w", err)
	}

	if access, err := verifyWriteBlock(db, userID, blockID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying write block: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("write blocks access denied to user %d on block %d", userID, blockID)
	}

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {
		// Delete block row (delete is cascade; subblocks will also be deleted)
		_, err := tx.Exec(`
			DELETE FROM spec_block
			WHERE id=$1
			`, blockID)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("deleting block: %w", err)
		}

		return nil, http.StatusOK, nil
	})
}
