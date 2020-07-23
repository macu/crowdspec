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
