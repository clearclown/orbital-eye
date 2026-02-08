package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/clearclown/orbital-eye/internal/geo"
)

const (
	stacSearchURL = "https://planetarycomputer.microsoft.com/api/stac/v1/search"
	stacSignURL   = "https://planetarycomputer.microsoft.com/api/sas/v1/sign"
)

type Sentinel2 struct {
	httpClient *http.Client
	cacheDir   string
}

type STACSearchRequest struct {
	Collections []string               `json:"collections"`
	Bbox        [4]float64             `json:"bbox"`
	Datetime    string                 `json:"datetime"`
	Limit       int                    `json:"limit"`
	Query       map[string]interface{} `json:"query,omitempty"`
	SortBy      []STACSortBy           `json:"sortby,omitempty"`
}

type STACSortBy struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type STACResponse struct {
	Features []STACFeature `json:"features"`
}

type STACFeature struct {
	ID         string                    `json:"id"`
	Properties STACProperties            `json:"properties"`
	Assets     map[string]STACAsset      `json:"assets"`
	Bbox       [4]float64                `json:"bbox"`
}

type STACProperties struct {
	DateTime   string  `json:"datetime"`
	CloudCover float64 `json:"eo:cloud_cover"`
	GSD        float64 `json:"gsd"`
	Platform   string  `json:"platform"`
}

type STACAsset struct {
	Href string `json:"href"`
	Type string `json:"type"`
}

type SearchParams struct {
	BBox       geo.BBox
	DateFrom   time.Time
	DateTo     time.Time
	MaxCloud   float64
	MaxResults int
}

type ImageResult struct {
	ID         string
	Date       time.Time
	CloudCover float64
	GSD        float64
	Platform   string
	LocalPath  string
	Assets     map[string]string
}

func NewSentinel2(cacheDir string) *Sentinel2 {
	return &Sentinel2{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		cacheDir:   cacheDir,
	}
}

func (s *Sentinel2) Search(ctx context.Context, params SearchParams) ([]ImageResult, error) {
	if params.MaxResults == 0 {
		params.MaxResults = 10
	}

	reqBody := STACSearchRequest{
		Collections: []string{"sentinel-2-l2a"},
		Bbox:        [4]float64{params.BBox.West, params.BBox.South, params.BBox.East, params.BBox.North},
		Datetime:    fmt.Sprintf("%s/%s", params.DateFrom.Format(time.RFC3339), params.DateTo.Format(time.RFC3339)),
		Limit:       params.MaxResults,
		Query: map[string]interface{}{
			"eo:cloud_cover": map[string]interface{}{"lte": params.MaxCloud},
		},
		SortBy: []STACSortBy{
			{Field: "properties.eo:cloud_cover", Direction: "asc"},
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", stacSearchURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("STAC search failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("STAC search returned %d: %s", resp.StatusCode, string(body))
	}

	var stacResp STACResponse
	if err := json.NewDecoder(resp.Body).Decode(&stacResp); err != nil {
		return nil, fmt.Errorf("parse STAC response: %w", err)
	}

	var results []ImageResult
	for _, f := range stacResp.Features {
		dt, _ := time.Parse(time.RFC3339, f.Properties.DateTime)
		r := ImageResult{
			ID:         f.ID,
			Date:       dt,
			CloudCover: f.Properties.CloudCover,
			GSD:        f.Properties.GSD,
			Platform:   f.Properties.Platform,
			Assets:     make(map[string]string),
		}
		for k, v := range f.Assets {
			r.Assets[k] = v.Href
		}
		results = append(results, r)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CloudCover < results[j].CloudCover
	})

	return results, nil
}

func (s *Sentinel2) signURL(ctx context.Context, rawURL string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", stacSignURL+"?href="+rawURL, nil)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Href string `json:"href"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Href == "" {
		return rawURL, nil
	}
	return result.Href, nil
}

func (s *Sentinel2) Download(ctx context.Context, result ImageResult, bands []string) (string, error) {
	outDir := filepath.Join(s.cacheDir, result.ID)
	os.MkdirAll(outDir, 0755)

	if len(bands) == 0 {
		bands = []string{"visual"} // True-color composite
	}

	for _, band := range bands {
		href, ok := result.Assets[band]
		if !ok {
			continue
		}

		signed, err := s.signURL(ctx, href)
		if err != nil {
			return "", fmt.Errorf("sign URL for %s: %w", band, err)
		}

		outPath := filepath.Join(outDir, band+".tif")
		if _, err := os.Stat(outPath); err == nil {
			continue // Already downloaded
		}

		req, _ := http.NewRequestWithContext(ctx, "GET", signed, nil)
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("download %s: %w", band, err)
		}
		defer resp.Body.Close()

		f, err := os.Create(outPath)
		if err != nil {
			return "", err
		}
		_, err = io.Copy(f, resp.Body)
		f.Close()
		if err != nil {
			return "", fmt.Errorf("write %s: %w", band, err)
		}

		fmt.Printf("Downloaded %s → %s\n", band, outPath)
	}

	return outDir, nil
}

func (s *Sentinel2) FetchBest(ctx context.Context, lat, lon, radiusKm, maxCloud float64) (*ImageResult, string, error) {
	bbox := geo.BBoxFromCenter(geo.Point{Lat: lat, Lon: lon}, radiusKm)

	results, err := s.Search(ctx, SearchParams{
		BBox:       bbox,
		DateFrom:   time.Now().AddDate(0, -3, 0),
		DateTo:     time.Now(),
		MaxCloud:   maxCloud,
		MaxResults: 5,
	})
	if err != nil {
		return nil, "", err
	}
	if len(results) == 0 {
		return nil, "", fmt.Errorf("no imagery found for (%.4f, %.4f) with <%g%% cloud", lat, lon, maxCloud)
	}

	best := &results[0]
	path, err := s.Download(ctx, *best, []string{"visual"})
	if err != nil {
		return nil, "", err
	}

	return best, path, nil
}
