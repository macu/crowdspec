package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func validateURL(s string) error {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return fmt.Errorf("parsing URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
	return nil
}

func ajaxSpecURLs(db *sql.DB, userID *uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// GET
	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		logError(r, userID, fmt.Errorf("parsing specId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyReadSpec(r, db, userID, specID); !access {
		return nil, status
	}

	rows, err := db.Query(`
		SELECT id, created_at, url, url_title, url_desc, url_image_data, updated_at
		FROM spec_url
		WHERE spec_id = $1
		ORDER BY url_title, url`, specID)
	if err != nil {
		logError(r, userID, fmt.Errorf("querying links: %w", err))
		return nil, http.StatusInternalServerError
	}

	links := []*URLObject{}

	for rows.Next() {
		o := &URLObject{
			SpecID: specID,
		}
		err = rows.Scan(&o.ID, &o.Created, &o.URL, &o.Title, &o.Desc, &o.ImageData, &o.Updated)
		if err != nil {
			if err2 := rows.Close(); err2 != nil {
				logError(r, userID, fmt.Errorf("closing rows: %s; on scan error: %w", err2, err))
				return nil, http.StatusInternalServerError
			}
			logError(r, userID, fmt.Errorf("scanning spec_url: %w", err))
			return nil, http.StatusInternalServerError
		}
		links = append(links, o)
	}

	return links, http.StatusOK
}

func ajaxSpecCreateURL(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	err := r.ParseForm()
	if err != nil {
		logError(r, &userID, err)
		return nil, http.StatusInternalServerError
	}

	specID, err := AtoInt64(r.Form.Get("specId"))
	if err != nil {
		logError(r, &userID, fmt.Errorf("parsing specId: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyWriteSpec(r, db, &userID, specID); !access {
		return nil, status
	}

	url := strings.TrimSpace(r.Form.Get("url"))
	if url == "" {
		logError(r, &userID, fmt.Errorf("url required"))
		return nil, http.StatusBadRequest
	}
	if err = validateURL(url); err != nil {
		logError(r, &userID, fmt.Errorf("invalid url: %w", err))
		return nil, http.StatusBadRequest
	}

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		urlObject, err := createURLObject(tx, specID, url)

		if err != nil {
			logError(r, &userID, fmt.Errorf("creating spec_url: %w", err))
			return nil, http.StatusInternalServerError
		}

		return urlObject, http.StatusOK
	})
}

func ajaxSpecRefreshURL(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	err := r.ParseForm()
	if err != nil {
		logError(r, &userID, err)
		return nil, http.StatusInternalServerError
	}

	id, err := AtoInt64(r.Form.Get("id"))
	if err != nil {
		logError(r, &userID, fmt.Errorf("parsing id: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyWriteURL(r, db, &userID, id); !access {
		return nil, status
	}

	url := strings.TrimSpace(r.Form.Get("url"))
	if url == "" {
		logError(r, &userID, fmt.Errorf("url required"))
		return nil, http.StatusBadRequest
	}
	if err = validateURL(url); err != nil {
		logError(r, &userID, fmt.Errorf("invalid url: %w", err))
		return nil, http.StatusBadRequest
	}

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		urlObject, err := updateURLObject(tx, id, url)

		if err != nil {
			logError(r, &userID, fmt.Errorf("updating spec_url: %w", err))
			return nil, http.StatusInternalServerError
		}

		return urlObject, http.StatusOK
	})
}

func ajaxSpecDeleteURL(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// POST

	err := r.ParseForm()
	if err != nil {
		logError(r, &userID, err)
		return nil, http.StatusInternalServerError
	}

	id, err := AtoInt64(r.Form.Get("id"))
	if err != nil {
		logError(r, &userID, fmt.Errorf("parsing id: %w", err))
		return nil, http.StatusBadRequest
	}

	if access, status := verifyWriteURL(r, db, &userID, id); !access {
		return nil, status
	}

	return handleInTransaction(r, db, &userID, func(tx *sql.Tx) (interface{}, int) {

		// Don't clear references from blocks - display "content unavailable" message

		_, err = tx.Exec(`
				DELETE FROM spec_url
				WHERE id=$1
				`, id)

		if err != nil {
			logError(r, &userID, fmt.Errorf("deleting spec_url: %w", err))
			return nil, http.StatusInternalServerError
		}

		return nil, http.StatusOK
	})
}

// Returns a URL preview.
func ajaxFetchURLPreview(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int) {
	// GET

	err := r.ParseForm()
	if err != nil {
		logError(r, &userID, err)
		return nil, http.StatusInternalServerError
	}

	url := strings.TrimSpace(r.Form.Get("url"))
	if url == "" {
		logError(r, &userID, fmt.Errorf("url required"))
		return nil, http.StatusBadRequest
	}
	if err = validateURL(url); err != nil {
		logError(r, &userID, fmt.Errorf("invalid url: %w", err))
		return nil, http.StatusBadRequest
	}

	data, err := fetchMetadata(url)
	if err != nil {
		logError(r, &userID, fmt.Errorf("loading url metadata: %w", err))
		return nil, http.StatusInternalServerError
	}

	urlObject := &URLObject{}

	if data.Title != "" {
		urlObject.Title = &data.Title
	}

	if data.Description != "" {
		urlObject.Desc = &data.Description
	}

	if data.CanonicalURL != "" {
		urlObject.URL = data.CanonicalURL
	} else {
		urlObject.URL = url
	}

	if data.ImageURL.Host != "" {
		imageData, err := loadImageThumbData(data.ImageURL.String())
		if err == nil {
			// Silently ignore errors
			urlObject.ImageData = &imageData
		}
	}

	return urlObject, http.StatusOK
}
