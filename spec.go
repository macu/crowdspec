package main

import "time"

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

	// Joined from owner if owner is user
	Username string `json:"username,omitempty"`

	// Root level blocks in this spec
	Blocks []*SpecBlock `json:"blocks,omitempty"`
}

const (
	// OwnerTypeUser represents an individual user as the owner of a database entity.
	OwnerTypeUser = "user"
	// OwnerTypeOrg  = "org"
)
