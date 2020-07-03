package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

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

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

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
		return specID, http.StatusCreated, nil
	})
}

func ajaxSaveSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specID, err := AtoInt64(r.Form.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing specId: %w", err)
	}

	// TODO Verify access

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("Blank spec name")
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		_, err := tx.Exec(`UPDATE spec SET spec_name=$1, spec_desc=$2 WHERE id=$3`, name, desc, specID)

		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("updating spec: %w", err)
		}

		spec := &Spec{
			ID:   specID,
			Name: name,
			Desc: desc,
		}

		return spec, http.StatusOK, nil
	})
}

func ajaxDeleteSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specID, err := AtoInt64(r.Form.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing specId: %w", err)
	}

	// TODO Verify access

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		_, err := tx.Exec(`DELETE FROM spec WHERE id=$1`, specID)

		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("deleting spec: %w", err)
		}

		return nil, http.StatusOK, nil
	})
}

// Returns the requested spec with immediate blocks.
func ajaxSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET
	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("Invalid specId: %w", err)
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

	args := []interface{}{specID}
	query := `
		SELECT spec_block.id, spec_block.spec_id, spec_block.subspace_id, spec_block.parent_id,
		spec_block.order_number,
		spec_block.style_type, spec_block.content_type, spec_block.ref_type, spec_block.ref_id,
		spec_block.block_title, spec_block.block_body,
		spec_subspace.subspace_name, spec_subspace.subspace_desc, spec_subspace.created_at AS subspace_created_at,
		spec_block_url.url AS url_url, spec_block_url.url_title, spec_block_url.url_desc, spec_block_url.url_image_data
		FROM spec_block
		LEFT JOIN spec_subspace
		ON spec_block.ref_type='subspace'
		AND spec_subspace.id=spec_block.ref_id
		LEFT JOIN spec_block_url
		ON spec_block.ref_type='url'
		AND spec_block_url.id=spec_block.ref_id
		WHERE spec_block.spec_id=$1
		AND spec_block.` + subspaceCond(subspaceID, &args) + `
		ORDER BY spec_block.parent_id, spec_block.order_number`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	blocks := []*SpecBlock{}
	blocksByID := map[int64]*SpecBlock{}

	for rows.Next() {
		b := &SpecBlock{}
		var subspaceName, subspaceDesc *string
		var subspaceCreated *time.Time
		var urlURL, urlTitle, urlDesc, urlImageData *string
		err = rows.Scan(&b.ID, &b.SpecID, &b.SubspaceID, &b.ParentID, &b.OrderNumber,
			&b.StyleType, &b.ContentType, &b.RefType, &b.RefID, &b.Title, &b.Body,
			&subspaceName, &subspaceDesc, &subspaceCreated,
			&urlURL, &urlTitle, &urlDesc, &urlImageData)
		if err != nil {
			if err2 := rows.Close(); err2 != nil { // TODO Add everywhere
				return nil, fmt.Errorf("error closing rows: %s; on scan error: %w", err2, err)
			}
			return nil, fmt.Errorf("error scanning spec: %w", err)
		}
		if b.RefType != nil {
			switch *b.RefType {
			case BlockRefURL:
				if urlURL != nil {
					b.RefItem = &URLObject{
						ID:        *b.RefID,
						URL:       *urlURL,
						Title:     urlTitle,
						Desc:      urlDesc,
						ImageData: urlImageData,
					}
				}
			case BlockRefSubspace:
				b.RefItem = &SpecSubspace{
					ID:      *b.RefID,
					SpecID:  specID,
					Created: *subspaceCreated,
					Name:    *subspaceName,
					Desc:    subspaceDesc,
				}
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
