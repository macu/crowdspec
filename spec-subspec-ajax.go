package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func ajaxSubspecs(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET
	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid specId: %w", err)
	}

	// TODO Verify read access

	rows, err := db.Query(`SELECT id, created_at, subspec_name, subspec_desc
		FROM spec_subspec
		WHERE spec_id = $1
		ORDER BY subspec_name`, specID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("querying blocks: %w", err)
	}

	subspecs := []*SpecSubspec{}

	for rows.Next() {
		s := &SpecSubspec{
			SpecID: specID,
		}
		err = rows.Scan(&s.ID, &s.Created, &s.Name, &s.Desc)
		if err != nil {
			if err2 := rows.Close(); err2 != nil { // TODO Add everywhere
				return nil, http.StatusInternalServerError, fmt.Errorf("error closing rows: %s; on scan error: %w", err2, err)
			}
			return nil, http.StatusInternalServerError, fmt.Errorf("error scanning subspec: %w", err)
		}
		subspecs = append(subspecs, s)
	}

	return subspecs, http.StatusOK, nil
}

func ajaxSpecCreateSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specID, err := AtoInt64(r.Form.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing specId: %w", err)
	}

	// TODO Verify write access

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("subspec name required")
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {
		var subspecID int64
		err = tx.QueryRow(`INSERT INTO spec_subspec (spec_id, created_at, subspec_name, subspec_desc)
			VALUES ($1, $2, $3, $4) RETURNING id`,
			specID, time.Now(), name, desc).Scan(&subspecID)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("inserting new subspec: %w", err)
		}
		return subspecID, http.StatusCreated, nil
	})
}

func ajaxSpecSaveSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	subspecID, err := AtoInt64(r.Form.Get("subspecId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing subspecId: %w", err)
	}

	// TODO Verify write access

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("subspec name required")
	}

	desc := AtoPointerNilIfEmpty(strings.TrimSpace(r.Form.Get("desc")))

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		s := &SpecSubspec{
			ID: subspecID,
		}

		err = tx.QueryRow(`UPDATE spec_subspec SET subspec_name = $2, subspec_desc = $3
			WHERE id = $1
			RETURNING spec_id, created_at, subspec_name, subspec_desc, (
				SELECT spec_name FROM spec WHERE spec.id = spec_subspec.spec_id
			)`,
			subspecID, name, desc).Scan(&s.SpecID, &s.Created, &s.Name, &s.Desc, &s.SpecName)

		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("updating subspec: %w", err)
		}

		return s, http.StatusOK, nil
	})
}

func ajaxSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET
	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid specId: %w", err)
	}

	// TODO Verify read access

	subspecID, err := AtoInt64(query.Get("subspecId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid subspecId: %w", err)
	}

	s := &SpecSubspec{
		ID:     subspecID,
		SpecID: specID,
	}

	err = db.QueryRow(`SELECT spec_subspec.created_at, spec_subspec.subspec_name, spec_subspec.subspec_desc,
		spec.spec_name
		FROM spec_subspec
		INNER JOIN spec ON spec.id = $2
		WHERE spec_subspec.id = $1
		AND spec_subspec.spec_id = $2`,
		subspecID, specID).Scan(&s.Created, &s.Name, &s.Desc, &s.SpecName)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("loading subspec: %w", err)
	}

	s.Blocks, err = loadBlocks(db, specID, &subspecID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("loading subspec block: %w", err)
	}

	return s, http.StatusOK, nil
}

func ajaxSpecDeleteSubspec(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	subspecID, err := AtoInt64(r.Form.Get("subspecId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing subspecId: %w", err)
	}

	// TODO Verify write access

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {
		_, err := tx.Exec(`UPDATE spec_block SET ref_type = NULL, ref_id = NULL
				WHERE ref_type=$1 AND ref_id=$2`, BlockRefSubspec, subspecID)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("clearing block references: %w", err)
		}

		_, err = tx.Exec(`DELETE FROM spec_subspec WHERE id=$1`, subspecID)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("deleting subspec: %w", err)
		}

		return nil, http.StatusOK, nil
	})
}
