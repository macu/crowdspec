package main

import (
	"database/sql"
	"fmt"
	"net/http"
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

	// Community attributes
	UnreadCount   uint `json:"unreadCount"`
	CommentsCount uint `json:"commentsCount"`

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

func loadBlocksByID(db DBConn, userID uint, specID int64, blockIDs ...int64) ([]*SpecBlock, error) {

	var args = []interface{}{specID, userID}

	var orderNumbers []interface{}
	for i := 0; i < len(blockIDs); i++ {
		orderNumbers = append(orderNumbers, i)
	}

	// Used to sort blocks by order requested
	var values = createIDsValuesMap(&args, blockIDs, orderNumbers)

	// Ref items are only loaded if belonging to the same spec.
	// TODO Allow linking to any ref item, but verify current user's access when joining info.
	query := `
		WITH RECURSIVE block_tree(id) AS (
			-- Anchor
			SELECT spec_block.id, selected.order_number
			FROM spec_block
			INNER JOIN (` + values + `) AS selected(id, order_number)
				ON selected.id = spec_block.id
			WHERE spec_id = $1
			UNION ALL
			-- Recursive Member
			SELECT spec_block.id, spec_block.order_number
			FROM spec_block, block_tree
			WHERE spec_block.spec_id = $1
				AND spec_block.parent_id = block_tree.id
		)
		SELECT spec_block.id, spec_block.spec_id, spec_block.created_at, spec_block.updated_at,
		spec_block.subspec_id, spec_block.parent_id, spec_block.order_number,
		spec_block.style_type, spec_block.content_type, spec_block.ref_type, spec_block.ref_id,
		spec_block.block_title, spec_block.block_body,
		ref_subspec.spec_id AS subspec_spec_id, ref_subspec.subspec_name, ref_subspec.subspec_desc,
		ref_url.spec_id AS url_spec_id, ref_url.created_at AS url_created, ref_url.updated_at AS url_updated,
		ref_url.url AS url_url, ref_url.url_title, ref_url.url_desc, ref_url.url_image_data,
		-- select number of unread comments
		(SELECT COUNT(c.id) FROM spec_community_comment AS c
			LEFT JOIN spec_community_read AS r
				ON r.user_id = $2 AND r.target_type = 'comment' AND r.target_id = c.id
			WHERE c.spec_id = spec_block.spec_id
				AND c.target_type = 'block' AND c.target_id = spec_block.id
				AND r.user_id IS NULL
		) AS unread_count,
		-- select total number of comments
		(SELECT COUNT(*) FROM spec_community_comment AS c
			WHERE c.spec_id = spec_block.spec_id
				AND c.target_type = 'block' AND c.target_id = spec_block.id
		) AS comments_count
		FROM spec_block
		INNER JOIN block_tree
			ON block_tree.id = spec_block.id
		LEFT JOIN spec_subspec AS ref_subspec
			ON spec_block.ref_type = ` + argPlaceholder(BlockRefSubspec, &args) + `
			AND ref_subspec.id = spec_block.ref_id
			AND ref_subspec.spec_id = spec_block.spec_id
		LEFT JOIN spec_url AS ref_url
			ON spec_block.ref_type = ` + argPlaceholder(BlockRefURL, &args) + `
			AND ref_url.id = spec_block.ref_id
			AND ref_url.spec_id = spec_block.spec_id
		ORDER BY spec_block.parent_id, block_tree.order_number`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying blocks: %w", err)
	}

	blocks := []*SpecBlock{}
	blocksByID := map[int64]*SpecBlock{}
	readBlocks(rows, &blocks, &blocksByID)

	requestedBlocks := []*SpecBlock{}
	for _, b := range blocks {
		if int64InSlice(b.ID, blockIDs) {
			requestedBlocks = append(requestedBlocks, b)
		}
	}

	return requestedBlocks, nil
}

// load the blocks in a spec or subspec
func loadContextBlocks(db *sql.DB, userID uint, specID int64, subspecID *int64) ([]*SpecBlock, error) {

	// Ref items are only loaded if belonging to the same spec.
	// TODO Allow linking to any ref item, but verify current user's access when joining info.
	args := []interface{}{userID, specID}
	query := `
		SELECT spec_block.id, spec_block.spec_id, spec_block.created_at, spec_block.updated_at,
		spec_block.subspec_id, spec_block.parent_id, spec_block.order_number,
		spec_block.style_type, spec_block.content_type, spec_block.ref_type, spec_block.ref_id,
		spec_block.block_title, spec_block.block_body,
		ref_subspec.spec_id AS subspec_spec_id, ref_subspec.subspec_name, ref_subspec.subspec_desc,
		ref_url.spec_id AS url_spec_id, ref_url.created_at AS url_created, ref_url.updated_at AS url_updated,
		ref_url.url AS url_url, ref_url.url_title, ref_url.url_desc, ref_url.url_image_data,
		-- select number of unread comments
		(SELECT COUNT(c.id) FROM spec_community_comment AS c
			LEFT JOIN spec_community_read AS r
				ON r.user_id = $1 AND r.target_type = 'comment' AND r.target_id = c.id
			WHERE c.spec_id = spec_block.spec_id
				AND c.target_type = 'block' AND c.target_id = spec_block.id
				AND r.user_id IS NULL
		) AS unread_count,
		-- select total number of comments
		(SELECT COUNT(*) FROM spec_community_comment AS c
			WHERE c.spec_id = spec_block.spec_id
				AND c.target_type = 'block' AND c.target_id = spec_block.id
		) AS comments_count
		FROM spec_block
		LEFT JOIN spec_subspec AS ref_subspec
			ON spec_block.ref_type=` + argPlaceholder(BlockRefSubspec, &args) + `
			AND ref_subspec.id=spec_block.ref_id
			AND ref_subspec.spec_id = spec_block.spec_id
		LEFT JOIN spec_url AS ref_url
			ON spec_block.ref_type=` + argPlaceholder(BlockRefURL, &args) + `
			AND ref_url.id=spec_block.ref_id
			AND ref_url.spec_id = spec_block.spec_id
		WHERE spec_block.spec_id=$2
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
			&urlSpecID, &urlCreated, &urlUpdated, &urlURL, &urlTitle, &urlDesc, &urlImageData,
			&b.UnreadCount, &b.CommentsCount)
		if err != nil {
			if err2 := rows.Close(); err2 != nil {
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

// Increments order numbers of blocks starting at the specified block,
// and returns the order number preceeding that block.
// If insertBeforeID is nil, returns the order number at the end of the list.
func makeInsertAt(tx *sql.Tx,
	specID int64, subspecID *int64, parentID *int64, insertBeforeID *int64,
	requiredPositions int) (int, int, error) {

	if requiredPositions <= 0 {
		return 0, 0, fmt.Errorf("invalid requiredPositions: %d", requiredPositions)
	}

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

	args := []interface{}{specID, requiredPositions}

	query := `UPDATE spec_block
		SET order_number = order_number + $2
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
