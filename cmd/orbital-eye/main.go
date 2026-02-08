package main

import (
	"fmt"
	"os"

	"github.com/clearclown/orbital-eye/internal/config"
)

var version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := config.Load()
	_ = cfg

	switch os.Args[1] {
	case "fetch":
		cmdFetch(os.Args[2:])
	case "detect":
		cmdDetect(os.Args[2:])
	case "monitor":
		cmdMonitor(os.Args[2:])
	case "search":
		cmdSearch(os.Args[2:])
	case "report":
		cmdReport(os.Args[2:])
	case "serve":
		cmdServe(os.Args[2:])
	case "version":
		fmt.Printf("orbital-eye %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Orbital Eye — Satellite Imagery Intelligence Platform

Usage:
  orbital-eye <command> [flags]

Commands:
  fetch       Fetch satellite imagery for an area
  detect      Detect objects in satellite imagery
  monitor     Monitor a location for changes over time
  search      Search an area for facilities of interest
  report      Generate intelligence report
  serve       Start API server + web dashboard
  version     Show version

Use "orbital-eye <command> --help" for more information.`)
}

func cmdFetch(args []string)   { fmt.Println("fetch: not yet implemented") }
func cmdDetect(args []string)  { fmt.Println("detect: not yet implemented") }
func cmdMonitor(args []string) { fmt.Println("monitor: not yet implemented") }
func cmdSearch(args []string)  { fmt.Println("search: not yet implemented") }
func cmdReport(args []string)  { fmt.Println("report: not yet implemented") }
func cmdServe(args []string)   { fmt.Println("serve: not yet implemented") }
