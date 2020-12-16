package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// AjaxRoute represents an authenticated AJAX handler that returns
// a response object to be sent as JSON, or an error to log, and a status code.
type AjaxRoute func(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int)

var ajaxHandlers = map[string]map[string]AjaxRoute{
	http.MethodGet: map[string]AjaxRoute{
		"/ajax/test":     ajaxTest,
		"/ajax/home":     ajaxUserHome,
		"/ajax/settings": ajaxUserSettings,

		"/ajax/spec":                 ajaxSpec,
		"/ajax/spec/subspecs":        ajaxSubspecs,
		"/ajax/spec/subspec":         ajaxSubspec,
		"/ajax/spec/urls":            ajaxSpecURLs,
		"/ajax/spec/block-community": ajaxSpecLoadBlockCommunity,
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
			response, statusCode := handler(db, userID, w, r)
			if statusCode >= 400 {
				w.WriteHeader(statusCode)
				// Send current version stamp
				w.Write([]byte("VersionStamp: " + cacheControlVersionStamp))
				return
			}
			if response != nil {
				js, err := json.Marshal(response)
				if err != nil {
					logError(r, userID, fmt.Errorf("marshalling response: %w", err))
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

func ajaxTest(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	return struct {
		Message string `json:"message"`
	}{"Message retrieved using AJAX"}, http.StatusOK
}

func inTransaction(r *http.Request, db *sql.DB, userID uint, f func(*sql.Tx) (interface{}, int)) (interface{}, int) {
	c := r.Context()
	tx, err := db.BeginTx(c, nil)
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			logError(r, userID, fmt.Errorf("rollback: %v; on begin transaction: %w", rbErr, err))
			return nil, http.StatusInternalServerError
		}
		logError(r, userID, fmt.Errorf("begin transaction: %w", err))
		return nil, http.StatusInternalServerError
	}

	response, statusCode := f(tx)
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			logError(r, userID, fmt.Errorf("rollback: %v; on run function: %w", rbErr, err))
			return nil, http.StatusInternalServerError
		}
		logError(r, userID, fmt.Errorf("run function: %w", err))
		return nil, statusCode
	}

	err = tx.Commit()
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			logError(r, userID, fmt.Errorf("rollback: %v; on commit: %w", rbErr, err))
			return nil, http.StatusInternalServerError
		}
		logError(r, userID, fmt.Errorf("commit: %w", err))
		return nil, http.StatusInternalServerError
	}

	return response, statusCode
}
