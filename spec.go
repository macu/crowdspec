package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Spec represents a db spec row
type Spec struct {
	ID      uint      `json:"id"`
	Created time.Time `json:"created"`

	OwnerType string `json:"ownerType"`
	OwnerID   uint   `json:"ownerId"`
	OwnerName string `json:"ownerName"`

	Name string `json:"name"`
	Desc string `json:"desc"`

	Public bool `json:"public"`

	RootPoints []*SpecSubpoint `json:"points,omitempty"`
}

// SpecUserAccess represents a spec accessed by a user
type SpecUserAccess struct {
	Spec

	UserIsAdmin       bool `json:"userIsAdmin"`
	UserIsContributor bool `json:"userIsContributor"`
}

// SpecSubpoint represents a subpoint in a nesting list spec.
type SpecSubpoint struct {
	ID      uint      `json:"id"`
	SpecID  uint      `json:"specId"`
	Created time.Time `json:"created"`

	ParentID uint `json:"parentId"`

	Title string `json:"title"`
	Desc  string `json:"desc"`

	OrderNumber uint `json:"orderNumber"`

	SubPoints []*SpecSubpoint `json:"points,omitempty"`
}

// Returns the ID of the newly created spec.
func ajaxCreateSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST
	var ownerType string
	var ownerID uint

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	orgID := r.Form.Get("orgId")
	if strings.TrimSpace(orgID) != "" {
		// TODO Verify org admin
		ownerType = "org"
		ownerID, err = AtoUint(orgID)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
	} else {
		ownerType = "user"
		ownerID = userID
	}

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("Blank spec name")
	}

	desc := strings.TrimSpace(r.Form.Get("desc"))

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {
		res, err := tx.Exec(`
				INSERT INTO spec (owner_type, owner_id, created, name, description) VALUES (?, ?, ?, ?, ?)
				`, ownerType, ownerID, time.Now(), name, desc)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		specID, err := res.LastInsertId()
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return specID, http.StatusOK, nil
	})
}

func ajaxSpec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET
	query := r.URL.Query()

	specID := query.Get("specId")
	if specID == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("No specId provided")
	}

	// TODO Verify read access

	// TODO Finish owner_name, user_is_admin, user_is_contributor
	s := &SpecUserAccess{}
	row := db.QueryRow(`
		SELECT spec.id, spec.created, spec.owner_type, spec.owner_id, spec.name, spec.description, spec.public,
			IF(spec.owner_type="user", user.username, "TODO") AS owner_name,
			IF(spec.owner_type="user" AND spec.owner_id=?, 1, 0) AS user_is_admin,
			IF(spec.owner_type="user" AND spec.owner_id=?, 1, 0) AS user_is_contributor
		FROM spec
		LEFT JOIN user
			ON spec.owner_type="user"
			AND user.id=spec.owner_id
		WHERE spec.id=?
		`, userID, userID, specID)
	err := row.Scan(&s.ID, &s.Created, &s.OwnerType, &s.OwnerID, &s.Name, &s.Desc, &s.Public,
		&s.OwnerName, &s.UserIsAdmin, &s.UserIsContributor)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	points := []*SpecSubpoint{}
	pointsByID := map[uint]*SpecSubpoint{}
	rows, err := db.Query(`
		SELECT id, spec_id, created, parent_id, title, description, order_number
		FROM spec_subpoint
		WHERE spec_id=?
		ORDER BY order_number
		`, specID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	for rows.Next() {
		p := &SpecSubpoint{}
		err = rows.Scan(&p.ID, &p.SpecID, &p.Created, &p.ParentID, &p.Title, &p.Desc, &p.OrderNumber)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		points = append(points, p)
		pointsByID[p.ID] = p
	}

	rootPoints := []*SpecSubpoint{}
	for _, p := range points {
		if p.ParentID == 0 {
			rootPoints = append(rootPoints, p)
		} else {
			parentPoint, ok := pointsByID[p.ParentID]
			if ok {
				parentPoint.SubPoints = append(parentPoint.SubPoints, p)
			}
		}
	}

	s.RootPoints = rootPoints

	return s, http.StatusOK, nil
}

func ajaxUserSpecs(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET
	rows, err := db.Query(`
		SELECT spec.id, spec.created, spec.owner_type, spec.owner_id, spec.name, spec.description, spec.public,
			user.username AS owner_name
		FROM spec
		INNER JOIN user
			ON user.id=spec.owner_id
		WHERE spec.owner_type="user" AND spec.owner_id=?
		`, userID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specs := []Spec{}
	for rows.Next() {
		s := Spec{}
		err = rows.Scan(&s.ID, &s.Created, &s.OwnerType, &s.OwnerID, &s.Name, &s.Desc, &s.Public,
			&s.OwnerName)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		specs = append(specs, s)
	}

	return specs, http.StatusOK, nil
}

func ajaxPublicSpecs(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET
	// TODO Finish owner_name, user_is_admin, user_is_contributor
	rows, err := db.Query(`
		SELECT spec.id, spec.created, spec.owner_type, spec.owner_id, spec.name, spec.description, spec.public,
			IF(spec.owner_type="user", user.username, "TODO") AS owner_name,
			IF(user.id=?, 1, 0) AS user_is_admin,
			IF(user.id=?, 1, 0) AS user_is_contributor
		FROM spec
		INNER JOIN user
			ON user.id=spec.owner_id
		WHERE spec.public=1
		`, userID, userID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specs := []SpecUserAccess{}
	for rows.Next() {
		s := SpecUserAccess{}
		err = rows.Scan(&s.ID, &s.Created, &s.OwnerType, &s.OwnerID, &s.Name, &s.Desc, &s.Public,
			&s.OwnerName, &s.UserIsAdmin, &s.UserIsContributor)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		specs = append(specs, s)
	}

	return specs, http.StatusOK, nil
}

func ajaxSpecAddSubpoint(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST
	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specID, err := AtoUint(r.Form.Get("specId"))
	if specID == 0 || err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("Valid specId required: %d, %v", specID, err)
	}

	// TODO Verify write access to spec

	parentID, err := AtoUint(r.Form.Get("parentId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("Valid parentId required: %v", err)
	}

	// TODO Verify parent point exists in same spec

	title := strings.TrimSpace(r.Form.Get("title"))
	desc := strings.TrimSpace(r.Form.Get("desc"))
	if title == "" && desc == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("Either title or desc required")
	}

	orderNumber, err := AtoUint(r.Form.Get("orderNumber"))
	if err != nil {
		orderNumber = 0
	}

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {
		res, err := tx.Exec(`
			INSERT INTO spec_subpoint (spec_id, created, parent_id, title, description, order_number)
			VALUES (?, ?, ?, ?, ?, ?)
			`, specID, time.Now(), parentID, title, desc, orderNumber)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		newID, err := res.LastInsertId()
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		// Load new point from DB
		newPoint := &SpecSubpoint{}
		row := tx.QueryRow(`
			SELECT id, spec_id, created, parent_id, title, description, order_number
			FROM spec_subpoint
			WHERE id=?
			`, newID)
		err = row.Scan(&newPoint.ID, &newPoint.SpecID, &newPoint.Created,
			&newPoint.ParentID, &newPoint.Title, &newPoint.Desc, &newPoint.OrderNumber)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("Failed to read new subpoint: %v", err)
		}

		return newPoint, http.StatusCreated, nil
	})
}
