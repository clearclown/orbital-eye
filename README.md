# Orbital Eye 🛰️

**Open-Source Satellite Imagery Intelligence Platform** — AI-powered OSINT tool for automated analysis of satellite imagery, military facility detection, and geospatial change monitoring.

## Overview

Orbital Eye automates the labor-intensive process of satellite imagery analysis for open-source intelligence. It detects military installations, counts vessels and aircraft, tracks facility changes over time, and generates actionable intelligence reports.

**Architecture**: Go (core/CLI/API) + Python AI worker (gRPC) for ML inference.

## Features

- 🎯 **Object Detection** — Vessels, aircraft, vehicles, facilities (YOLOv8)
- 🏗️ **Facility Detection** — Airfields, naval bases, missile sites, nuclear facilities
- 📡 **Change Detection** — Temporal before/after analysis with automated alerts
- 🗺️ **Geospatial Analysis** — Measurement, clustering, pattern-of-life
- 📊 **Reporting** — HTML/PDF reports, GeoJSON export, web dashboard

## Quick Start

```bash
# Build
make build

# Start AI worker (separate terminal)
make ai-deps
make ai-serve

# Fetch imagery
orbital-eye fetch --lat 18.2269 --lon 109.5331 --radius 10 --source sentinel2

# Detect objects
orbital-eye detect image.tif --objects vessels,aircraft

# Monitor a location
orbital-eye monitor --lat 38.9 --lon 125.7 --interval 7d

# Generate report
orbital-eye report --location "Yulin Naval Base" --period 30d
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Go (Core)                                              │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────┐  │
│  │ CLI      │ │ API      │ │Collector │ │ Visualizer│  │
│  │ (typer)  │ │ (HTTP)   │ │Sentinel-2│ │ Report/Map│  │
│  └────┬─────┘ └────┬─────┘ │Landsat   │ └───────────┘  │
│       │             │       │Planet    │                 │
│       └──────┬──────┘       └──────────┘                │
│              │                                           │
│              │ gRPC                                      │
│──────────────┼───────────────────────────────────────────│
│              ▼                                           │
│  Python (AI Worker)                                     │
│  ┌──────────────────────────────────────────────┐       │
│  │ Object Detection (YOLOv8)                     │       │
│  │ Change Detection (Siamese UNet)               │       │
│  │ Classification (ResNet/EfficientNet)          │       │
│  │ Super-Resolution                              │       │
│  └──────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────┘
```

## Data Sources

### Free / Open Access
| Source | Resolution | Revisit | API |
|--------|-----------|---------|-----|
| Sentinel-2 (ESA) | 10m | 5 days | Planetary Computer STAC |
| Landsat 8/9 (USGS) | 15-30m | 16 days | Planetary Computer STAC |
| MODIS (NASA) | 250m | Daily | NASA Earthdata |

### Commercial (optional)
| Source | Resolution | Notes |
|--------|-----------|-------|
| Planet | 3m | Daily global |
| Maxar | 30cm | Highest resolution |
| Capella Space | 0.5m SAR | All-weather |

## Training Data (Open)

| Dataset | Objects | Images | License |
|---------|---------|--------|---------|
| [DOTA v2.0](https://captain-whu.github.io/DOTA/) | Ship, plane, vehicle, etc. | 11,268 | Research |
| [xView](http://xviewdataset.org/) | 60 classes | 1,400 | CC BY-NC-SA 4.0 |
| [ShipRSImageNet](https://github.com/zzndream/ShipRSImageNet) | 50 ship types | 3,435 | Research |
| [HRPlanesv2](https://github.com/dilsadunsal/HRPlanesv2) | Aircraft | 2,120 | Open |
| [RarePlanes](https://www.cosmiqworks.org/RarePlanes/) | Aircraft | 50,000 annot. | CC BY-SA 4.0 |
| [SpaceNet](https://spacenet.ai/) | Buildings, roads | Multiple | CC BY-SA 4.0 |
| [FAIR1M](https://gaofen-challenge.com/) | Fine-grained | 15,000+ | Research |

## Development

```bash
# Prerequisites
go 1.23+, Python 3.11+, protoc

# Build everything
make proto   # Generate gRPC stubs
make build   # Build Go binary
make ai-deps # Install Python deps

# Run
make ai-serve &          # Start AI worker
./bin/orbital-eye serve   # Start API server
```

## Roadmap

- [ ] Phase 1: Sentinel-2 collector + vessel detection (YOLOv8)
- [ ] Phase 2: Change detection pipeline
- [ ] Phase 3: Web dashboard + monitoring
- [ ] Phase 4: Fine-grained classification (ship/aircraft types)
- [ ] Phase 5: SAR imagery support
- [ ] Phase 6: Equipment database matching

## Ethics & Legal

- Uses **publicly available** satellite imagery and **open data** only
- Designed for **transparency and accountability** in conflict monitoring
- **Not** for targeting, offensive operations, or surveillance of individuals
- Significant findings should be reported to appropriate authorities

## References

- [DEEP DIVE](https://deepdive146.com/) — Civilian intelligence analysis
- [Bellingcat Satellite Guide](https://www.bellingcat.com/resources/)
- [Beyond Parallel (CSIS)](https://beyondparallel.csis.org/)
- [Open Nuclear Network](https://opennuclear.org/)

## License

MIT
