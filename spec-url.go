package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"image"
	"net/http"
	"net/url"
	"strings"
	"time"

	_ "image/jpeg"
	"image/png"

	m "github.com/keighl/metabolize" // (Custom license) https://github.com/keighl/metabolize/blob/master/LICENSE
	"github.com/nfnt/resize"         // (ISC license) https://github.com/nfnt/resize/blob/master/LICENSE
)

// URLObject contains display information about a URL.
type URLObject struct {
	ID        int64     `json:"id"`
	SpecID    int64     `json:"specId"`
	URL       string    `json:"url"`
	Title     *string   `json:"title,omitempty"`
	Desc      *string   `json:"desc,omitempty"`
	ImageData *string   `json:"imageData,omitempty"`
	UpdatedAt time.Time `json:"updated"`
}

// URLMetadata represents metadata extracted from a request to a URL.
type URLMetadata struct {
	Title        string  `meta:"og:title,twitter:title"`
	Description  string  `meta:"og:description,description,twitter:description"`
	CanonicalURL string  `meta:"og:url"`
	ImageURL     url.URL `meta:"og:image,twitter:image"`
	SiteName     string  `meta:"og:site_name"`
}

// URLMetadataTimeout is the max connection time when loading URL metadata.
const URLMetadataTimeout = 5 * time.Second

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

func createURLObject(tx *sql.Tx, specID int64, url string) (*URLObject, error) {
	data, err := fetchMetadata(url)
	if err != nil {
		return nil, fmt.Errorf("error loading url metadata: %w", err)
	}

	urlObject := &URLObject{
		SpecID: specID,
		URL:    url,
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
		if err == nil {
			// Silently ignore errors
			urlObject.ImageData = &imageData
		}
	}

	// Save URLObject
	err = tx.QueryRow(
		`INSERT INTO spec_url (spec_id, url, url_title, url_desc, url_image_data, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, url_title, url_desc, updated_at`,
		specID, url, urlObject.Title, urlObject.Desc, urlObject.ImageData, time.Now(),
	).Scan(&urlObject.ID, &urlObject.Title, &urlObject.Desc, &urlObject.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error inserting new spec_url: %w", err)
	}

	return urlObject, nil
}

func updateURLObject(tx *sql.Tx, id int64, url string) (*URLObject, error) {
	data, err := fetchMetadata(url)
	if err != nil {
		return nil, fmt.Errorf("error loading url metadata: %w", err)
	}

	urlObject := &URLObject{
		ID:  id,
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
		if err == nil {
			// Silently ignore errors
			urlObject.ImageData = &imageData
		}
	}

	// Save URLObject
	err = tx.QueryRow(
		`UPDATE spec_url SET url=$2, url_title=$3, url_desc=$4, url_image_data=$5, updated_at=$6
		WHERE id=$1 RETURNING url_title, url_desc, updated_at`,
		id, url, urlObject.Title, urlObject.Desc, urlObject.ImageData, time.Now(),
	).Scan(&urlObject.Title, &urlObject.Desc, &urlObject.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error updating spec_url: %w", err)
	}

	return urlObject, nil
}

func loadURLObject(db DBConn, id int64) (*URLObject, error) {
	var urlObject = &URLObject{
		ID: id,
	}

	err := db.QueryRow(
		`SELECT spec_id, url, url_title, url_desc, url_image_data, updated_at
		FROM spec_url WHERE id=$1`, id).Scan(&urlObject.SpecID, &urlObject.URL,
		&urlObject.Title, &urlObject.Desc, &urlObject.ImageData, &urlObject.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return urlObject, nil
}

func deleteURLObject(tx *sql.Tx, id int64) error {
	_, err := tx.Exec(`DELETE FROM spec_url WHERE id=$1`, id)
	return err
}
