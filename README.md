# Orbital Eye 🛰️

**Open-Source Satellite Imagery Intelligence Platform** — AI-powered OSINT tool for automated analysis of satellite imagery, military facility detection, and geospatial change monitoring.

## Overview

Orbital Eye automates the labor-intensive process of satellite imagery analysis for open-source intelligence. It detects military installations, counts vessels and aircraft, tracks facility changes over time, and generates actionable intelligence reports — tasks that currently require trained analysts spending hours per image.

Inspired by organizations like [DEEP DIVE](https://deepdive146.com/) that perform civilian satellite intelligence analysis, this project aims to democratize access to these capabilities.

## Features

### 🎯 Object Detection
- **Vessel detection & classification**: Surface combatants, submarines, merchant ships, fishing vessels
- **Aircraft detection**: Fighter jets, bombers, transport aircraft, helicopters, UAVs on airfields
- **Vehicle detection**: Tanks, APCs, artillery, TELs (Transporter Erector Launchers), trucks
- **Building classification**: Hardened aircraft shelters, radar installations, SAM sites, bunkers

### 🏗️ Facility Detection
- **Military base identification**: Automated detection of airfields, naval bases, army installations
- **Nuclear facilities**: Reactor buildings, cooling towers, reprocessing plants
- **Missile sites**: Launch pads, TEL garages, support facilities
- **Port infrastructure**: Dry docks, piers, cranes, submarine pens

### 📡 Change Detection
- **Temporal comparison**: Before/after analysis of the same location
- **Activity monitoring**: Track deployment patterns, construction progress, operational tempo
- **Alert generation**: Automated notifications when significant changes are detected
- **Historical trending**: Long-term activity pattern analysis

### 🗺️ Geospatial Analysis
- **Measurement**: Automated length/area measurement of objects and facilities
- **Clustering**: Group related facilities and installations
- **Pattern of life**: Daily/seasonal activity pattern analysis
- **Attribution**: Match detected objects to known equipment databases

### 📊 Reporting
- **HTML/PDF reports**: Annotated imagery with findings
- **GeoJSON export**: For integration with GIS tools (QGIS, Google Earth)
- **API**: REST API for integration with other intelligence platforms
- **Dashboard**: Web-based monitoring dashboard

## Quick Start

```bash
# Install
pip install orbital-eye

# Analyze a satellite image
orbital-eye detect image.tif --objects vessels,aircraft

# Monitor a location for changes
orbital-eye monitor --lat 38.9 --lon 125.7 --interval 7d

# Generate intelligence report
orbital-eye report --location "Yulin Naval Base" --period 30d --output report.html
```

## Data Sources

### Free / Open Access
| Source | Resolution | Revisit | Coverage |
|--------|-----------|---------|----------|
| Sentinel-2 (ESA) | 10m | 5 days | Global |
| Landsat 8/9 (USGS) | 15-30m | 16 days | Global |
| MODIS (NASA) | 250m-1km | Daily | Global |

### Commercial (API integration)
| Source | Resolution | Notes |
|--------|-----------|-------|
| Planet (PlanetScope) | 3m | Daily global coverage |
| Maxar (WorldView) | 30cm | Highest resolution |
| Airbus (Pléiades Neo) | 30cm | Tasking available |
| BlackSky | 1m | Rapid revisit |
| Capella Space | 0.5m SAR | All-weather, day/night |

## Architecture

```
orbital-eye/
├── src/
│   ├── collectors/          # Satellite imagery acquisition
│   │   ├── sentinel.py      # Copernicus/Sentinel-2 API
│   │   ├── landsat.py       # USGS EarthExplorer / Landsat
│   │   ├── planet.py        # Planet Labs API
│   │   └── tiles.py         # Web map tile sources
│   ├── detectors/           # AI detection models
│   │   ├── vessels.py       # Ship detection & classification
│   │   ├── aircraft.py      # Aircraft detection on airfields
│   │   ├── vehicles.py      # Military vehicle detection
│   │   ├── facilities.py    # Base/facility classification
│   │   └── infrastructure.py # Buildings, roads, runways
│   ├── analyzers/           # Intelligence analysis
│   │   ├── change.py        # Temporal change detection
│   │   ├── measure.py       # Object measurement
│   │   ├── activity.py      # Pattern of life analysis
│   │   └── attribution.py   # Object → equipment database matching
│   ├── visualizers/         # Output generation
│   │   ├── annotate.py      # Image annotation
│   │   ├── map.py           # Interactive map (Leaflet/Mapbox)
│   │   ├── report.py        # HTML/PDF report generation
│   │   └── dashboard.py     # Web dashboard (FastAPI)
│   ├── models/              # Model management
│   │   ├── registry.py      # Model download & versioning
│   │   └── training.py      # Fine-tuning utilities
│   ├── cli.py               # CLI entry point
│   └── api.py               # REST API server
├── models/                  # Pre-trained model weights
├── data/
│   ├── equipment_db/        # Known military equipment database
│   ├── known_facilities/    # Known facility coordinates
│   └── samples/             # Sample imagery for testing
├── web/                     # Dashboard frontend
├── docs/
│   ├── methodology.md       # Detection methodology
│   ├── data-sources.md      # Imagery source guide
│   └── model-training.md    # Training custom models
└── tests/
```

## Models

### Pre-trained Models
| Model | Task | Base | mAP | Notes |
|-------|------|------|-----|-------|
| `vessel-det-v1` | Ship detection | YOLOv8 | ~85% | Surface vessels in optical imagery |
| `aircraft-det-v1` | Aircraft detection | YOLOv8 | ~80% | Parked aircraft on airfields |
| `vehicle-det-v1` | Vehicle detection | YOLOv8 | ~75% | Military vehicles |
| `facility-cls-v1` | Facility classification | ResNet-50 | ~82% | Military vs civilian |
| `change-det-v1` | Change detection | Siamese UNet | ~78% | Structural changes |

### Training Data Sources
- [DOTA](https://captain-whu.github.io/DOTA/) — Large-scale object detection in aerial images
- [xView](http://xviewdataset.org/) — Overhead imagery fine-grained detection
- [DIOR](https://gcheng-nwpu.github.io/#Datasets) — Object detection in optical remote sensing
- [SpaceNet](https://spacenet.ai/) — Building/road extraction from satellite imagery
- [FAIR1M](https://gaofen-challenge.com/) — Fine-grained recognition in high-res imagery
- [ShipRSImageNet](https://github.com/zzndream/ShipRSImageNet) — Ship detection dataset
- [RarePlanes](https://www.cosmiqworks.org/RarePlanes/) — Aircraft detection dataset
- [HRPlanesv2](https://github.com/dilsadunsal/HRPlanesv2) — Aircraft detection in Google Earth

## Methodology

### Detection Pipeline
```
Satellite Image → Preprocessing → Tiling → Object Detection → NMS → Classification → Measurement → Report
                     ↓                                                      ↓
              Geo-referencing                                        Equipment DB Match
```

### Change Detection Pipeline
```
Image_t0 + Image_t1 → Co-registration → Difference Map → Change Mask → Significance Filter → Alert
```

## Ethics & Legal

- **Public data only**: Uses commercially available or open-access satellite imagery
- **Defensive purpose**: Designed for transparency, accountability, and conflict monitoring
- **No targeting**: Not designed for kinetic targeting or offensive operations
- **Responsible disclosure**: Significant findings should be reported to appropriate authorities

## Roadmap

- [ ] **Phase 1**: Core detection models (vessels, aircraft) with Sentinel-2
- [ ] **Phase 2**: Change detection and temporal analysis
- [ ] **Phase 3**: Commercial imagery integration (Planet, Maxar)
- [ ] **Phase 4**: Web dashboard and monitoring
- [ ] **Phase 5**: SAR imagery support (all-weather capability)
- [ ] **Phase 6**: Fine-grained classification (ship class, aircraft type)

## References

- Satellite Imagery Analysis for OSINT — [Bellingcat Guide](https://www.bellingcat.com/resources/2024/01/09/using-satellite-imagery-for-osint/)
- DEEP DIVE — [Civilian Intelligence Analysis](https://deepdive146.com/)
- Center for Strategic and International Studies — [Beyond Parallel](https://beyondparallel.csis.org/)
- Middlebury Institute — [Open Nuclear Network](https://opennuclear.org/)

## License

MIT
