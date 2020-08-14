package main

import (
	"database/sql"
	"fmt"
	"net/http"
)

func makeCronHandler(db *sql.DB, handler func(*sql.DB) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// X-Appengine-Cron is added by App Engine; if a client sends this header, it is removed.
		if r.Header.Get("X-Appengine-Cron") == "" {
			logError(r, 0, fmt.Errorf("illegal request to cron"))
			w.WriteHeader(http.StatusForbidden)
		} else {
			if err := handler(db); err != nil {
				logError(r, 0, fmt.Errorf("error running cron handler: %w", err))
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}
	}
}

func cleanupHandler(db *sql.DB) error {
	if err := deleteExpiredSessions(db); err != nil {
		return err
	}
	return nil
}
