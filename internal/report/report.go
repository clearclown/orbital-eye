package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Detection mirrors the protobuf Detection message for JSON input.
type Detection struct {
	ClassName        string            `json:"class_name"`
	Confidence       float32           `json:"confidence"`
	Bbox             BBox              `json:"bbox"`
	GeoCenter        *GeoPoint         `json:"geo_center,omitempty"`
	EstimatedLengthM float32           `json:"estimated_length_m"`
	EstimatedWidthM  float32           `json:"estimated_width_m"`
	Attributes       map[string]string `json:"attributes,omitempty"`
}

type BBox struct {
	XMin float32 `json:"x_min"`
	YMin float32 `json:"y_min"`
	XMax float32 `json:"x_max"`
	YMax float32 `json:"y_max"`
}

type GeoPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// DetectResult is the JSON structure output by `detect --json`.
type DetectResult struct {
	Detections      []Detection `json:"detections"`
	InferenceTimeMs float32     `json:"inference_time_ms"`
	ModelVersion    string      `json:"model_version"`
}

// ReportMeta holds metadata for the report.
type ReportMeta struct {
	Location string
	Lat      float64
	Lon      float64
	Period   string
	Source   string
}

// Summary holds aggregated statistics from detections.
type Summary struct {
	TotalDetections int
	ClassCounts     map[string]int
	AvgConfidence   float32
	Detections      []Detection
}

// Summarize aggregates detection results.
func Summarize(result *DetectResult) *Summary {
	s := &Summary{
		TotalDetections: len(result.Detections),
		ClassCounts:     make(map[string]int),
		Detections:      result.Detections,
	}

	var confSum float32
	for _, d := range result.Detections {
		s.ClassCounts[d.ClassName]++
		confSum += d.Confidence
	}
	if s.TotalDetections > 0 {
		s.AvgConfidence = confSum / float32(s.TotalDetections)
	}

	return s
}

// PrintText writes a human-readable report to the given file or stdout.
func PrintText(summary *Summary, meta ReportMeta, w *os.File) {
	fmt.Fprintf(w, "═══════════════════════════════════════════════════════\n")
	fmt.Fprintf(w, "  ORBITAL EYE — Intelligence Report\n")
	fmt.Fprintf(w, "  Generated: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(w, "═══════════════════════════════════════════════════════\n\n")

	if meta.Location != "" {
		fmt.Fprintf(w, "  Location:  %s\n", meta.Location)
	}
	if meta.Lat != 0 || meta.Lon != 0 {
		fmt.Fprintf(w, "  Coords:    %.4f, %.4f\n", meta.Lat, meta.Lon)
	}
	if meta.Period != "" {
		fmt.Fprintf(w, "  Period:    %s\n", meta.Period)
	}
	if meta.Source != "" {
		fmt.Fprintf(w, "  Source:    %s\n", meta.Source)
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "── Summary ────────────────────────────────────────────\n")
	fmt.Fprintf(w, "  Total detections:    %d\n", summary.TotalDetections)
	fmt.Fprintf(w, "  Avg confidence:      %.1f%%\n", summary.AvgConfidence*100)
	fmt.Fprintf(w, "\n")

	// Sort classes by count descending
	type classCount struct {
		name  string
		count int
	}
	var sorted []classCount
	for name, count := range summary.ClassCounts {
		sorted = append(sorted, classCount{name, count})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].count > sorted[j].count })

	fmt.Fprintf(w, "  Object breakdown:\n")
	for _, cc := range sorted {
		fmt.Fprintf(w, "    %-20s %d\n", cc.name, cc.count)
	}
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "── Detections ─────────────────────────────────────────\n")
	for i, d := range summary.Detections {
		fmt.Fprintf(w, "  [%d] %s (%.1f%%)", i+1, d.ClassName, d.Confidence*100)
		if d.EstimatedLengthM > 0 {
			fmt.Fprintf(w, "  ~%.0fm x %.0fm", d.EstimatedLengthM, d.EstimatedWidthM)
		}
		if d.GeoCenter != nil && (d.GeoCenter.Latitude != 0 || d.GeoCenter.Longitude != 0) {
			fmt.Fprintf(w, "  @ (%.6f, %.6f)", d.GeoCenter.Latitude, d.GeoCenter.Longitude)
		}
		fmt.Fprintf(w, "\n")
	}
	fmt.Fprintf(w, "\n═══════════════════════════════════════════════════════\n")
}

// GeoJSON types
type geojsonCollection struct {
	Type     string           `json:"type"`
	Features []geojsonFeature `json:"features"`
}

type geojsonFeature struct {
	Type       string                 `json:"type"`
	Geometry   geojsonGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type geojsonGeometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

// WriteGeoJSON writes detections as a GeoJSON FeatureCollection.
func WriteGeoJSON(summary *Summary, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	fc := geojsonCollection{
		Type:     "FeatureCollection",
		Features: make([]geojsonFeature, 0, len(summary.Detections)),
	}

	for i, d := range summary.Detections {
		if d.GeoCenter == nil || (d.GeoCenter.Latitude == 0 && d.GeoCenter.Longitude == 0) {
			continue
		}

		props := map[string]interface{}{
			"id":         i + 1,
			"class":      d.ClassName,
			"confidence": d.Confidence,
		}
		if d.EstimatedLengthM > 0 {
			props["length_m"] = d.EstimatedLengthM
			props["width_m"] = d.EstimatedWidthM
		}
		for k, v := range d.Attributes {
			props[k] = v
		}

		fc.Features = append(fc.Features, geojsonFeature{
			Type: "Feature",
			Geometry: geojsonGeometry{
				Type:        "Point",
				Coordinates: [2]float64{d.GeoCenter.Longitude, d.GeoCenter.Latitude},
			},
			Properties: props,
		})
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(fc)
}

// LoadDetectResult reads a DetectResult from a JSON file.
func LoadDetectResult(path string) (*DetectResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read detection results: %w", err)
	}
	var result DetectResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse detection results: %w", err)
	}
	return &result, nil
}
