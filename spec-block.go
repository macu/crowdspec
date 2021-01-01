package main

import (
	"database/sql"
	"fmt"
	"time"
)

const blockTitleMaxLen = 255

// SpecBlock represents a section within a spec or spec subspec.
type SpecBlock struct {
	ID          int64     `json:"id"`
	SpecID      int64     `json:"specId"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	SubspecID   *int64    `json:"subspecId"` // may be null (belongs to spec directly)
	ParentID    *int64    `json:"parentId"`  // may be null (root level)
	OrderNumber int       `json:"orderNumber"`
	StyleType   string    `json:"styleType"`
	ContentType *string   `json:"contentType"`
	RefType     *string   `json:"refType"`
	RefID       *int64    `json:"refId"`
	Title       *string   `json:"title"`
	Body        *string   `json:"body"`

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

// Load a sinle block and subblocks.
func loadBlock(c DBConn, blockID int64) (*SpecBlock, error) {

	// Ref items are only loaded if belonging to the same spec.
	// TODO Allow linking to any ref item, but verify current user's access when joining info.
	args := []interface{}{blockID}
	query := `
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
		SELECT spec_block.id, spec_block.spec_id, spec_block.created_at, spec_block.updated_at,
		spec_block.subspec_id, spec_block.parent_id, spec_block.order_number,
		spec_block.style_type, spec_block.content_type, spec_block.ref_type, spec_block.ref_id,
		spec_block.block_title, spec_block.block_body,
		spec_subspec.spec_id AS subspec_spec_id, spec_subspec.subspec_name, spec_subspec.subspec_desc,
		spec_url.spec_id AS url_spec_id, spec_url.created_at AS url_created, spec_url.updated_at AS url_updated,
		spec_url.url AS url_url, spec_url.url_title, spec_url.url_desc, spec_url.url_image_data
		FROM spec_block
		LEFT JOIN spec_subspec
		ON spec_block.ref_type = ` + argPlaceholder(BlockRefSubspec, &args) + `
		AND spec_subspec.id = spec_block.ref_id
		AND spec_subspec.spec_id = spec_block.spec_id
		LEFT JOIN spec_url
		ON spec_block.ref_type = ` + argPlaceholder(BlockRefURL, &args) + `
		AND spec_url.id = spec_block.ref_id
		AND spec_url.spec_id = spec_block.spec_id
		WHERE spec_block.id IN (SELECT id FROM block_tree)
		ORDER BY spec_block.parent_id, spec_block.order_number`

	rows, err := c.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying blocks: %w", err)
	}

	blocks := []*SpecBlock{}
	blocksByID := map[int64]*SpecBlock{}
	readBlocks(rows, &blocks, &blocksByID)

	return blocksByID[blockID], nil
}

// load the blocks in a spec or subspec
func loadBlocks(db *sql.DB, specID int64, subspecID *int64) ([]*SpecBlock, error) {

	// Ref items are only loaded if belonging to the same spec.
	// TODO Allow linking to any ref item, but verify current user's access when joining info.
	args := []interface{}{specID}
	query := `
		SELECT spec_block.id, spec_block.spec_id, spec_block.created_at, spec_block.updated_at,
		spec_block.subspec_id, spec_block.parent_id, spec_block.order_number,
		spec_block.style_type, spec_block.content_type, spec_block.ref_type, spec_block.ref_id,
		spec_block.block_title, spec_block.block_body,
		spec_subspec.spec_id AS subspec_spec_id, spec_subspec.subspec_name, spec_subspec.subspec_desc,
		spec_url.spec_id AS url_spec_id, spec_url.created_at AS url_created, spec_url.updated_at AS url_updated,
		spec_url.url AS url_url, spec_url.url_title, spec_url.url_desc, spec_url.url_image_data
		FROM spec_block
		LEFT JOIN spec_subspec
		ON spec_block.ref_type=` + argPlaceholder(BlockRefSubspec, &args) + `
		AND spec_subspec.id=spec_block.ref_id
		AND spec_subspec.spec_id = spec_block.spec_id
		LEFT JOIN spec_url
		ON spec_block.ref_type=` + argPlaceholder(BlockRefURL, &args) + `
		AND spec_url.id=spec_block.ref_id
		AND spec_url.spec_id = spec_block.spec_id
		WHERE spec_block.spec_id=$1
		AND ` + eqCond("spec_block.subspec_id", subspecID, &args) + `
		ORDER BY spec_block.parent_id, spec_block.order_number`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying blocks: %w", err)
	}

	blocks := []*SpecBlock{}
	blocksByID := map[int64]*SpecBlock{}
	readBlocks(rows, &blocks, &blocksByID)

	rootBlocks := []*SpecBlock{}
	for _, b := range blocks {
		if b.ParentID == nil {
			rootBlocks = append(rootBlocks, b)
		}
	}

	return rootBlocks, nil
}

func readBlocks(rows *sql.Rows, blocks *[]*SpecBlock, blocksByID *map[int64]*SpecBlock) error {

	for rows.Next() {
		b := &SpecBlock{}
		var subspecSpecID, urlSpecID *int64
		var subspecName, subspecDesc *string
		var urlCreated, urlUpdated *time.Time
		var urlURL, urlTitle, urlDesc, urlImageData *string

		err := rows.Scan(&b.ID, &b.SpecID, &b.Created, &b.Updated,
			&b.SubspecID, &b.ParentID, &b.OrderNumber,
			&b.StyleType, &b.ContentType, &b.RefType, &b.RefID, &b.Title, &b.Body,
			&subspecSpecID, &subspecName, &subspecDesc,
			&urlSpecID, &urlCreated, &urlUpdated, &urlURL, &urlTitle, &urlDesc, &urlImageData)
		if err != nil {
			if err2 := rows.Close(); err2 != nil { // TODO Add everywhere
				return fmt.Errorf("closing rows: %s; on scan error: %w", err2, err)
			}
			return fmt.Errorf("scanning block: %w", err)
		}

		if b.RefType != nil {
			switch *b.RefType {
			case BlockRefURL:
				if urlSpecID != nil {
					b.RefItem = &URLObject{
						ID:        *b.RefID,
						SpecID:    *urlSpecID,
						Created:   *urlCreated,
						URL:       *urlURL,
						Title:     urlTitle,
						Desc:      urlDesc,
						ImageData: urlImageData,
						Updated:   *urlUpdated,
					}
				}
			case BlockRefSubspec:
				if subspecSpecID != nil {
					b.RefItem = &SpecSubspec{
						ID:     *b.RefID,
						SpecID: *subspecSpecID,
						Name:   *subspecName,
						Desc:   subspecDesc,
					}
				}
			}
		}

		*blocks = append(*blocks, b)
		(*blocksByID)[b.ID] = b
	}

	// Link blocks to parents
	for _, b := range *blocks {
		if b.ParentID != nil {
			parentBlock, ok := (*blocksByID)[*b.ParentID]
			if ok {
				parentBlock.SubBlocks = append(parentBlock.SubBlocks, b)
			}
		}
	}

	return nil
}
