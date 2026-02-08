package geo

import "math"

const (
	EarthRadiusKm = 6371.0
	DegToRad      = math.Pi / 180.0
	RadToDeg      = 180.0 / math.Pi
)

// Point represents a geographic coordinate.
type Point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// BBox represents a bounding box.
type BBox struct {
	West  float64 `json:"west"`
	South float64 `json:"south"`
	East  float64 `json:"east"`
	North float64 `json:"north"`
}

// BBoxFromCenter creates a bounding box from a center point and radius.
func BBoxFromCenter(center Point, radiusKm float64) BBox {
	dLat := (radiusKm / EarthRadiusKm) * RadToDeg
	dLon := (radiusKm / (EarthRadiusKm * math.Cos(center.Lat*DegToRad))) * RadToDeg
	return BBox{
		West:  center.Lon - dLon,
		South: center.Lat - dLat,
		East:  center.Lon + dLon,
		North: center.Lat + dLat,
	}
}

// Haversine returns the distance in km between two points.
func Haversine(a, b Point) float64 {
	dLat := (b.Lat - a.Lat) * DegToRad
	dLon := (b.Lon - a.Lon) * DegToRad
	aLat := a.Lat * DegToRad
	bLat := b.Lat * DegToRad

	sinDLat := math.Sin(dLat / 2)
	sinDLon := math.Sin(dLon / 2)
	h := sinDLat*sinDLat + math.Cos(aLat)*math.Cos(bLat)*sinDLon*sinDLon
	return 2 * EarthRadiusKm * math.Asin(math.Sqrt(h))
}

// PixelToGeo converts pixel coordinates to geographic coordinates given reference points and GSD.
func PixelToGeo(px, py int, topLeft Point, gsdMeters float64) Point {
	dLatPerPx := (gsdMeters / 1000.0 / EarthRadiusKm) * RadToDeg
	dLonPerPx := (gsdMeters / 1000.0 / (EarthRadiusKm * math.Cos(topLeft.Lat*DegToRad))) * RadToDeg
	return Point{
		Lat: topLeft.Lat - float64(py)*dLatPerPx,
		Lon: topLeft.Lon + float64(px)*dLonPerPx,
	}
}
