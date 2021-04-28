package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// AjaxRoute represents an authenticated AJAX handler that returns
// a response object to be sent as JSON, or an error to log, and a status code.
type AjaxRoute func(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int)

var ajaxHandlers = map[string]map[string]AjaxRoute{
	http.MethodGet: {
		"/ajax/test":     ajaxTest,
		"/ajax/home":     ajaxUserHome,
		"/ajax/settings": ajaxUserSettings,

		"/ajax/spec":                 ajaxSpec,
		"/ajax/spec/subspecs":        ajaxSubspecs,
		"/ajax/spec/subspec":         ajaxSubspec,
		"/ajax/spec/urls":            ajaxSpecURLs,
		"/ajax/spec/community":       ajaxSpecLoadCommunity,
		"/ajax/spec/community/page":  ajaxSpecCommunityLoadCommentsPage,
		"/ajax/spec/load-edit-block": ajaxLoadBlockForEditing,
		"/ajax/community-review":     ajaxLoadCommuntyReviewPage,

		// admin
		"/ajax/admin/signup-requests": ajaxAdminLoadSignupRequests,
		"/ajax/admin/users":           ajaxAdminLoadUsers,
	},
	http.MethodPost: {
		"/ajax/user/change-password": ajaxUserChangePassword,
		"/ajax/user/save-settings":   ajaxUserSaveSettings,

		// specs
		"/ajax/spec/create-spec":    ajaxCreateSpec,
		"/ajax/spec/save-spec":      ajaxSaveSpec,
		"/ajax/spec/delete-spec":    ajaxDeleteSpec,
		"/ajax/spec/create-block":   ajaxSpecCreateBlock,
		"/ajax/spec/save-block":     ajaxSpecSaveBlock,
		"/ajax/spec/move-blocks":    ajaxSpecMoveBlocks,
		"/ajax/spec/delete-block":   ajaxSpecDeleteBlock,
		"/ajax/spec/create-subspec": ajaxSpecCreateSubspec,
		"/ajax/spec/save-subspec":   ajaxSpecSaveSubspec,
		"/ajax/spec/delete-subspec": ajaxSpecDeleteSubspec,
		"/ajax/spec/create-url":     ajaxSpecCreateURL,
		"/ajax/spec/refresh-url":    ajaxSpecRefreshURL,
		"/ajax/spec/delete-url":     ajaxSpecDeleteURL,

		"/ajax/spec/community/mark-read":      ajaxSpecCommunityMarkRead,
		"/ajax/spec/community/add-comment":    ajaxSpecCommunityAddComment,
		"/ajax/spec/community/update-comment": ajaxSpecCommunityUpdateComment,
		"/ajax/spec/community/delete-comment": ajaxSpecCommunityDeleteComment,

		// util
		"/ajax/render-markdown": ajaxRenderMarkdown,

		// admin
		"/ajax/admin/review-signup": ajaxAdminSubmitSignupRequestReview,
	},
}

func ajaxHandler(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) {
	// var rt = NewResponseTracker(w)
	handlers, foundMethod := ajaxHandlers[r.Method]
	if foundMethod {
		handler, fouundPath := handlers[r.URL.Path]
		if fouundPath {
			// Verify access to admin routes
			if strings.HasPrefix(r.URL.Path, "/ajax/admin") && userID != adminUserID {
				logError(r, userID, fmt.Errorf("forbidden admin access"))
				w.WriteHeader(http.StatusForbidden)
				return
			}
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
