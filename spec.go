package main

import (
	"fmt"
	"net/http"
	"time"
)

const spenNameMaxLen = 255

// Spec represents a db spec row
type Spec struct {
	ID        int64     `json:"id"`
	OwnerType string    `json:"ownerType"`
	OwnerID   int64     `json:"ownerId"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	Name      string    `json:"name"`
	Desc      *string   `json:"desc"`
	Public    bool      `json:"public"`

	// Representing time spec was loaded from db
	RenderTime time.Time `json:"renderTime,omitempty"`

	// Joined from owner if owner is user
	Username  string  `json:"username,omitempty"`
	Highlight *string `json:"highlight,omitempty"`

	// Community attributes
	UnreadCount uint `json:"unreadCount"`

	// Root level blocks in this spec
	Blocks []*SpecBlock `json:"blocks,omitempty"`
}

const (
	// OwnerTypeUser represents an individual user as the owner of a database entity.
	OwnerTypeUser = "user"
	// OwnerTypeOrg  = "org"
)

func recordSpecBlocksUpdated(db DBConn, r *http.Request, userID uint, specID int64) int {
	_, err := db.Exec(`UPDATE spec SET blocks_updated_at=$2 WHERE id=$1`, specID, time.Now())
	if err != nil {
		logError(r, userID, fmt.Errorf("recording update time on spec: %w", err))
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
