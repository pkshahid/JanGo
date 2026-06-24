// Package gis implements geospatial data support inspired by GeoDjango.
// It provides geometry types, spatial lookups, and distance calculations.
package gis

import (
	"fmt"
	"math"
)

// SRID constants for common spatial reference systems.
const (
	WGS84       = 4326 // GPS coordinate system
	WebMercator = 3857 // Web maps (Google, OSM)
)

// Point represents a 2D geographic point.
type Point struct {
	X    float64 // Longitude
	Y    float64 // Latitude
	SRID int
}

// NewPoint creates a new geographic point.
func NewPoint(lon, lat float64, srid int) *Point {
	return &Point{X: lon, Y: lat, SRID: srid}
}

// String returns WKT representation.
func (p *Point) String() string {
	return fmt.Sprintf("POINT(%f %f)", p.X, p.Y)
}

// GeoJSON returns GeoJSON representation.
func (p *Point) GeoJSON() string {
	return fmt.Sprintf(`{"type":"Point","coordinates":[%f,%f]}`, p.X, p.Y)
}

// LineString represents a series of connected points.
type LineString struct {
	Points []*Point
	SRID   int
}

// NewLineString creates a new line string.
func NewLineString(srid int, points ...*Point) *LineString {
	return &LineString{Points: points, SRID: srid}
}

// Length returns the total length of the line string in degrees.
func (ls *LineString) Length() float64 {
	total := 0.0
	for i := 1; i < len(ls.Points); i++ {
		total += EuclideanDistance(ls.Points[i-1], ls.Points[i])
	}
	return total
}

// String returns WKT representation.
func (ls *LineString) String() string {
	s := "LINESTRING("
	for i, p := range ls.Points {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%f %f", p.X, p.Y)
	}
	return s + ")"
}

// Polygon represents a closed area.
type Polygon struct {
	ExteriorRing []*Point
	Holes        [][]*Point
	SRID         int
}

// NewPolygon creates a new polygon from an exterior ring.
func NewPolygon(srid int, exterior ...*Point) *Polygon {
	return &Polygon{
		ExteriorRing: exterior,
		SRID:         srid,
	}
}

// AddHole adds an interior ring (hole) to the polygon.
func (p *Polygon) AddHole(ring []*Point) {
	p.Holes = append(p.Holes, ring)
}

// Contains checks if a point is inside the polygon using ray casting.
func (p *Polygon) Contains(pt *Point) bool {
	return pointInPolygon(pt, p.ExteriorRing)
}

// Area returns the approximate area of the polygon in square degrees.
func (p *Polygon) Area() float64 {
	return polygonArea(p.ExteriorRing)
}

// String returns WKT representation.
func (p *Polygon) String() string {
	s := "POLYGON(("
	for i, pt := range p.ExteriorRing {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%f %f", pt.X, pt.Y)
	}
	return s + "))"
}

// MultiPoint represents a collection of points.
type MultiPoint struct {
	Points []*Point
	SRID   int
}

// BoundingBox represents a geographic bounding box.
type BoundingBox struct {
	MinX, MinY, MaxX, MaxY float64
	SRID                   int
}

// NewBoundingBox creates a new bounding box.
func NewBoundingBox(minX, minY, maxX, maxY float64, srid int) *BoundingBox {
	return &BoundingBox{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY, SRID: srid}
}

// Contains checks if a point is within the bounding box.
func (bb *BoundingBox) Contains(p *Point) bool {
	return p.X >= bb.MinX && p.X <= bb.MaxX && p.Y >= bb.MinY && p.Y <= bb.MaxY
}

// Intersects checks if two bounding boxes overlap.
func (bb *BoundingBox) Intersects(other *BoundingBox) bool {
	return bb.MinX <= other.MaxX && bb.MaxX >= other.MinX &&
		bb.MinY <= other.MaxY && bb.MaxY >= other.MinY
}

// EuclideanDistance calculates the Euclidean distance between two points.
func EuclideanDistance(p1, p2 *Point) float64 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// HaversineDistance calculates the great-circle distance between two points
// on Earth in meters, using the Haversine formula.
func HaversineDistance(p1, p2 *Point) float64 {
	const earthRadius = 6371000.0 // meters

	lat1 := toRadians(p1.Y)
	lat2 := toRadians(p2.Y)
	dLat := toRadians(p2.Y - p1.Y)
	dLon := toRadians(p2.X - p1.X)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// WithinDistance checks if two points are within the given distance (meters).
func WithinDistance(p1, p2 *Point, distanceMeters float64) bool {
	return HaversineDistance(p1, p2) <= distanceMeters
}

// GeoField represents a geometry field type for models.
type GeoField struct {
	FieldType string // "PointField", "LineStringField", "PolygonField", "MultiPointField"
	SRID      int
	Dimension int // 2 or 3
	Geography bool
}

// PointField creates a point field definition.
func PointField(srid int) *GeoField {
	return &GeoField{FieldType: "PointField", SRID: srid, Dimension: 2}
}

// LineStringField creates a line string field definition.
func LineStringField(srid int) *GeoField {
	return &GeoField{FieldType: "LineStringField", SRID: srid, Dimension: 2}
}

// PolygonField creates a polygon field definition.
func PolygonField(srid int) *GeoField {
	return &GeoField{FieldType: "PolygonField", SRID: srid, Dimension: 2}
}

// SpatialLookup represents spatial query operations.
type SpatialLookup string

const (
	LookupContains   SpatialLookup = "contains"
	LookupWithin     SpatialLookup = "within"
	LookupIntersects SpatialLookup = "intersects"
	LookupOverlaps   SpatialLookup = "overlaps"
	LookupDistance   SpatialLookup = "distance_lte"
	LookupBBContains SpatialLookup = "bbcontains"
	LookupBBOverlaps SpatialLookup = "bboverlaps"
)

// Helper functions

func toRadians(deg float64) float64 {
	return deg * math.Pi / 180.0
}

func pointInPolygon(pt *Point, polygon []*Point) bool {
	n := len(polygon)
	if n < 3 {
		return false
	}

	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		if (polygon[i].Y > pt.Y) != (polygon[j].Y > pt.Y) &&
			pt.X < (polygon[j].X-polygon[i].X)*(pt.Y-polygon[i].Y)/(polygon[j].Y-polygon[i].Y)+polygon[i].X {
			inside = !inside
		}
		j = i
	}
	return inside
}

func polygonArea(ring []*Point) float64 {
	n := len(ring)
	if n < 3 {
		return 0
	}
	area := 0.0
	j := n - 1
	for i := 0; i < n; i++ {
		area += (ring[j].X + ring[i].X) * (ring[j].Y - ring[i].Y)
		j = i
	}
	return math.Abs(area / 2.0)
}
