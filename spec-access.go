package main

import (
	"fmt"
	"strings"
)

// Current policy is that only the user who owns the spec may write to it.

func verifyReadSpec(db DBConn, userID uint, specID int64) (bool, error) {

	var count uint
	err := db.QueryRow(`
		SELECT COUNT(*) FROM spec
		WHERE id = $1
		AND (is_public OR (
			owner_type = $2 AND owner_id = $3
		))`, specID, OwnerTypeUser, userID).Scan(&count)

	return err == nil && count > 0, err
}

func verifyWriteSpec(db DBConn, userID uint, specID int64) (bool, error) {

	var count uint
	err := db.QueryRow(`
		SELECT COUNT(*) FROM spec
		WHERE id = $1
		AND owner_type = $2
		AND owner_id = $3
		`, specID, OwnerTypeUser, userID).Scan(&count)

	return err == nil && count > 0, err
}

func verifyReadSubspec(db DBConn, userID uint, subspecID int64) (bool, error) {

	var count uint
	err := db.QueryRow(`
		SELECT COUNT(*) FROM spec_subspec
		INNER JOIN spec ON spec.id = spec_subspec.spec_id
		WHERE spec_subspec.id = $1
		AND (spec.is_public OR (
			spec.owner_type = $2 AND spec.owner_id = $3
		))`, subspecID, OwnerTypeUser, userID).Scan(&count)

	return err == nil && count > 0, err
}

func verifyWriteSubspec(db DBConn, userID uint, subspecID int64) (bool, error) {

	var count uint
	err := db.QueryRow(`
		SELECT COUNT(*) FROM spec_subspec
		INNER JOIN spec ON spec.id = spec_subspec.spec_id
		WHERE spec_subspec.id = $1
		AND spec.owner_type = $2
		AND spec.owner_id = $3
		`, subspecID, OwnerTypeUser, userID).Scan(&count)

	return err == nil && count > 0, err
}

func verifyWriteURL(db DBConn, userID uint, urlID int64) (bool, error) {

	var count uint
	err := db.QueryRow(`
		SELECT COUNT(*) FROM spec_url
		INNER JOIN spec ON spec.id = spec_url.spec_id
		WHERE spec_url.id = $1
		AND spec.owner_type = $2
		AND spec.owner_id = $3
		`, urlID, OwnerTypeUser, userID).Scan(&count)

	return err == nil && count > 0, err
}

func verifyReadBlock(db DBConn, userID uint, blockID int64) (bool, error) {

	var count uint
	err := db.QueryRow(`
		SELECT COUNT(*) FROM spec_block
		INNER JOIN spec ON spec.id = spec_block.spec_id
		WHERE spec_block.id = $1
		AND (spec.is_public OR (
			spec.owner_type = $2 AND spec.owner_id = $3
		))
		`, blockID, OwnerTypeUser, userID).Scan(&count)

	return err == nil && count > 0, err
}

func verifyWriteBlock(db DBConn, userID uint, blockID int64) (bool, error) {

	var count uint
	err := db.QueryRow(`
		SELECT COUNT(*) FROM spec_block
		INNER JOIN spec ON spec.id = spec_block.spec_id
		WHERE spec_block.id = $1
		AND spec.owner_type = $2
		AND spec.owner_id = $3
		`, blockID, OwnerTypeUser, userID).Scan(&count)

	return err == nil && count > 0, err
}

// - verifies the given block belongs to the given spec
func verifyWriteSpecBlock(db DBConn, userID uint, specID int64, blockID int64) (bool, error) {
	return verifyWriteSpecSubspecBlocks(db, userID, specID, nil, &blockID)
}

// - verifies subspec belongs to spec if given
// - verifies all given blocks belong to spec
func verifyWriteSpecSubspecBlocks(db DBConn, userID uint, specID int64, subspecID *int64, blockIDs ...*int64) (bool, error) {

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
		if id != nil {
			tableName := "block_" + IntToA(i)
			blockJoins = append(blockJoins, `
				INNER JOIN spec_block AS `+tableName+`
				ON `+tableName+`.id = `+argPlaceholder(*id, &args)+`
				AND `+tableName+`.spec_id = spec.id
				`)
		}
	}

	var count uint
	err := db.QueryRow(`
		SELECT COUNT(*) FROM spec`+
		subspecJoin+
		strings.Join(blockJoins, "")+`
		WHERE spec.owner_type = `+argPlaceholder(OwnerTypeUser, &args)+`
		AND spec.owner_id = `+argPlaceholder(userID, &args),
		args...).Scan(&count)

	// TODO return detailed error message about what aspect doesn't check out

	return err == nil && count > 0, err
}

// Current policy is that the ref item must belong to the spec it is referenced in.
func verifyRefAccess(db DBConn, specID int64, refType *string, refID *int64) (bool, error) {

	if refType == nil || refID == nil {
		return true, nil
	}

	var count uint
	var err error

	switch *refType {

	case BlockRefURL:
		err = db.QueryRow(`
			SELECT COUNT(*) FROM spec_url
			WHERE id = $1 AND spec_id = $2
			`, *refID, specID).Scan(&count)

	case BlockRefSubspec:
		err = db.QueryRow(`
			SELECT COUNT(*) FROM spec_subspec
			WHERE id = $1 AND spec_id = $2
			`, *refID, specID).Scan(&count)

	default:
		return false, fmt.Errorf("unsupported refType: %s", *refType)
	}

	return err == nil && count > 0, err
}
