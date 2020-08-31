package main

import (
	"database/sql"
	"fmt"
	"time"
)

// SpecSubspec represents a portion of the spec that is loaded separately.
type SpecSubspec struct {
	ID      int64     `json:"id"`
	SpecID  int64     `json:"specId"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	Name    string    `json:"name"`
	Desc    *string   `json:"desc"`

	// convenience fields (joined for certain responses)
	SpecName  string `json:"specName,omitempty"`
	OwnerType string `json:"ownerType,omitempty"`
	OwnerID   int64  `json:"ownerId,omitempty"`

	Blocks []*SpecBlock `json:"blocks,omitempty"`
}

// called when creating or saving block with new subspec params
func createSubspec(tx *sql.Tx, specID int64, name string, desc *string) (*SpecSubspec, error) {
	s := &SpecSubspec{
		SpecID: specID,
	}

	err := tx.QueryRow(`INSERT INTO spec_subspec (spec_id, created_at, updated_at, subspec_name, subspec_desc)
		VALUES ($1, $2, $2, $3, $4)
		RETURNING id, created_at, updated_at, subspec_name, subspec_desc`,
		specID, time.Now(), name, desc).Scan(&s.ID, &s.Created, &s.Updated, &s.Name, &s.Desc)
	if err != nil {
		return nil, fmt.Errorf("inserting new subspec: %w", err)
	}

	return s, nil
}

// currently used in loading block ref headers piecewise
func loadSubspecHeader(db DBConn, subspecID int64) (*SpecSubspec, error) {
	s := &SpecSubspec{ID: subspecID}
	err := db.QueryRow(`SELECT spec_id, subspec_name, subspec_desc
		FROM spec_subspec
		WHERE id = $1`,
		subspecID).Scan(&s.SpecID, &s.Name, &s.Desc)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func recordSubspecBlocksUpdated(db DBConn, subspecID int64) error {
	_, err := db.Exec(`UPDATE spec_subspec SET blocks_updated_at=$2 WHERE id=$1`, subspecID, time.Now())
	return err
}
