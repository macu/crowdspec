package main

import (
	"fmt"
	"net/http"
	"strings"
)

// temporary usage limits
const maxSpecCount = 100
const maxSubspecCount = 100

func verifyWriteTarget(r *http.Request, db DBConn, userID uint, targetType string, targetID int64) (bool, int) {

	var allowed bool

	var err = db.QueryRow(`SELECT verify_write_target($1, $2, $3)`,
		targetType, targetID, userID,
	).Scan(&allowed)
	if err != nil {
		logError(r, &userID, fmt.Errorf("verifying write target access: %w; user %d", err, userID))
		return false, http.StatusInternalServerError
	}

	if !allowed {
		return false, http.StatusForbidden
	}

	return allowed, http.StatusOK

}

func verifyReadTarget(r *http.Request, db DBConn, userID *uint, targetType string, targetID int64) (bool, int) {

	var allowed bool

	if userID == nil {
		var err = db.QueryRow(`SELECT verify_read_target_public($1, $2)`,
			targetType, targetID,
		).Scan(&allowed)
		if err != nil {
			logError(r, userID, fmt.Errorf("verifying read target public access: %w", err))
			return false, http.StatusInternalServerError
		}
	} else {
		var err = db.QueryRow(`SELECT verify_read_target($1, $2, $3)`,
			targetType, targetID, *userID,
		).Scan(&allowed)
		if err != nil {
			logError(r, userID, fmt.Errorf("verifying read target access: %w; userID %d", err, *userID))
			return false, http.StatusInternalServerError
		}
	}

	if !allowed {
		return false, http.StatusForbidden
	}

	return allowed, http.StatusOK

}

func verifyDeleteTarget(r *http.Request, db DBConn, userID uint, targetType string, targetID int64) (bool, int) {

	var allowed bool

	var err = db.QueryRow(`SELECT verify_delete_target($1, $2, $3)`,
		targetType, targetID, userID,
	).Scan(&allowed)
	if err != nil {
		logError(r, &userID, fmt.Errorf("verifying delete target access: %w; user %d", err, userID))
		return false, http.StatusInternalServerError
	}

	if !allowed {
		return false, http.StatusForbidden
	}

	return allowed, http.StatusOK

}

func verifyCreateSpec(r *http.Request, db DBConn, userID uint) (bool, int) {

	var count uint
	var err = db.QueryRow(`SELECT COUNT(*) FROM spec
		WHERE owner_type = $1 AND owner_id = $2`,
		OwnerTypeUser, userID,
	).Scan(&count)
	if err != nil {
		logError(r, &userID, fmt.Errorf("validating create spec access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count < maxSpecCount {
		return true, http.StatusOK
	}

	// alert admin
	logError(r, &userID, fmt.Errorf("user %d attempted to exceed spec limit %d", userID, maxSpecCount))

	return false, http.StatusForbidden

}

func verifyCreateSubspec(r *http.Request, db DBConn, userID uint, specID int64) (bool, int) {

	if write, status := verifyWriteTarget(r, db, userID, CommunityTargetSpec, specID); !write {
		return write, status
	}

	var count uint
	var err = db.QueryRow(`SELECT COUNT(*) FROM spec
		INNER JOIN spec_subspec ON spec_subspec.spec_id = spec.id
		WHERE spec.id = $1`,
		specID,
	).Scan(&count)
	if err != nil {
		logError(r, &userID, fmt.Errorf("validating create subspec access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count < maxSubspecCount {
		return true, http.StatusOK
	}

	// alert admin
	logError(r, &userID,
		fmt.Errorf("user %d attempted to exceed subspec limit %d", userID, maxSubspecCount))

	return false, http.StatusForbidden

}

func verifyWriteURL(r *http.Request, db DBConn, userID uint, urlID int64) (bool, int) {

	var count uint
	err := db.QueryRow(
		`SELECT COUNT(*) FROM spec_url
		INNER JOIN spec
			ON spec.id = spec_url.spec_id
		WHERE spec_url.id = $1
			AND spec.owner_type = $2
			AND spec.owner_id = $3`,
		urlID, OwnerTypeUser, userID,
	).Scan(&count)

	if err != nil {
		logError(r, &userID, fmt.Errorf("validating write url access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, &userID, fmt.Errorf("write url access denied to user %d for url %d", userID, urlID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

// - verifies given subspec belongs to spec
// - verifies all given blocks belong to spec
func verifyWriteSpecSubspecBlocks(r *http.Request, db DBConn, userID uint,
	specID int64, subspecID *int64, blockIDs ...int64) (bool, int) {

	args := []interface{}{}

	subspecJoin := ""
	if subspecID != nil {
		subspecJoin = `
			INNER JOIN spec_subspec
				ON spec_subspec.id = ` + argPlaceholder(*subspecID, &args) + `
				AND spec_subspec.spec_id = spec.id
			`
	}

	blockJoins := []string{}
	for i, id := range blockIDs {
		tableName := "block_" + IntToA(i)
		blockJoins = append(blockJoins, `
			INNER JOIN spec_block AS `+tableName+`
				ON `+tableName+`.id = `+argPlaceholder(id, &args)+`
				AND `+tableName+`.spec_id = spec.id
			`)
	}

	var count uint
	err := db.QueryRow(
		`SELECT COUNT(*) FROM spec`+
			subspecJoin+
			strings.Join(blockJoins, "")+`
		WHERE spec.owner_type = `+argPlaceholder(OwnerTypeUser, &args)+`
			AND spec.owner_id = `+argPlaceholder(userID, &args)+`
			AND spec.id = `+argPlaceholder(specID, &args),
		args...).Scan(&count)

	if err != nil {
		logError(r, &userID, fmt.Errorf("validating write blocks access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, &userID, fmt.Errorf("write blocks access denied to user %d in spec %d", userID, specID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

// Validate an association between a spec and ref target.
// Current policy is that the ref item must belong to the spec it is referenced in.
func validateRefAccess(r *http.Request, db DBConn, userID uint,
	specID int64, refType *string, refID *int64) (bool, int) {

	// Function is used when checking request parameters,
	// which are valid when nil
	if refType == nil || refID == nil {
		return true, http.StatusOK
	}

	var count uint
	var err error

	switch *refType {

	case BlockRefURL:
		err = db.QueryRow(
			`SELECT COUNT(*) FROM spec_url
			WHERE id = $1 AND spec_id = $2`,
			*refID, specID).Scan(&count)

	case BlockRefSubspec:
		err = db.QueryRow(
			`SELECT COUNT(*) FROM spec_subspec
			WHERE id = $1 AND spec_id = $2`,
			*refID, specID).Scan(&count)

	default:
		logError(r, &userID, fmt.Errorf("unsupported refType: %s", *refType))
		return false, http.StatusBadRequest
	}

	if err != nil {
		logError(r, &userID, fmt.Errorf("validating read ref access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, &userID,
			fmt.Errorf("read ref access denied to user %d in spec %d, refType %s refID %d",
				userID, specID, *refType, *refID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}
