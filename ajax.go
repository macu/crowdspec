package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// AjaxRoute represents an authenticated AJAX handler that returns
// a response object to be sent as JSON, or an error to log, and a status code.
type AjaxRoute func(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error)

var ajaxHandlers = map[string]map[string]AjaxRoute{
	http.MethodGet: map[string]AjaxRoute{
		"/ajax/test":          ajaxTest,
		"/ajax/home":          ajaxUserHome,
		"/ajax/settings":      ajaxUserSettings,
		"/ajax/spec":          ajaxSpec,
		"/ajax/spec/subspecs": ajaxSubspecs,
		"/ajax/spec/subspec":  ajaxSubspec,
		"/ajax/spec/urls":     ajaxSpecURLs,
		"/ajax/fetch-url":     ajaxFetchURLPreview,
	},
	http.MethodPost: map[string]AjaxRoute{
		"/ajax/user/change-password": ajaxUserChangePassword,
		"/ajax/user/save-settings":   ajaxUserSaveSettings,

		// specs
		"/ajax/spec/create-spec":    ajaxCreateSpec,
		"/ajax/spec/save-spec":      ajaxSaveSpec,
		"/ajax/spec/delete-spec":    ajaxDeleteSpec,
		"/ajax/spec/create-block":   ajaxSpecCreateBlock,
		"/ajax/spec/save-block":     ajaxSpecSaveBlock,
		"/ajax/spec/move-block":     ajaxSpecMoveBlock,
		"/ajax/spec/delete-block":   ajaxSpecDeleteBlock,
		"/ajax/spec/create-subspec": ajaxSpecCreateSubspec,
		"/ajax/spec/save-subspec":   ajaxSpecSaveSubspec,
		"/ajax/spec/delete-subspec": ajaxSpecDeleteSubspec,
		"/ajax/spec/create-url":     ajaxSpecCreateURL,
		"/ajax/spec/refresh-url":    ajaxSpecRefreshURL,
		"/ajax/spec/delete-url":     ajaxSpecDeleteURL,
	},
}

func ajaxHandler(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) {
	handlers, foundMethod := ajaxHandlers[r.Method]
	if foundMethod {
		handler, fouundPath := handlers[r.URL.Path]
		if fouundPath {
			response, statusCode, err := handler(db, userID, w, r)
			if err != nil {
				logError(r, userID, fmt.Errorf("[%s]: %w", r.URL.Path, err))
				w.WriteHeader(statusCode)
				// Send current version stamp
				w.Write([]byte("VersionStamp: " + cacheControlVersionStamp))
				return
			}
			if response != nil {
				js, err := json.Marshal(response)
				if err != nil {
					logError(r, userID, fmt.Errorf("[%s] marshalling response: %w", r.URL.Path, err))
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
	// return nil, http.StatusNotImplemented, fmt.Errorf("test error")
	return struct {
		Message string `json:"message"`
	}{"Message retrieved using AJAX"}, http.StatusOK, nil
}

func inTransaction(c context.Context, db *sql.DB, f func(tx *sql.Tx) (interface{}, int, error)) (interface{}, int, error) {
	tx, err := db.BeginTx(c, nil)
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("rollback: %v; on begin transaction: %w", rbErr, err)
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("begin transaction: %w", err)
	}

	response, statusCode, err := f(tx)
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("rollback: %v; on run function: %w", rbErr, err)
		}
		return nil, statusCode, fmt.Errorf("run function: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("rollback: %v; on commit: %w", rbErr, err)
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("commit: %w", err)
	}

	return response, statusCode, nil
}
