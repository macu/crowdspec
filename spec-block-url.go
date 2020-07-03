package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"image"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	_ "image/jpeg"
	"image/png"

	m "github.com/keighl/metabolize" // (Custom license) https://github.com/keighl/metabolize/blob/master/LICENSE
	"github.com/nfnt/resize"         // (ISC license) https://github.com/nfnt/resize/blob/master/LICENSE
)

// URLMetadataTimeout is the max connection time when loading URL metadata.
const URLMetadataTimeout = 5 * time.Second

// URLObject contains display information about a URL.
type URLObject struct {
	ID        int64   `json:"id"`
	BlockID   int64   `json:"blockId"`
	URL       string  `json:"url"`
	Title     *string `json:"title,omitempty"`
	Desc      *string `json:"desc,omitempty"`
	ImageData *string `json:"imageData,omitempty"`
}

// URLMetadata represents metadata extracted from a request to a URL.
type URLMetadata struct {
	Title        string  `meta:"og:title,twitter:title"`
	Description  string  `meta:"og:description,description,twitter:description"`
	CanonicalURL string  `meta:"og:url"`
	ImageURL     url.URL `meta:"og:image,twitter:image"`
	SiteName     string  `meta:"og:site_name"`
}

func fetchMetadata(url string) (*URLMetadata, error) {
	client := http.Client{
		Timeout: URLMetadataTimeout,
	}
	res, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching URL: %w", err)
	}

	defer res.Body.Close()

	data := &URLMetadata{}

	err = m.Metabolize(res.Body, data)
	if err != nil {
		return nil, fmt.Errorf("error reading meta tags: %w", err)
	}

	return data, nil
}

func loadImageThumbData(imageURL string) (string, error) {
	client := http.Client{
		Timeout: URLMetadataTimeout,
	}
	res, err := client.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("error fetching URL: %w", err)
	}

	defer res.Body.Close()

	// Load image
	image, _, err := image.Decode(res.Body)
	if err != nil {
		return "", fmt.Errorf("error decoding image: %w", err)
	}

	thumb := resize.Thumbnail(300, 300, image, resize.Lanczos3)

	stringBuilder := new(strings.Builder)
	base64Writer := base64.NewEncoder(base64.StdEncoding, stringBuilder)

	err = png.Encode(base64Writer, thumb)
	if err != nil {
		return "", fmt.Errorf("error encoding thumbnail: %w", err)
	}

	return "data:image/png;base64," + stringBuilder.String(), nil
}

func createURLObject(tx *sql.Tx, blockID int64, url string) (*URLObject, error) {
	data, err := fetchMetadata(url)
	if err != nil {
		return nil, fmt.Errorf("error loading url metadata: %w", err)
	}

	urlObject := &URLObject{
		URL: url,
	}

	if strings.TrimSpace(data.Title) != "" {
		title := strings.TrimSpace(data.Title)
		urlObject.Title = &title
	}

	if strings.TrimSpace(data.Description) != "" {
		desc := strings.TrimSpace(data.Description)
		urlObject.Desc = &desc
	}

	if data.ImageURL.Host != "" {
		imageData, err := loadImageThumbData(data.ImageURL.String())
		if err != nil {
			// Continue but log
			log.Println(fmt.Errorf("error loading url image thumbnail: %w", err))
		} else {
			urlObject.ImageData = &imageData
		}
	}

	// Save URLObject
	err = tx.QueryRow(
		`INSERT INTO spec_block_url (block_id, url, url_title, url_desc, url_image_data)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, url_title, url_desc`,
		blockID, url, urlObject.Title, urlObject.Desc, urlObject.ImageData,
	).Scan(&urlObject.ID, &urlObject.Title, &urlObject.Desc)
	if err != nil {
		return nil, fmt.Errorf("error inserting new spec_block_url: %w", err)
	}

	return urlObject, nil
}

func updateURLObject(tx *sql.Tx, refID int64, url string) (interface{}, error) {
	data, err := fetchMetadata(url)
	if err != nil {
		return nil, fmt.Errorf("error loading url metadata: %w", err)
	}

	urlObject := &URLObject{
		ID:  refID,
		URL: url,
	}

	if strings.TrimSpace(data.Title) != "" {
		title := strings.TrimSpace(data.Title)
		urlObject.Title = &title
	}

	if strings.TrimSpace(data.Description) != "" {
		desc := strings.TrimSpace(data.Description)
		urlObject.Desc = &desc
	}

	if data.ImageURL.Host != "" {
		imageData, err := loadImageThumbData(data.ImageURL.String())
		if err != nil {
			// Continue but log
			log.Println(fmt.Errorf("error loading url image thumbnail: %w", err))
		} else {
			urlObject.ImageData = &imageData
		}
	}

	// Save URLObject
	err = tx.QueryRow(
		`UPDATE spec_block_url SET url=$1, url_title=$2, url_desc=$3, url_image_data=$4
		WHERE id=$5 RETURNING url_title, url_desc`,
		url, urlObject.Title, urlObject.Desc, urlObject.ImageData, refID,
	).Scan(&urlObject.Title, &urlObject.Desc)
	if err != nil {
		return nil, fmt.Errorf("error updating spec_block_url: %w", err)
	}

	return urlObject, nil
}

func loadURLObject(tx *sql.Tx, refID int64) (*URLObject, error) {
	var urlObject = &URLObject{
		ID: refID,
	}

	err := tx.QueryRow(
		`SELECT url, url_title, url_desc, url_image_data
		FROM spec_block_url WHERE id=$1`, refID).Scan(&urlObject.URL,
		&urlObject.Title, &urlObject.Desc, &urlObject.ImageData)
	if err != nil {
		return nil, err
	}

	return urlObject, nil
}

func deleteURLObject(tx *sql.Tx, refID int64) error {
	_, err := tx.Exec(`DELETE FROM spec_block_url WHERE id=$1`, refID)
	return err
}

// Returns a URL preview.
func ajaxFetchURLObject(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	// GET

	err := r.ParseForm()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	url := r.Form.Get("url")

	if url == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("url required")
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
		if err != nil {
			// Continue but log
			log.Println(fmt.Errorf("error loading url image thumbnail: %w", err))
		} else {
			urlObject.ImageData = &imageData
		}
	}

	return urlObject, http.StatusOK, nil
}
