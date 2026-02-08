#!/usr/bin/env bash
set -euo pipefail

# Download open datasets for training/evaluation
# All datasets are publicly available for research use

DATADIR="${1:-data/training}"
mkdir -p "$DATADIR"

echo "=== Orbital Eye — Dataset Downloader ==="

# 1. DOTA (Large-scale Dataset for Object DeTection in Aerial images)
#    15 categories including ship, plane, helicopter, vehicle, etc.
#    https://captain-whu.github.io/DOTA/
echo "[1/6] DOTA v2.0 — download from https://captain-whu.github.io/DOTA/"
echo "  → Manual download required (registration needed)"
echo "  → Place in $DATADIR/dota/"

# 2. xView — large-scale overhead imagery detection
#    60 classes, 1M+ objects, ~0.3m GSD
#    http://xviewdataset.org/
echo "[2/6] xView — download from http://xviewdataset.org/"
echo "  → Manual download required (registration)"
echo "  → Place in $DATADIR/xview/"

# 3. ShipRSImageNet — ship detection in remote sensing
#    ~3,500 images with 17,573 ship instances, 50 ship types
#    Freely downloadable
echo "[3/6] ShipRSImageNet"
if [ ! -d "$DATADIR/shiprs" ]; then
    echo "  → Download from https://github.com/zzndream/ShipRSImageNet"
    echo "  → Place in $DATADIR/shiprs/"
fi

# 4. HRPlanesv2 — aircraft detection in Google Earth imagery
#    ~2,000 images with ~14,000 aircraft annotations
echo "[4/6] HRPlanesv2"
if [ ! -d "$DATADIR/hrplanes" ]; then
    echo "  Cloning dataset..."
    git clone --depth 1 https://github.com/dilsadunsal/HRPlanesv2.git "$DATADIR/hrplanes" 2>/dev/null || \
        echo "  → Clone from https://github.com/dilsadunsal/HRPlanesv2"
fi

# 5. RarePlanes — satellite imagery aircraft detection (synthetic + real)
#    ~50,000 aircraft annotations across 112 locations
echo "[5/6] RarePlanes — download from https://www.cosmiqworks.org/RarePlanes/"
echo "  → AWS S3 bucket: s3://rareplanes-public"
echo "  → Place in $DATADIR/rareplanes/"

# 6. SpaceNet — building/road/facility footprints
#    Multiple challenge datasets with high-res imagery
echo "[6/6] SpaceNet datasets"
echo "  → AWS: aws s3 ls s3://spacenet-dataset/"
echo "  → Place in $DATADIR/spacenet/"

echo ""
echo "=== OpenData Satellite Imagery Sources ==="
echo "Copernicus Data Space:  https://dataspace.copernicus.eu/"
echo "USGS EarthExplorer:     https://earthexplorer.usgs.gov/"
echo "NASA Earthdata:         https://earthdata.nasa.gov/"
echo "Planetary Computer:     https://planetarycomputer.microsoft.com/"
echo "OpenAerialMap:          https://openaerialmap.org/"
echo ""
echo "Done. See docs/data-sources.md for details."
