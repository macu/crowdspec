package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Spec represents a db spec row
type Spec struct {
	ID        int64     `json:"id"`
	OwnerType string    `json:"ownerType"`
	OwnerID   int64     `json:"ownerId"`
	Created   time.Time `json:"created"`
	Name      string    `json:"name"`
	Desc      *string   `json:"desc"`
	Public    bool      `json:"public"`

	Blocks []*SpecBlock `json:"blocks,omitempty"`
}

// SpecSubspace represents a portion of the spec that is loaded separately.
type SpecSubspace struct {
	ID      int64     `json:"id"`
	SpecID  int64     `json:"specId"`
	Created time.Time `json:"created"`
	Name    string    `json:"name"`
	Desc    *string   `json:"desc"`

	Blocks []*SpecBlock `json:"blocks,omitempty"`
}

// SpecBlock represents a section within a spec or spec subspace.
type SpecBlock struct {
	ID          int64   `json:"id"`
	SpecID      int64   `json:"specId"`
	SubspaceID  *int64  `json:"subspaceId"` // may be null (belongs to spec directly)
	ParentID    *int64  `json:"parentId"`   // may be null (root level)
	OrderNumber int     `json:"orderNumber"`
	Type        string  `json:"type"` //  markdown, plaintext, subspace
	RefID       *int64  `json:"refId"`
	Title       *string `json:"title"`
	Body        *string `json:"body"`

	RefItem interface{} `json:"refItem,omitempty"`

	SubBlocks []*SpecBlock `json:"subblocks,omitempty"`
}

const (
	// BlockTypeText indicates a text block with basic html.
	BlockTypeText = "text"
	// BlockTypeBullet indicates a bullet point with basic html.
	BlockTypeBullet = "bullet"
	// BlockTypeNumbered indicates a numbered list point with basic html.
	BlockTypeNumbered = "numbered"
	// BlockTypeImageRef indicates an image block with optional local title and body.
	BlockTypeImageRef = "image-ref"
	// BlockTypeVideoRef indicates a video block with optional local title and body.
	BlockTypeVideoRef = "video-ref"
	// BlockTypeSubspaceRef indicates a subspace reference block with optional local title and body.
	BlockTypeSubspaceRef = "subspace-ref"
	// BlockTypeSpecRef indicates a spec reference block with optional local title and body.
	BlockTypeSpecRef = "spec-ref"
)

func isValidBlockType(t string) bool {
	return t == BlockTypeText || t == BlockTypeBullet || t == BlockTypeNumbered ||
		t == BlockTypeImageRef || t == BlockTypeVideoRef ||
		t == BlockTypeSubspaceRef || t == BlockTypeSpecRef
}

func isReferenceBlockType(t string) bool {
	return t == BlockTypeImageRef || t == BlockTypeVideoRef ||
		t == BlockTypeSubspaceRef || t == BlockTypeSpecRef
}

// Returns the ID of the newly created spec.
func ajaxCreateSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// TODO ALlow creating within an org

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("Blank spec name")
	}

	desc := strings.TrimSpace(r.Form.Get("desc"))

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {
		var specID int64
		err := tx.QueryRow(`
				INSERT INTO spec (owner_type, owner_id, created_at, spec_name, spec_desc)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id
				`, "user", userID, time.Now(), name, desc).Scan(&specID)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return specID, http.StatusOK, nil
	})
}

func ajaxSaveSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	return nil, http.StatusNotImplemented, nil
}

func ajaxDeleteSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	return nil, http.StatusNotImplemented, nil
}

// Returns the requested spec with immediate blocks.
func ajaxSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET
	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("Invalid specId: %s", err)
	}

	// TODO Verify read access

	// TODO Finish owner_name, user_is_admin, user_is_contributor
	s := &Spec{}
	row := db.QueryRow(`
		SELECT spec.id, spec.created_at, spec.owner_type, spec.owner_id, spec.spec_name, spec.spec_desc, spec.is_public
		FROM spec
		INNER JOIN user_account
		ON user_account.id=spec.owner_id
		WHERE spec.id=$1 AND spec.owner_type=$2 AND spec.owner_id=$3
		`, specID, "user", userID)
	err = row.Scan(&s.ID, &s.Created, &s.OwnerType, &s.OwnerID, &s.Name, &s.Desc, &s.Public)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	s.Blocks, err = loadBlocks(db, specID, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return s, http.StatusOK, nil
}

func loadBlocks(db *sql.DB, specID int64, subspaceID *int64) ([]*SpecBlock, error) {
	var rows *sql.Rows
	var err error
	if subspaceID != nil {
		rows, err = db.Query(`
			SELECT spec_block.id, spec_block.spec_id, spec_block.subspace_id, spec_block.parent_id,
			spec_block.order_number,
			spec_block.block_type, spec_block.ref_id,
			spec_block.block_title, spec_block.block_body,
			spec_subspace.subspace_name, spec_subspace.subspace_desc, spec_subspace.created_at AS subspace_created_at
			FROM spec_block
			LEFT JOIN spec_subspace
			ON spec_block.block_type=$1
			AND spec_subspace.id=spec_block.ref_id
			WHERE spec_block.spec_id=$2 AND spec_block.subspace_id=$3
			ORDER BY spec_block.parent_id, spec_block.order_number
			`, "subspace-ref", specID, subspaceID)
	} else {
		rows, err = db.Query(`
			SELECT spec_block.id, spec_block.spec_id, spec_block.subspace_id, spec_block.parent_id,
			spec_block.order_number,
			spec_block.block_type, spec_block.ref_id,
			spec_block.block_title, spec_block.block_body,
			NULL AS subspace_name, NULL AS subspace_desc, NULL AS subspace_created_at
			FROM spec_block
			WHERE spec_block.spec_id=$1 AND spec_block.subspace_id IS NULL
			ORDER BY spec_block.parent_id, spec_block.order_number
			`, specID)
	}
	if err != nil {
		return nil, err
	}

	blocks := []*SpecBlock{}
	blocksByID := map[int64]*SpecBlock{}

	for rows.Next() {
		b := &SpecBlock{}
		var subspaceName, subspaceDesc *string
		var subspaceCreated *time.Time
		err = rows.Scan(&b.ID, &b.SpecID, &b.SubspaceID, &b.ParentID, &b.OrderNumber,
			&b.Type, &b.RefID, &b.Title, &b.Body, &subspaceName, &subspaceDesc, &subspaceCreated)
		if err != nil {
			return nil, err
		}
		if b.Type == "subspace-ref" && subspaceName != nil {
			b.RefItem = &SpecSubspace{
				ID:      *b.RefID,
				SpecID:  specID,
				Created: *subspaceCreated,
				Name:    *subspaceName,
				Desc:    subspaceDesc,
			}
		}
		blocks = append(blocks, b)
		blocksByID[b.ID] = b
	}

	rootBlocks := []*SpecBlock{}
	for _, b := range blocks {
		if b.ParentID == nil {
			rootBlocks = append(rootBlocks, b)
		} else {
			parentBlock, ok := blocksByID[*b.ParentID]
			if ok {
				parentBlock.SubBlocks = append(parentBlock.SubBlocks, b)
			}
		}
	}

	return rootBlocks, nil
}

// Returns a list of the current user's specs.
func ajaxUserSpecs(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET
	rows, err := db.Query(`
		SELECT id, owner_type, owner_id, created_at, spec_name, spec_desc, is_public
		FROM spec
		WHERE owner_type='user' AND owner_id=$1
		ORDER BY created_at DESC
		`, userID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specs := []Spec{}
	for rows.Next() {
		s := Spec{}
		err = rows.Scan(&s.ID, &s.OwnerType, &s.OwnerID, &s.Created, &s.Name, &s.Desc, &s.Public)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		specs = append(specs, s)
	}

	return specs, http.StatusOK, nil
}

// Creates a block within a spec.
func ajaxSpecCreateBlock(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specID, err := AtoInt64(r.Form.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// TODO Verify write access to spec

	subspaceID, err := AtoInt64ne(r.Form.Get("subspaceId"))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// TODO Verify subspace is within spec

	parentID, err := AtoInt64ne(r.Form.Get("parentId"))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// TODO Verify parent block is within spec/subspace

	insertAt, err := AtoInt(r.Form.Get("insertAt"))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	refType := r.Form.Get("refType")
	if !isValidBlockType(refType) {
		return nil, http.StatusBadRequest, fmt.Errorf("Invalid refType: %s", refType)
	}

	refID, err := AtoInt64ne(r.Form.Get("refId"))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if refID == nil && isReferenceBlockType(refType) {
		return nil, http.StatusBadRequest, fmt.Errorf("refId is required for reference blocks")
	}

	title := strings.TrimSpace(r.Form.Get("title"))

	body := strings.TrimSpace(r.Form.Get("body"))

	// TODO Html sanitize title and body

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		var err error

		if insertAt == -1 {
			// Insert at end - get next order_number
			err = tx.QueryRow(`
				SELECT COALESCE(MAX(order_number), -1) + 1 AS insert_at
				FROM spec_block
				WHERE spec_id = $1
				AND subspace_id = $2
				AND parent_id = $3
				`, specID, subspaceID, parentID).Scan(&insertAt)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
		} else {
			// Increase order numbers of following blocks
			_, err = tx.Exec(`
				UPDATE spec_block
				SET order_number = order_number + 1
				WHERE spec_id = $1
				AND subspace_id = $2
				AND parent_id = $3
				AND order_number >= $4
				`, specID, subspaceID, parentID, insertAt)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
		}

		// Create block row
		var blockID int64
		err = tx.QueryRow(`
				INSERT INTO spec_block (spec_id, subspace_id, parent_id, order_number, block_type, ref_id, block_title, block_body)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id
				`, specID, subspaceID, parentID, insertAt, refType, refID, title, body).Scan(&blockID)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		// Return block
		block := &SpecBlock{
			ID:          blockID,
			SpecID:      specID,
			SubspaceID:  subspaceID,
			ParentID:    parentID,
			OrderNumber: insertAt,
			Type:        refType,
			RefID:       refID,
		}
		if title != "" {
			block.Title = &title
		}
		if body != "" {
			block.Body = &body
		}

		return block, http.StatusOK, nil
	})
}

func ajaxSpecSaveBlock(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	return nil, http.StatusNotImplemented, nil
}

func ajaxSpecMoveBlock(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	return nil, http.StatusNotImplemented, nil
}

func ajaxSpecDeleteBlock(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	return nil, http.StatusNotImplemented, nil
}

func ajaxSpecCreateSubspace(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	return nil, http.StatusNotImplemented, nil
}

func ajaxSpecSaveSubspace(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	return nil, http.StatusNotImplemented, nil
}

func ajaxSpecDeleteSubspace(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	return nil, http.StatusNotImplemented, nil
}

// func ajaxSpecAddSubpoint(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
// 	// POST
// 	err := r.ParseForm()
// 	if err != nil {
// 		return nil, http.StatusInternalServerError, err
// 	}
//
// 	specID, err := AtoUint(r.Form.Get("specId"))
// 	if specID == 0 || err != nil {
// 		return nil, http.StatusBadRequest, fmt.Errorf("Valid specId required: %d, %v", specID, err)
// 	}
//
// 	// TODO Verify write access to spec
//
// 	parentID, err := AtoUint(r.Form.Get("parentId"))
// 	if err != nil {
// 		return nil, http.StatusBadRequest, fmt.Errorf("Valid parentId required: %v", err)
// 	}
//
// 	// TODO Verify parent point exists in same spec
//
// 	title := strings.TrimSpace(r.Form.Get("title"))
// 	desc := strings.TrimSpace(r.Form.Get("desc"))
// 	if title == "" && desc == "" {
// 		return nil, http.StatusBadRequest, fmt.Errorf("Either title or desc required")
// 	}
//
// 	orderNumber, err := AtoUint(r.Form.Get("orderNumber"))
// 	if err != nil {
// 		orderNumber = 0
// 	}
//
// 	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {
// 		res, err := tx.Exec(`
// 			INSERT INTO spec_subpoint (spec_id, created, parent_id, title, description, order_number)
// 			VALUES (?, ?, ?, ?, ?, ?)
// 			`, specID, time.Now(), parentID, title, desc, orderNumber)
// 		if err != nil {
// 			return nil, http.StatusInternalServerError, err
// 		}
//
// 		newID, err := res.LastInsertId()
// 		if err != nil {
// 			return nil, http.StatusInternalServerError, err
// 		}
//
// 		// Load new point from DB
// 		newPoint := &SpecSubpoint{}
// 		row := tx.QueryRow(`
// 			SELECT id, spec_id, created, parent_id, title, description, order_number
// 			FROM spec_subpoint
// 			WHERE id=?
// 			`, newID)
// 		err = row.Scan(&newPoint.ID, &newPoint.SpecID, &newPoint.Created,
// 			&newPoint.ParentID, &newPoint.Title, &newPoint.Desc, &newPoint.OrderNumber)
// 		if err != nil {
// 			return nil, http.StatusInternalServerError, fmt.Errorf("Failed to read new subpoint: %v", err)
// 		}
//
// 		return newPoint, http.StatusCreated, nil
// 	})
// }
