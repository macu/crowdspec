package main

import (
	"database/sql"
	"fmt"
	"net/http"
)

func ajaxSpecLoadBlockCommunity(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// GET

	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("invalid specId: %w", err))
		return nil, http.StatusBadRequest
	}

	blockID, err := AtoInt64(query.Get("blockId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("invalid blockId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, err := verifyReadBlock(db, userID, specID, blockID); !access || err != nil {
		if err != nil {
			logError(r, userID, fmt.Errorf("validating read block access: %w", err))
			return nil, http.StatusInternalServerError
		}
		logError(r, userID,
			fmt.Errorf("read block access denied to user %d in spec %d", userID, specID))
		return nil, http.StatusForbidden
	}

	block, err := loadBlock(db, blockID)
	if err != nil {
		logError(r, userID, fmt.Errorf("loading block: %w", err))
		return nil, http.StatusInternalServerError
	}

	payload := struct {
		Block *SpecBlock `json:"block"`
	}{block}

	return payload, http.StatusOK
}

// func ajaxSpecBlockLoadComments(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
// }
//
// func ajaxSpecBlockAddComment(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
// }
//
// func ajaxSpecBlockUpdateComment(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
// }
//
// func ajaxSpecBlockCommentVote(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
// }
//
// func ajaxSpecBlockDeleteComment(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
// }
//
// func ajaxAdminSpecBlockCommentTag(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
//
// }
