package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

const subspecNameMaxLen = 255

// SpecSubspec represents a portion of the spec that is loaded separately.
type SpecSubspec struct {
	ID      int64     `json:"id"`
	SpecID  int64     `json:"specId"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	Name    string    `json:"name"`
	Desc    *string   `json:"desc"`

	// convenience fields (joined for certain responses)
	// Representing time spec was loaded from db
	RenderTime time.Time `json:"renderTime,omitempty"`
	SpecName   string    `json:"specName,omitempty"`
	OwnerType  string    `json:"ownerType,omitempty"`
	OwnerID    int64     `json:"ownerId,omitempty"`

	// Community attributes
	UnreadCount   uint `json:"unreadCount"`
	CommentsCount uint `json:"commentsCount"`

	// Note on omitempty: https://play.golang.org/p/Lk_FdWeL4i8
	// empty slice will be omitted
	Blocks []*SpecBlock `json:"blocks,omitempty"`
}

// SpecSubspecHeader represents a subspec preview in the community modal.
type SpecSubspecHeader struct {
	ID      int64     `json:"id"`
	SpecID  int64     `json:"specId"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	Name    string    `json:"name"`
	Desc    *string   `json:"desc"`
}

// called when creating or saving block with new subspec params
func createSubspec(tx *sql.Tx, specID int64, name string, desc *string) (*SpecSubspecHeader, error) {
	var s = &SpecSubspecHeader{
		SpecID: specID,
	}

	name = Substr(name, subspecNameMaxLen)

	err := tx.QueryRow(
		`INSERT INTO spec_subspec (
			spec_id, created_at, updated_at, subspec_name, subspec_desc
		) VALUES (
			$1, $2, $2, $3, $4
		) RETURNING id, created_at, updated_at, subspec_name, subspec_desc`,
		specID, time.Now(), name, desc,
	).Scan(&s.ID, &s.Created, &s.Updated, &s.Name, &s.Desc)
	if err != nil {
		return nil, fmt.Errorf("creating subspec: %w", err)
	}

	return s, nil
}

// currently used in loading block ref headers piecewise
func loadSubspecHeader(db DBConn, subspecID int64) (*SpecSubspecHeader, error) {
	var s = &SpecSubspecHeader{
		ID: subspecID,
	}
	var err = db.QueryRow(
		`SELECT spec_id, created_at, updated_at,
			subspec_name, subspec_desc
		FROM spec_subspec
		WHERE id = $1`,
		subspecID,
	).Scan(&s.SpecID, &s.Created, &s.Updated, &s.Name, &s.Desc)
	if err != nil {
		return nil, fmt.Errorf("reading subspec: %w", err)
	}
	return s, nil
}

func recordSubspecBlocksUpdated(db DBConn, r *http.Request, userID uint, subspecID int64) int {
	_, err := db.Exec(`UPDATE spec_subspec SET blocks_updated_at=$2 WHERE id=$1`, subspecID, time.Now())
	if err != nil {
		logError(r, userID, fmt.Errorf("recording update time on subspec: %w", err))
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
