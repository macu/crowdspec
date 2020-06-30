package main

import (
	"database/sql"
	"net/http"
)

func ajaxSpecCreateSubspace(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	return nil, http.StatusNotImplemented, nil
}

func ajaxSpecSaveSubspace(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	return nil, http.StatusNotImplemented, nil
}

func ajaxSpecDeleteSubspace(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	return nil, http.StatusNotImplemented, nil
}
