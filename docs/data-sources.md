# Data Sources Guide

## Free Sources

### Sentinel-2 (Copernicus)
- Resolution: 10m (RGB), 20m (NIR/SWIR), 60m (coastal/cirrus)
- Revisit: 5 days
- Access: Copernicus Data Space Ecosystem API
- Best for: Wide-area monitoring, change detection, vessel detection (large ships)

### Landsat 8/9 (USGS)
- Resolution: 15m (panchromatic), 30m (multispectral)
- Revisit: 16 days (8 days combined)
- Access: USGS EarthExplorer, Google Earth Engine
- Best for: Long-term historical analysis (Landsat archive goes back to 1972)

### Microsoft Planetary Computer
- Aggregates multiple sources (Sentinel, Landsat, NAIP, etc.)
- Free STAC API with cloud-optimized GeoTIFFs
- Best for: Easy programmatic access to multiple sources

## Commercial Sources

### Planet (PlanetScope)
- Resolution: 3m daily global
- Best for: Daily monitoring, rapid change detection
- Pricing: From $4,000/month for limited areas

### Maxar (WorldView-3/Legion)
- Resolution: 30cm
- Best for: Fine-grained detection (individual vehicles, aircraft types)
- Pricing: Per-image, ~$15-25/km²

## For Military OSINT Specifically
- Sentinel-2 at 10m can detect: large ships (>50m), runway activity, building construction
- Planet at 3m can detect: individual aircraft, medium vehicles, small vessels
- Maxar at 30cm can detect: vehicle types, weapon systems, personnel activity
