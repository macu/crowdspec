package main

import (
	"fmt"
	"net/http"
	"strings"
)

// Current policy is that only the user who owns the spec may write to it.

func verifyReadSpec(r *http.Request, db DBConn, userID uint, specID int64) (bool, int) {

	var count uint
	err := db.QueryRow(
		`SELECT COUNT(*) FROM spec
		WHERE id = $1
			AND (is_public OR
				(owner_type = $2 AND owner_id = $3)
			)`,
		specID, OwnerTypeUser, userID).Scan(&count)

	if err != nil {
		logError(r, userID, fmt.Errorf("validating read spec access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, userID, fmt.Errorf("read spec access denied to user %d in spec %d", userID, specID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

func verifyWriteSpec(r *http.Request, db DBConn, userID uint, specID int64) (bool, int) {

	var count uint
	err := db.QueryRow(
		`SELECT COUNT(*) FROM spec
		WHERE id = $1
			AND owner_type = $2
			AND owner_id = $3`,
		specID, OwnerTypeUser, userID).Scan(&count)

	if err != nil {
		logError(r, userID, fmt.Errorf("validating write spec access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, userID, fmt.Errorf("write spec access denied to user %d in spec %d", userID, specID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

func verifyReadSubspec(r *http.Request, db DBConn, userID uint, subspecID int64) (bool, int) {

	var count uint
	err := db.QueryRow(
		`SELECT COUNT(*) FROM spec_subspec
		INNER JOIN spec
			ON spec.id = spec_subspec.spec_id
		WHERE spec_subspec.id = $1
			AND (spec.is_public OR
				(spec.owner_type = $2 AND spec.owner_id = $3)
			)`,
		subspecID, OwnerTypeUser, userID).Scan(&count)

	if err != nil {
		logError(r, userID, fmt.Errorf("validating read subspec access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, userID, fmt.Errorf("read subspec access denied to user %d in subspec %d", userID, subspecID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

func verifyWriteSubspec(r *http.Request, db DBConn, userID uint, subspecID int64) (bool, int) {

	var count uint
	err := db.QueryRow(
		`SELECT COUNT(*) FROM spec_subspec
		INNER JOIN spec
			ON spec.id = spec_subspec.spec_id
		WHERE spec_subspec.id = $1
			AND spec.owner_type = $2
			AND spec.owner_id = $3`,
		subspecID, OwnerTypeUser, userID).Scan(&count)

	if err != nil {
		logError(r, userID, fmt.Errorf("validating write subspec access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, userID, fmt.Errorf("write subspec access denied to user %d in subspec %d", userID, subspecID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
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
		urlID, OwnerTypeUser, userID).Scan(&count)

	if err != nil {
		logError(r, userID, fmt.Errorf("validating write url access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, userID, fmt.Errorf("write url access denied to user %d for url %d", userID, urlID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

func verifyReadBlock(r *http.Request, db DBConn, userID uint, specID, blockID int64) (bool, int) {

	var count uint
	err := db.QueryRow(
		`SELECT COUNT(*) FROM spec_block
		INNER JOIN spec
			ON spec.id = spec_block.spec_id
		WHERE
			spec_block.id = $2
			AND spec.id = $1
			AND (spec.is_public OR
				(spec.owner_type = $3 AND spec.owner_id = $4)
			)`,
		specID, blockID, OwnerTypeUser, userID).Scan(&count)

	if err != nil {
		logError(r, userID, fmt.Errorf("validating read block access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, userID, fmt.Errorf("read block access denied to user %d in spec %d", userID, specID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

func verifyWriteBlock(r *http.Request, db DBConn, userID uint, blockID int64) (bool, int) {

	var count uint
	err := db.QueryRow(
		`SELECT COUNT(*) FROM spec_block
		INNER JOIN spec ON spec.id = spec_block.spec_id
		WHERE spec_block.id = $1
		AND spec.owner_type = $2
		AND spec.owner_id = $3`,
		blockID, OwnerTypeUser, userID).Scan(&count)

	if err != nil {
		logError(r, userID, fmt.Errorf("validating write block access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, userID, fmt.Errorf("write block access denied to user %d to block %d", userID, blockID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

// - verifies the given block belongs to the given spec
func verifyWriteSpecBlock(r *http.Request, db DBConn, userID uint,
	specID int64, blockID int64) (bool, int) {
	return verifyWriteSpecSubspecBlocks(r, db, userID, specID, nil, blockID)
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
		logError(r, userID, fmt.Errorf("validating write blocks access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, userID, fmt.Errorf("write blocks access denied to user %d in spec %d", userID, specID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

// Current policy is that the ref item must belong to the spec it is referenced in.
func verifyRefAccess(r *http.Request, db DBConn, userID uint,
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
		logError(r, userID, fmt.Errorf("unsupported refType: %s", *refType))
		return false, http.StatusBadRequest
	}

	if err != nil {
		logError(r, userID, fmt.Errorf("validating read ref access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, userID,
			fmt.Errorf("read ref access denied to user %d in spec %d, refType %s refID %d",
				userID, specID, *refType, *refID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

// Community access

func verifyAddComment(r *http.Request, db DBConn, userID uint,
	specID int64, targetType string, targetID int64) (bool, int) {

	var count uint
	var err error

	var queryCount = func(tableName string) {
		err = db.QueryRow(
			`SELECT COUNT(*) FROM `+tableName+`
				INNER JOIN spec
					ON spec.id = `+tableName+`.spec_id
				WHERE
					`+tableName+`.id = $2
					AND spec.id = $1 -- verify spec association
					AND (spec.is_public
						OR (spec.owner_type = $3 AND spec.owner_id = $4) -- allow spec owner
					)`,
			specID, targetID, OwnerTypeUser, userID).Scan(&count)
	}

	switch targetType {

	case "spec":
		err = db.QueryRow(
			`SELECT COUNT(*) FROM spec
				WHERE
					spec.id = $1
					AND (spec.is_public
						OR (spec.owner_type = $2 AND spec.owner_id = $3) -- allow spec owner
					)`,
			specID, OwnerTypeUser, userID).Scan(&count)

	case "subspec":
		queryCount("spec_subspec")

	case "block":
		queryCount("spec_block")

	case "comment":
		queryCount("spec_community_comment")

	default:
		return false, http.StatusBadRequest

	}

	if err != nil {
		logError(r, userID, fmt.Errorf("validating add comment access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, userID, fmt.Errorf("add comment access denied to user %d in spec %d", userID, specID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

// allow all in public specs
// allow comment author or spec owner in private specs
func verifyReadComment(r *http.Request, db DBConn, userID uint, specID, commentID int64) (bool, int) {

	var count uint
	err := db.QueryRow(
		`SELECT COUNT(*)
		FROM spec_community_comment AS c
		INNER JOIN spec
			ON spec.id = c.spec_id
		WHERE
			c.id = $2
			AND spec.id = $1 -- verify spec association
			AND (spec.is_public
				OR (spec.owner_type = $3 AND spec.owner_id = $4) -- allow spec owner
				OR c.user_id = $4 -- allow comment author
			)`,
		specID, commentID, OwnerTypeUser, userID).Scan(&count)

	if err != nil {
		logError(r, userID, fmt.Errorf("validating read comment access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, userID, fmt.Errorf("read comment access denied to user %d in spec %d", userID, specID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

// allow comment author only
func verifyUpdateComment(r *http.Request, db DBConn, userID uint, specID, commentID int64) (bool, int) {

	var count uint
	err := db.QueryRow(
		`SELECT COUNT(*)
		FROM spec_community_comment AS c
		INNER JOIN spec
			ON spec.id = c.spec_id
		WHERE
			c.id = $2
			AND spec.id = $1 -- verify spec association
			AND c.user_id = $3 -- allow comment author
		`, specID, commentID, userID).Scan(&count)

	if err != nil {
		logError(r, userID, fmt.Errorf("validating write comment access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, userID, fmt.Errorf("write comment access denied to user %d in spec %d", userID, specID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

// allow comment author or spec owner
func verifyDeleteComment(r *http.Request, db DBConn, userID uint, specID, commentID int64) (bool, int) {

	var count uint
	err := db.QueryRow(
		`SELECT COUNT(*)
		FROM spec_community_comment AS c
		INNER JOIN spec
			ON spec.id = c.spec_id
		WHERE
			c.id = $2
			AND c.spec_id = $1 -- verify spec association
			AND (
				(spec.owner_type = $3 AND spec.owner_id = $4) -- allow spec owner
				OR c.user_id = $4 -- allow comment author
			)
		`, specID, commentID, OwnerTypeUser, userID,
	).Scan(&count)

	if err != nil {
		logError(r, userID, fmt.Errorf("validating delete comment access: %w", err))
		return false, http.StatusInternalServerError
	}

	if count == 0 {
		logError(r, userID, fmt.Errorf("delete comment access denied to user %d in spec %d", userID, specID))
		return false, http.StatusForbidden
	}

	return true, http.StatusOK
}

func verifyReadCommunityTarget(r *http.Request, db DBConn, userID uint,
	specID int64, targetType string, targetID int64) (bool, int) {
	switch targetType {
	case CommunityTargetSpec:
		return verifyReadSpec(r, db, userID, specID)
	case CommunityTargetSubspec:
		return verifyReadSubspec(r, db, userID, targetID)
	case CommunityTargetBlock:
		return verifyReadBlock(r, db, userID, specID, targetID)
	case CommunityTargetComment:
		return verifyReadComment(r, db, userID, specID, targetID)
	default:
		logError(r, userID, fmt.Errorf("unrecognized target type: %s", targetType))
		return false, http.StatusBadRequest
	}
}
