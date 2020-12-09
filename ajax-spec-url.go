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

func ajaxSpecURLs(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET
	query := r.URL.Query()

	specID, err := AtoInt64(query.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid specId: %w", err)
	}

	if access, err := verifyReadSpec(db, userID, specID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying read spec: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("read spec access denied to user %d in spec %d", userID, specID)
	}

	rows, err := db.Query(`
		SELECT id, created_at, url, url_title, url_desc, url_image_data, updated_at
		FROM spec_url
		WHERE spec_id = $1
		ORDER BY url_title, url`, specID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("querying links: %w", err)
	}

	links := []*URLObject{}

	for rows.Next() {
		o := &URLObject{
			SpecID: specID,
		}
		err = rows.Scan(&o.ID, &o.Created, &o.URL, &o.Title, &o.Desc, &o.ImageData, &o.Updated)
		if err != nil {
			if err2 := rows.Close(); err2 != nil { // TODO Add everywhere
				return nil, http.StatusInternalServerError, fmt.Errorf("error closing rows: %s; on scan error: %w", err2, err)
			}
			return nil, http.StatusInternalServerError, fmt.Errorf("error scanning spec_url: %w", err)
		}
		links = append(links, o)
	}

	return links, http.StatusOK, nil
}

func ajaxSpecCreateURL(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	specID, err := AtoInt64(r.Form.Get("specId"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing specId: %w", err)
	}

	if access, err := verifyWriteSpec(db, userID, specID); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying write spec: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("write spec access denied to user %d in spec %d", userID, specID)
	}

	url := strings.TrimSpace(r.Form.Get("url"))
	if url == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("url required")
	}
	if err = validateURL(url); err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid url: %w", err)
	}

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		urlObject, err := createURLObject(tx, specID, url)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("creating spec_url: %w", err)
		}

		return urlObject, http.StatusOK, nil
	})
}

func ajaxSpecRefreshURL(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	id, err := AtoInt64(r.Form.Get("id"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing id: %w", err)
	}

	if access, err := verifyWriteURL(db, userID, id); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying write url: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("write url access denied to user %d for url %d", userID, id)
	}

	url := strings.TrimSpace(r.Form.Get("url"))
	if url == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("url required")
	}
	if err = validateURL(url); err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid url: %w", err)
	}

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		urlObject, err := updateURLObject(tx, id, url)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("updating spec_url: %w", err)
		}

		return urlObject, http.StatusOK, nil
	})
}

func ajaxSpecDeleteURL(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// POST

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	id, err := AtoInt64(r.Form.Get("id"))
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("parsing id: %w", err)
	}

	if access, err := verifyWriteURL(db, userID, id); !access || err != nil {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("verifying write url: %w", err)
		}
		return nil, http.StatusForbidden,
			fmt.Errorf("write url access denied to user %d for url %d", userID, id)
	}

	return inTransaction(r.Context(), db, func(tx *sql.Tx) (interface{}, int, error) {

		// Leave block references

		_, err = tx.Exec(`
				DELETE FROM spec_url
				WHERE id=$1
				`, id)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("deleting spec_url: %w", err)
		}

		return nil, http.StatusOK, nil
	})
}

// Returns a URL preview.
func ajaxFetchURLPreview(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	url := strings.TrimSpace(r.Form.Get("url"))
	if url == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("url required")
	}
	if err = validateURL(url); err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid url: %w", err)
	}

	data, err := fetchMetadata(url)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("error loading url metadata: %w", err)
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

	return urlObject, http.StatusOK, nil
}