package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Sentinel2 fetches imagery from the Copernicus Data Space Ecosystem (CDSE).
// Free, open access, 10m resolution, 5-day revisit.
// Docs: https://documentation.dataspace.copernicus.eu/
type Sentinel2 struct {
	clientID     string
	clientSecret string
	token        string
	tokenExpiry  time.Time
	httpClient   *http.Client
}

type S2SearchResult struct {
	Features []S2Feature `json:"features"`
}

type S2Feature struct {
	ID         string     `json:"id"`
	Properties S2Props    `json:"properties"`
	Assets     S2Assets   `json:"assets,omitempty"`
}

type S2Props struct {
	DateTime     string  `json:"datetime"`
	CloudCover   float64 `json:"eo:cloud_cover"`
	Platform     string  `json:"platform"`
	GSD          float64 `json:"gsd"`
	Constellation string `json:"constellation"`
}

type S2Assets struct {
	Visual S2Asset `json:"visual,omitempty"`
	B02    S2Asset `json:"B02,omitempty"` // Blue
	B03    S2Asset `json:"B03,omitempty"` // Green
	B04    S2Asset `json:"B04,omitempty"` // Red
	B08    S2Asset `json:"B08,omitempty"` // NIR
}

type S2Asset struct {
	Href string `json:"href"`
	Type string `json:"type"`
}

// BBox represents a geographic bounding box [west, south, east, north].
type BBox [4]float64

// SearchParams for STAC catalog search.
type SearchParams struct {
	BBox       BBox
	DateFrom   time.Time
	DateTo     time.Time
	MaxCloud   float64 // 0-100
	MaxResults int
	Collection string // "sentinel-2-l2a"
}

func NewSentinel2(clientID, clientSecret string) *Sentinel2 {
	return &Sentinel2{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}
}

// Search finds Sentinel-2 scenes matching the parameters using the STAC API.
func (s *Sentinel2) Search(ctx context.Context, params SearchParams) ([]S2Feature, error) {
	if params.Collection == "" {
		params.Collection = "sentinel-2-l2a"
	}
	if params.MaxResults == 0 {
		params.MaxResults = 20
	}

	// Use Planetary Computer STAC API (no auth needed for search)
	baseURL := "https://planetarycomputer.microsoft.com/api/stac/v1/search"

	body := map[string]interface{}{
		"collections": []string{params.Collection},
		"bbox":        params.BBox,
		"datetime":    fmt.Sprintf("%s/%s", params.DateFrom.Format(time.RFC3339), params.DateTo.Format(time.RFC3339)),
		"limit":       params.MaxResults,
		"query": map[string]interface{}{
			"eo:cloud_cover": map[string]interface{}{
				"lte": params.MaxCloud,
			},
		},
	}

	bodyJSON, _ := json.Marshal(body)
	_ = bodyJSON

	// TODO: Make HTTP request and parse response
	return nil, fmt.Errorf("not yet implemented")
}

// Download fetches the actual imagery for a scene.
func (s *Sentinel2) Download(ctx context.Context, feature S2Feature, outputDir string) (string, error) {
	// TODO: Download COG tiles via signed URLs
	return "", fmt.Errorf("not yet implemented")
}

// FetchForLocation is a convenience method that searches and downloads the best recent image.
func (s *Sentinel2) FetchForLocation(ctx context.Context, lat, lon, radiusKm float64, maxCloud float64) (string, error) {
	// Convert lat/lon + radius to bbox
	degPerKm := 1.0 / 111.0 // approximate
	dlat := radiusKm * degPerKm
	dlon := radiusKm * degPerKm / 1.0 // TODO: adjust for latitude

	bbox := BBox{lon - dlon, lat - dlat, lon + dlon, lat + dlat}

	params := SearchParams{
		BBox:     bbox,
		DateFrom: time.Now().AddDate(0, -1, 0), // Last month
		DateTo:   time.Now(),
		MaxCloud: maxCloud,
	}

	features, err := s.Search(ctx, params)
	if err != nil {
		return "", err
	}
	if len(features) == 0 {
		return "", fmt.Errorf("no imagery found for location (%.4f, %.4f)", lat, lon)
	}

	return s.Download(ctx, features[0], "")
}

// Landsat fetches from USGS via STAC (also on Planetary Computer).
type Landsat struct {
	httpClient *http.Client
}

func NewLandsat() *Landsat {
	return &Landsat{httpClient: &http.Client{Timeout: 60 * time.Second}}
}

func (l *Landsat) Search(ctx context.Context, params SearchParams) ([]S2Feature, error) {
	params.Collection = "landsat-c2-l2"
	// TODO: implement using Planetary Computer STAC
	return nil, fmt.Errorf("not yet implemented")
}
