package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/clearclown/orbital-eye/internal/collector"
	"github.com/clearclown/orbital-eye/internal/config"
	"github.com/clearclown/orbital-eye/internal/detector"
	"github.com/clearclown/orbital-eye/internal/geo"
	"github.com/clearclown/orbital-eye/internal/report"
)

var version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "fetch":
		cmdFetch(os.Args[2:])
	case "detect":
		cmdDetect(os.Args[2:])
	case "report":
		cmdReport(os.Args[2:])
	case "monitor":
		cmdMonitor(os.Args[2:])
	case "search":
		cmdSearch(os.Args[2:])
	case "health":
		cmdHealth(os.Args[2:])
	case "version":
		fmt.Printf("orbital-eye %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Orbital Eye 🛰️  — Satellite Imagery Intelligence Platform

Usage:
  orbital-eye <command> [flags]

Commands:
  fetch       Fetch satellite imagery for a location
  detect      Detect objects in satellite imagery
  report      Generate intelligence report from detection results
  monitor     Monitor a location for changes
  search      Search for imagery and detect objects in one step
  health      Check AI worker status
  version     Show version`)
}

func cmdFetch(args []string) {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	lat := fs.Float64("lat", 0, "Latitude")
	lon := fs.Float64("lon", 0, "Longitude")
	radius := fs.Float64("radius", 10, "Radius in km")
	maxCloud := fs.Float64("cloud", 20, "Max cloud cover %")
	dateFrom := fs.String("from", "", "Start date (YYYY-MM-DD)")
	dateTo := fs.String("to", "", "End date (YYYY-MM-DD)")
	outDir := fs.String("out", "data/cache", "Output directory")
	fs.Parse(args)

	if *lat == 0 && *lon == 0 {
		fmt.Fprintln(os.Stderr, "Error: --lat and --lon are required")
		fs.Usage()
		os.Exit(1)
	}

	ctx := context.Background()
	s2 := collector.NewSentinel2(*outDir)

	bbox := geo.BBoxFromCenter(geo.Point{Lat: *lat, Lon: *lon}, *radius)

	from := time.Now().AddDate(0, -3, 0)
	to := time.Now()
	if *dateFrom != "" {
		from, _ = time.Parse("2006-01-02", *dateFrom)
	}
	if *dateTo != "" {
		to, _ = time.Parse("2006-01-02", *dateTo)
	}

	fmt.Printf("🛰️  Searching Sentinel-2 imagery...\n")
	fmt.Printf("   Location: (%.4f, %.4f), Radius: %.1fkm, Cloud: <%.0f%%\n", *lat, *lon, *radius, *maxCloud)
	fmt.Printf("   Period: %s to %s\n", from.Format("2006-01-02"), to.Format("2006-01-02"))

	results, err := s2.Search(ctx, collector.SearchParams{
		BBox:       bbox,
		DateFrom:   from,
		DateTo:     to,
		MaxCloud:   *maxCloud,
		MaxResults: 10,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n📡 Found %d scenes:\n", len(results))
	for i, r := range results {
		fmt.Printf("  [%d] %s  Date: %s  Cloud: %.1f%%  GSD: %.1fm\n",
			i, r.ID, r.Date.Format("2006-01-02"), r.CloudCover, r.GSD)
	}

	if len(results) > 0 {
		fmt.Printf("\n⬇️  Downloading best scene: %s\n", results[0].ID)
		path, err := s2.Download(ctx, results[0], []string{"visual"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Download error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Saved to: %s\n", path)
	}
}

func cmdDetect(args []string) {
	fs := flag.NewFlagSet("detect", flag.ExitOnError)
	imagePath := fs.String("image", "", "Path to image file")
	objects := fs.String("objects", "all", "Object types: vessels,aircraft,vehicles,all")
	confidence := fs.Float64("confidence", 0.3, "Detection confidence threshold")
	gsd := fs.Float64("gsd", 10.0, "Ground sample distance in meters")
	aiAddr := fs.String("ai", "localhost:50051", "AI worker address")
	outputJSON := fs.Bool("json", false, "Output as JSON")
	fs.Parse(args)

	if *imagePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --image is required")
		fs.Usage()
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := detector.NewClient(*aiAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to AI worker: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	targets := strings.Split(*objects, ",")
	if *objects == "all" {
		targets = nil
	}

	fmt.Printf("🔍 Detecting objects in %s...\n", *imagePath)
	resp, err := client.DetectFromPath(ctx, *imagePath, targets, float32(*confidence), float32(*gsd), 0, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Detection error: %v\n", err)
		os.Exit(1)
	}

	if *outputJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(resp)
	} else {
		fmt.Printf("✅ Found %d objects (%.0fms)\n\n", len(resp.Detections), resp.InferenceTimeMs)
		for i, det := range resp.Detections {
			fmt.Printf("  [%d] %s (%.1f%%)", i+1, det.ClassName, det.Confidence*100)
			if det.EstimatedLengthM > 0 {
				fmt.Printf("  ~%.0fm × %.0fm", det.EstimatedLengthM, det.EstimatedWidthM)
			}
			if det.GeoCenter != nil && det.GeoCenter.Latitude != 0 {
				fmt.Printf("  @ (%.4f, %.4f)", det.GeoCenter.Latitude, det.GeoCenter.Longitude)
			}
			fmt.Println()
		}
	}
}

func cmdSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	lat := fs.Float64("lat", 0, "Latitude")
	lon := fs.Float64("lon", 0, "Longitude")
	radius := fs.Float64("radius", 10, "Radius in km")
	maxCloud := fs.Float64("cloud", 20, "Max cloud cover %")
	objects := fs.String("objects", "all", "Object types to detect")
	confidence := fs.Float64("confidence", 0.3, "Detection confidence")
	aiAddr := fs.String("ai", "localhost:50051", "AI worker address")
	fs.Parse(args)

	if *lat == 0 && *lon == 0 {
		fmt.Fprintln(os.Stderr, "Error: --lat and --lon are required")
		os.Exit(1)
	}

	ctx := context.Background()

	// Step 1: Fetch imagery
	fmt.Println("🛰️  Step 1: Fetching satellite imagery...")
	s2 := collector.NewSentinel2("data/cache")
	result, path, err := s2.FetchBest(ctx, *lat, *lon, *radius, *maxCloud)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fetch error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   Scene: %s (Cloud: %.1f%%, Date: %s)\n", result.ID, result.CloudCover, result.Date.Format("2006-01-02"))

	// Step 2: Run detection
	fmt.Println("🔍 Step 2: Running object detection...")
	client, err := detector.NewClient(*aiAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AI worker unavailable: %v\n", err)
		fmt.Println("   (Start AI worker: cd ai && python server.py)")
		os.Exit(1)
	}
	defer client.Close()

	targets := strings.Split(*objects, ",")
	if *objects == "all" {
		targets = nil
	}

	resp, err := client.DetectFromPath(ctx, path+"/visual.tif", targets, float32(*confidence), float32(result.GSD), *lat, *lon)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Detection error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Results: %d objects detected\n", len(resp.Detections))
	for i, det := range resp.Detections {
		fmt.Printf("  [%d] %s (%.1f%%)\n", i+1, det.ClassName, det.Confidence*100)
	}
}

func cmdReport(args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	inputFile := fs.String("input", "", "Path to detection results JSON (from detect --json)")
	location := fs.String("location", "", "Location name for report header")
	lat := fs.Float64("lat", 0, "Latitude (for metadata)")
	lon := fs.Float64("lon", 0, "Longitude (for metadata)")
	period := fs.String("period", "", "Analysis period (e.g. 30d)")
	geojson := fs.String("geojson", "", "Output path for GeoJSON file")
	outFile := fs.String("out", "", "Output path for text report (default: stdout)")
	fs.Parse(args)

	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --input is required (path to detect --json output)")
		fs.Usage()
		os.Exit(1)
	}

	result, err := report.LoadDetectResult(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	summary := report.Summarize(result)
	meta := report.ReportMeta{
		Location: *location,
		Lat:      *lat,
		Lon:      *lon,
		Period:   *period,
		Source:   result.ModelVersion,
	}

	// Write text report
	w := os.Stdout
	if *outFile != "" {
		f, err := os.Create(*outFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating report file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}
	report.PrintText(summary, meta, w)

	if *outFile != "" {
		fmt.Fprintf(os.Stderr, "Report saved to: %s\n", *outFile)
	}

	// Write GeoJSON if requested
	if *geojson != "" {
		if err := report.WriteGeoJSON(summary, *geojson); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing GeoJSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "GeoJSON saved to: %s\n", *geojson)
	}
}

func cmdMonitor(args []string) {
	fmt.Println("monitor: coming in Phase 2")
}

func cmdHealth(args []string) {
	fs := flag.NewFlagSet("health", flag.ExitOnError)
	aiAddr := fs.String("ai", "localhost:50051", "AI worker address")
	fs.Parse(args)

	cfg := config.Load()
	addr := cfg.AIWorker.Address
	if *aiAddr != "localhost:50051" {
		addr = *aiAddr
	}

	client, err := detector.NewClient(addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Cannot connect to AI worker at %s: %v\n", addr, err)
		os.Exit(1)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Health(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Health check failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ AI Worker: ready=%v\n", resp.Ready)
	fmt.Printf("   Models loaded: %v\n", resp.LoadedModels)
	if resp.GpuMemoryTotalMb > 0 {
		fmt.Printf("   GPU Memory: %dMB / %dMB\n", resp.GpuMemoryUsedMb, resp.GpuMemoryTotalMb)
	}
}
