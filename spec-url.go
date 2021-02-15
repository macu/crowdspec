package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	_ "image/jpeg"
	"image/png"

	m "github.com/keighl/metabolize" // (Custom license) https://github.com/keighl/metabolize/blob/master/LICENSE
	"github.com/nfnt/resize"         // (ISC license) https://github.com/nfnt/resize/blob/master/LICENSE
)

const urlMaxLen = 1024
const urlTitleMaxLen = 255
const urlDescMaxLen = 255
const maxThumbnailDims = 300

// URLObject contains display information about a URL.
type URLObject struct {
	ID      int64     `json:"id"`
	SpecID  int64     `json:"specId"`
	Created time.Time `json:"created"`
	URL     string    `json:"url"`
	// Note on omitempty: https://play.golang.org/p/Lk_FdWeL4i8
	// empty non-nil values are not omitted
	Title     *string   `json:"title,omitempty"`
	Desc      *string   `json:"desc,omitempty"`
	ImageData *string   `json:"imageData,omitempty"`
	Updated   time.Time `json:"updated"`
}

// URLMetadata represents metadata extracted from a request to a URL.
type URLMetadata struct {
	Title        string  `meta:"og:title,twitter:title"`
	Description  string  `meta:"og:description,description,twitter:description"`
	CanonicalURL string  `meta:"og:url"`
	ImageURL     url.URL `meta:"og:image,twitter:image"`
	SiteName     string  `meta:"og:site_name"`
}

// YouTubeVideoListResponse represents a response from YouTube API.
type YouTubeVideoListResponse struct {
	Items []YouTubeVideoMetadata `json:"items"`
}

// YouTubeVideoMetadata represents data about a video in a response from YouTube API.
type YouTubeVideoMetadata struct {
	Snippet *YouTubeVideoSnippet `json:"snippet"`
}

// YouTubeVideoSnippet represents snippet data about a video in a response from YouTube API.
type YouTubeVideoSnippet struct {
	Title       string                           `json:"title"`
	Description string                           `json:"description"`
	Thumbnails  map[string]YouTubeVideoThumbnail `json:"thumbnails"`
}

// YouTubeVideoThumbnail represents thumbnail data about a video in a response from YouTube API.
type YouTubeVideoThumbnail struct {
	URL    string `json:"url"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
}

// URLMetadataTimeout is the max connection time when loading URL metadata.
const URLMetadataTimeout = 5 * time.Second

// Recognized video URL formats:
// http://www.youtube.com/watch?v=My2FRPA3Gf8
// http://www.youtube.com/embed/My2FRPA3Gf8
// http://youtu.be/My2FRPA3Gf8
// https://youtube.googleapis.com/v/My2FRPA3Gf8
var youTubeLinkRegex = regexp.MustCompile("^https?:\\/\\/(?:www\\.|m\\.)?(?:youtube\\.com\\/watch\\?v=|youtube\\.com\\/embed\\/|youtu\\.be\\/|youtube\\.googleapis\\.com\\/v\\/)([a-zA-Z0-9_-]+)")

func fetchMetadata(url string) (*URLMetadata, error) {
	if m := youTubeLinkRegex.FindStringSubmatch(url); len(m) == 2 {
		// Load YouTube metadata through their API
		return fetchYouTubeVideoMetadata(m[1])
	}

	client := http.Client{
		Timeout: URLMetadataTimeout,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("initiating HTTP request: %w", err)
	}

	req.Header.Add("Referer", httpClientReferer)

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching URL: %w", err)
	}

	defer res.Body.Close()

	data := &URLMetadata{}

	// metabolize reads meta tags from head and returns before parsing body
	err = m.Metabolize(res.Body, data)
	if err != nil {
		return nil, fmt.Errorf("reading meta tags: %w", err)
	}

	return data, nil
}

func fetchYouTubeVideoMetadata(videoID string) (*URLMetadata, error) {
	var query = make(url.Values)
	query.Set("id", videoID)
	query.Set("part", "snippet")
	query.Set("key", youtubeAPIKey) // auth

	var requestURL = "https://youtube.googleapis.com/youtube/v3/videos?" + query.Encode()

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("initiating YouTube request: %w", err)
	}

	req.Header.Add("Referer", httpClientReferer)
	req.Header.Add("Accept", "application/json")

	var client = http.Client{
		Timeout: URLMetadataTimeout,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching URL: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response from YouTube: %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading YouTube response body: %w", err)
	}

	var response YouTubeVideoListResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling YouTube JSON response: %w", err)
	}

	var data = &URLMetadata{
		SiteName: "YouTube",
	}

	if len(response.Items) == 1 && response.Items[0].Snippet != nil {
		var snippet = response.Items[0].Snippet
		data.CanonicalURL = "https://www.youtube.com/watch?v=" + videoID
		data.Title = snippet.Title
		data.Description = snippet.Description

		var imageURLString string

		// https://developers.google.com/youtube/v3/docs/videos#snippet.thumbnails
		if high, ok := snippet.Thumbnails["high"]; ok {
			imageURLString = high.URL
		} else if medium, ok := snippet.Thumbnails["medium"]; ok {
			imageURLString = medium.URL
		} else if small, ok := snippet.Thumbnails["default"]; ok {
			imageURLString = small.URL
		}

		if imageURLString != "" {
			imageURL, err := url.ParseRequestURI(imageURLString)
			if err != nil {
				return nil, fmt.Errorf("parsing YouTube thumbnail URL: %w", err)
			}
			if imageURL != nil {
				data.ImageURL = *imageURL
			}
		}
	}

	return data, nil
}

func loadImageThumbData(imageURL string) (string, error) {
	client := http.Client{
		Timeout: URLMetadataTimeout,
	}
	res, err := client.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("fetching URL: %w", err)
	}

	defer res.Body.Close()

	// Load image
	image, _, err := image.Decode(res.Body)
	if err != nil {
		return "", fmt.Errorf("decoding image: %w", err)
	}

	thumb := resize.Thumbnail(maxThumbnailDims, maxThumbnailDims, image, resize.Lanczos3)

	stringBuilder := new(strings.Builder)
	base64Writer := base64.NewEncoder(base64.StdEncoding, stringBuilder)

	err = png.Encode(base64Writer, thumb)
	if err != nil {
		return "", fmt.Errorf("encoding thumbnail: %w", err)
	}

	return "data:image/png;base64," + stringBuilder.String(), nil
}

func createURLObject(tx *sql.Tx, specID int64, url string) (*URLObject, error) {
	data, err := fetchMetadata(url)
	if err != nil {
		return nil, fmt.Errorf("loading url metadata: %w", err)
	}

	urlObject := &URLObject{
		SpecID: specID,
		URL:    url,
	}

	if strings.TrimSpace(data.Title) != "" {
		title := Substr(strings.TrimSpace(data.Title), urlTitleMaxLen)
		urlObject.Title = &title
	}

	if strings.TrimSpace(data.Description) != "" {
		desc := Substr(strings.TrimSpace(data.Description), urlDescMaxLen)
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
		`INSERT INTO spec_url (spec_id, created_at, url, url_title, url_desc, url_image_data, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $2) RETURNING id, created_at, url_title, url_desc, updated_at`,
		specID, time.Now(), url, urlObject.Title, urlObject.Desc, urlObject.ImageData,
	).Scan(&urlObject.ID, &urlObject.Created, &urlObject.Title, &urlObject.Desc, &urlObject.Updated)
	if err != nil {
		return nil, fmt.Errorf("creating spec_url: %w", err)
	}

	return urlObject, nil
}

func updateURLObject(tx *sql.Tx, id int64, url string) (*URLObject, error) {

	data, err := fetchMetadata(url)
	if err != nil {
		return nil, fmt.Errorf("loading url metadata: %w", err)
	}

	urlObject := &URLObject{
		ID:  id,
		URL: url,
	}

	if strings.TrimSpace(data.Title) != "" {
		title := Substr(strings.TrimSpace(data.Title), urlTitleMaxLen)
		urlObject.Title = &title
	}

	if strings.TrimSpace(data.Description) != "" {
		desc := Substr(strings.TrimSpace(data.Description), urlDescMaxLen)
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
		WHERE id=$1 RETURNING created_at, url_title, url_desc, updated_at`,
		id, url, urlObject.Title, urlObject.Desc, urlObject.ImageData, time.Now(),
	).Scan(&urlObject.Created, &urlObject.Title, &urlObject.Desc, &urlObject.Updated)
	if err != nil {
		return nil, fmt.Errorf("updating spec_url: %w", err)
	}

	return urlObject, nil
}

// currently used in loading block ref headers piecewise
func loadURLObject(db DBConn, id int64) (*URLObject, error) {

	urlObject := &URLObject{
		ID: id,
	}

	err := db.QueryRow(
		`SELECT spec_id, created_at, url, url_title, url_desc, url_image_data, updated_at
		FROM spec_url WHERE id=$1`, id).Scan(&urlObject.SpecID, &urlObject.Created,
		&urlObject.URL, &urlObject.Title, &urlObject.Desc, &urlObject.ImageData, &urlObject.Updated)
	if err != nil {
		return nil, fmt.Errorf("reading spec_url: %w", err)
	}

	return urlObject, nil
}
