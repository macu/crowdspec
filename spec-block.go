package main

import (
	"database/sql"
	"fmt"
	"time"
)

// SpecBlock represents a section within a spec or spec subspec.
type SpecBlock struct {
	ID          int64   `json:"id"`
	SpecID      int64   `json:"specId"`
	SubspecID   *int64  `json:"subspecId"` // may be null (belongs to spec directly)
	ParentID    *int64  `json:"parentId"`  // may be null (root level)
	OrderNumber int     `json:"orderNumber"`
	StyleType   string  `json:"styleType"`
	ContentType *string `json:"contentType"`
	RefType     *string `json:"refType"`
	RefID       *int64  `json:"refId"`
	Title       *string `json:"title"`
	Body        *string `json:"body"`

	RefItem interface{} `json:"refItem,omitempty"`

	SubBlocks []*SpecBlock `json:"subblocks,omitempty"`
}

const (
	// ListStyleBullet indicates a bullet point style list item.
	ListStyleBullet = "bullet"
	// ListStyleNumbered indicates a list item numbered relative to other numbered items in the same list.
	ListStyleNumbered = "numbered"
	// ListStyleNone indicates a list item with no bullet.
	ListStyleNone = "none"

	// TextContentPlain indicates plaintext with potential newlines.
	TextContentPlain = "plaintext"
	// TextContentMarkdown indicates markdown processing is required for rendering.
	TextContentMarkdown = "markdown"
	// TextContentHTML indicates content should be sanitized and rendered as HTML.
	TextContentHTML = "html"

	// BlockRefOrg indicates a reference to an organisation.
	BlockRefOrg = "org"
	// BlockRefSpec indicates a reference to a spec.
	BlockRefSpec = "spec"
	// BlockRefSubspec indicates a reference to a subspec in this or another spec.
	BlockRefSubspec = "subspec"
	// BlockRefBlock indicates a reference to a block in this or another spec.
	BlockRefBlock = "block"
	// BlockRefImage indicates an image reference owned by the spec owner.
	BlockRefImage = "image"
	// BlockRefVideo indicates a reference to an external video.
	BlockRefVideo = "video"
	// BlockRefURL indicates a reference to a URL.
	BlockRefURL = "url"
	// BlockRefFile indicates a reference to a file owned by the spec owner.
	BlockRefFile = "file"
)

func isValidListStyleType(t string) bool {
	return stringInSlice(t, []string{
		ListStyleBullet,
		ListStyleNumbered,
		ListStyleNone,
	})
}

func isValidTextContentType(t string) bool {
	return stringInSlice(t, []string{
		TextContentPlain,
		// TextContentMarkdown,
		// TextContentHTML,
	})
}

func isValidBlockRefType(t string) bool {
	return stringInSlice(t, []string{
		// BlockRefOrg,
		// BlockRefSpec,
		BlockRefSubspec,
		// BlockRefBlock,
		// BlockRefImage,
		// BlockRefVideo,
		BlockRefURL,
		// BlockRefFile,
	})
}

func isRefIDRequiredForRefType(t *string) bool {
	if t == nil {
		return false
	}
	return stringInSlice(*t, []string{
		BlockRefOrg,
		BlockRefSpec,
		BlockRefSubspec,
		BlockRefBlock,
		BlockRefFile,
	})
}

func isURLRequiredForRefType(t *string) bool {
	if t == nil {
		return false
	}
	return stringInSlice(*t, []string{
		BlockRefURL,
	})
}

func loadBlocks(db *sql.DB, specID int64, subspecID *int64) ([]*SpecBlock, error) {

	args := []interface{}{specID}
	query := `
		SELECT spec_block.id, spec_block.spec_id, spec_block.subspec_id, spec_block.parent_id,
		spec_block.order_number,
		spec_block.style_type, spec_block.content_type, spec_block.ref_type, spec_block.ref_id,
		spec_block.block_title, spec_block.block_body,
		spec_subspec.spec_id AS subspec_spec_id,
		spec_subspec.subspec_name, spec_subspec.subspec_desc, spec_subspec.created_at AS subspec_created_at,
		spec_url.spec_id AS url_spec_id, spec_url.updated_at AS url_updated_at,
		spec_url.url AS url_url, spec_url.url_title, spec_url.url_desc, spec_url.url_image_data
		FROM spec_block
		LEFT JOIN spec_subspec
		ON spec_block.ref_type=` + argPlaceholder(BlockRefSubspec, &args) + `
		AND spec_subspec.id=spec_block.ref_id
		LEFT JOIN spec_url
		ON spec_block.ref_type=` + argPlaceholder(BlockRefURL, &args) + `
		AND spec_url.id=spec_block.ref_id
		WHERE spec_block.spec_id=$1
		AND spec_block.` + subspecCond(subspecID, &args) + `
		ORDER BY spec_block.parent_id, spec_block.order_number`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying blocks: %w", err)
	}

	blocks := []*SpecBlock{}
	blocksByID := map[int64]*SpecBlock{}

	for rows.Next() {
		b := &SpecBlock{}
		var subspecSpecID, urlSpecID *int64
		var subspecName, subspecDesc *string
		var subspecCreated, urlUpdated *time.Time
		var urlURL, urlTitle, urlDesc, urlImageData *string
		err = rows.Scan(&b.ID, &b.SpecID, &b.SubspecID, &b.ParentID, &b.OrderNumber,
			&b.StyleType, &b.ContentType, &b.RefType, &b.RefID, &b.Title, &b.Body,
			&subspecSpecID, &subspecName, &subspecDesc, &subspecCreated,
			&urlSpecID, &urlUpdated, &urlURL, &urlTitle, &urlDesc, &urlImageData)
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
						SpecID:    *urlSpecID,
						URL:       *urlURL,
						Title:     urlTitle,
						Desc:      urlDesc,
						ImageData: urlImageData,
						UpdatedAt: *urlUpdated,
					}
				}
			case BlockRefSubspec:
				b.RefItem = &SpecSubspec{
					ID:      *b.RefID,
					SpecID:  *subspecSpecID,
					Created: *subspecCreated,
					Name:    *subspecName,
					Desc:    subspecDesc,
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
