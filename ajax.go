package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

// AjaxRoute represents an authenticated AJAX handler that returns
// a response object to be sent as JSON, or an error to log, and a status code.
type AjaxRoute func(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error)

var ajaxHandlers = map[string]map[string]AjaxRoute{
	http.MethodGet: map[string]AjaxRoute{
		"/ajax/test":       ajaxTest,
		"/ajax/user-specs": ajaxUserSpecs,
		"/ajax/spec":       ajaxSpec,
		// "/ajax/fetch-url":  ajaxFetchURLObject,
	},
	http.MethodPost: map[string]AjaxRoute{
		"/ajax/spec/create-spec":     ajaxCreateSpec,
		"/ajax/spec/save-spec":       ajaxSaveSpec,
		"/ajax/spec/delete-spec":     ajaxDeleteSpec,
		"/ajax/spec/create-block":    ajaxSpecCreateBlock,
		"/ajax/spec/save-block":      ajaxSpecSaveBlock,
		"/ajax/spec/move-block":      ajaxSpecMoveBlock,
		"/ajax/spec/delete-block":    ajaxSpecDeleteBlock,
		"/ajax/spec/create-subspace": ajaxSpecCreateSubspace,
		"/ajax/spec/save-subspace":   ajaxSpecSaveSubspace,
		"/ajax/spec/delete-subspace": ajaxSpecDeleteSubspace,
	},
}

func ajaxHandler(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) {
	handlers, foundMethod := ajaxHandlers[r.Method]
	if foundMethod {
		handler, fouundPath := handlers[r.URL.Path]
		if fouundPath {
			response, statusCode, err := handler(db, userID, w, r)
			if err != nil {
				log.Printf("Error running ajax handler [%s]: %v\n", r.URL.Path, err)
				w.WriteHeader(statusCode)
				// Send current version stamp
				w.Write([]byte("VersionStamp: " + cacheControlVersionStamp))
				return
			}
			if response != nil {
				js, err := json.Marshal(response)
				if err != nil {
					log.Printf("Error marshalling response: %v\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(statusCode) // WriteHeader is called after setting headers
				w.Write(js)
			} else {
				w.WriteHeader(statusCode)
			}
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func ajaxTest(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	return struct {
		Message string `json:"message"`
	}{"Message retrieved using AJAX"}, http.StatusOK, nil
}

func inTransaction(c context.Context, db *sql.DB, f func(tx *sql.Tx) (interface{}, int, error)) (interface{}, int, error) {
	tx, err := db.BeginTx(c, nil)
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			log.Println(rbErr)
		}
		return nil, http.StatusInternalServerError, err
	}

	response, statusCode, err := f(tx)
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			log.Println(rbErr)
		}
		return nil, statusCode, err
	}

	err = tx.Commit()
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			log.Println(rbErr)
		}
		return nil, http.StatusInternalServerError, err
	}

	return response, statusCode, nil
}
