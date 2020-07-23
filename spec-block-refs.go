package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Returns nil if fields are valid for creating or setting a ref item during block create or update.
func validateCreateRefItemFields(fields url.Values) (*string, *int64, error) {
	refType := AtoPointerNilIfEmpty(fields.Get("refType"))
	if refType == nil {
		// Valid ref fields include null refType
		return nil, nil, nil
	} else if !isValidBlockRefType(*refType) {
		return nil, nil, fmt.Errorf("invalid refType: %s", *refType)
	}

	refID, err := AtoInt64NilIfEmpty(fields.Get("refId"))
	if err != nil {
		return nil, nil, fmt.Errorf("parsing refId: %w", err)
	}
	if refID != nil {
		// Valid ref fields include refId
		// TODO Validate refID refers to accessible object
		return refType, refID, nil
	}

	switch *refType {
	case BlockRefURL:
		refURL := AtoPointerNilIfEmpty(strings.TrimSpace(fields.Get("refUrl")))
		if refURL == nil {
			return nil, nil, fmt.Errorf("refUrl required for refType: %s", *refType)
		}
		// TODO validate URL syntax
		return refType, nil, nil

	case BlockRefSubspec:
		refName := AtoPointerNilIfEmpty(strings.TrimSpace(fields.Get("refName")))
		if refName == nil {
			return nil, nil, fmt.Errorf("refName required for refType: %s", *refType)
		}
		return refType, nil, nil
	}

	return nil, nil, fmt.Errorf("refId required")
}

// Creates and returns a ref item during block create or update.
func handleCreateRefItem(tx *sql.Tx, specID int64, fields url.Values) (*int64, interface{}, error) {
	refType := fields.Get("refType")
	switch refType {

	case BlockRefURL:
		refURL := strings.TrimSpace(fields.Get("refUrl"))
		refItem, err := createURLObject(tx, specID, refURL)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating URL ref item: %w", err)
		}
		return &refItem.ID, refItem, nil

	case BlockRefSubspec:
		refName := strings.TrimSpace(fields.Get("refName"))
		refDesc := AtoPointerNilIfEmpty(strings.TrimSpace(fields.Get("refDesc")))
		refItem, err := createSubspec(tx, specID, refName, refDesc)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("error loading subspec: %w", err)
		}
		return &refItem.ID, refItem, nil

	default:
		return nil, nil, fmt.Errorf("unsupported refType: %s", refType)
	}
}

func loadRefItem(db DBConn, refType string, refID int64) (interface{}, error) {
	switch refType {

	case BlockRefURL:
		return loadURLObject(db, refID)

	case BlockRefSubspec:
		return loadSubspecHeader(db, refID)

	default:
		return nil, fmt.Errorf("unsupported refType: %s", refType)
	}
}
