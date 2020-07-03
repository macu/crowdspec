package main

import "time"

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
	// BlockRefSubspace indicates a reference to a subspace in this or another spec.
	BlockRefSubspace = "subspace"
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
		// BlockRefSubspace,
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
		BlockRefSubspace,
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
