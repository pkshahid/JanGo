package gis

import (
	"math"
	"strings"
	"testing"
)

func TestPoint(t *testing.T) {
	p := NewPoint(-73.9857, 40.7484, WGS84)
	if p.X != -73.9857 || p.Y != 40.7484 {
		t.Errorf("unexpected coordinates: %f, %f", p.X, p.Y)
	}
	if p.SRID != WGS84 {
		t.Errorf("expected SRID %d, got %d", WGS84, p.SRID)
	}

	wkt := p.String()
	if !strings.HasPrefix(wkt, "POINT(") {
		t.Errorf("unexpected WKT: %s", wkt)
	}

	geojson := p.GeoJSON()
	if !strings.Contains(geojson, "Point") {
		t.Errorf("unexpected GeoJSON: %s", geojson)
	}
}

func TestLineString(t *testing.T) {
	ls := NewLineString(WGS84,
		NewPoint(0, 0, WGS84),
		NewPoint(1, 1, WGS84),
		NewPoint(2, 0, WGS84),
	)

	if len(ls.Points) != 3 {
		t.Errorf("expected 3 points, got %d", len(ls.Points))
	}

	length := ls.Length()
	if length <= 0 {
		t.Error("length should be positive")
	}

	wkt := ls.String()
	if !strings.HasPrefix(wkt, "LINESTRING(") {
		t.Errorf("unexpected WKT: %s", wkt)
	}
}

func TestPolygon(t *testing.T) {
	// Square polygon
	poly := NewPolygon(WGS84,
		NewPoint(0, 0, WGS84),
		NewPoint(10, 0, WGS84),
		NewPoint(10, 10, WGS84),
		NewPoint(0, 10, WGS84),
		NewPoint(0, 0, WGS84),
	)

	// Point inside
	inside := NewPoint(5, 5, WGS84)
	if !poly.Contains(inside) {
		t.Error("point (5,5) should be inside the polygon")
	}

	// Point outside
	outside := NewPoint(15, 15, WGS84)
	if poly.Contains(outside) {
		t.Error("point (15,15) should be outside the polygon")
	}

	area := poly.Area()
	if area != 100.0 {
		t.Errorf("expected area 100, got %f", area)
	}

	wkt := poly.String()
	if !strings.HasPrefix(wkt, "POLYGON((") {
		t.Errorf("unexpected WKT: %s", wkt)
	}
}

func TestBoundingBox(t *testing.T) {
	bb := NewBoundingBox(-1, -1, 1, 1, WGS84)

	inside := NewPoint(0, 0, WGS84)
	if !bb.Contains(inside) {
		t.Error("point should be inside bounding box")
	}

	outside := NewPoint(5, 5, WGS84)
	if bb.Contains(outside) {
		t.Error("point should be outside bounding box")
	}

	// Intersects
	other := NewBoundingBox(0, 0, 2, 2, WGS84)
	if !bb.Intersects(other) {
		t.Error("bounding boxes should intersect")
	}

	noOverlap := NewBoundingBox(5, 5, 10, 10, WGS84)
	if bb.Intersects(noOverlap) {
		t.Error("bounding boxes should not intersect")
	}
}

func TestHaversineDistance(t *testing.T) {
	// New York to London (approximate)
	nyc := NewPoint(-74.0060, 40.7128, WGS84)
	london := NewPoint(-0.1278, 51.5074, WGS84)

	dist := HaversineDistance(nyc, london)

	// Should be approximately 5570 km
	expectedKm := 5570.0
	actualKm := dist / 1000.0
	tolerance := 50.0 // 50km tolerance

	if math.Abs(actualKm-expectedKm) > tolerance {
		t.Errorf("expected ~%f km, got %f km", expectedKm, actualKm)
	}
}

func TestWithinDistance(t *testing.T) {
	p1 := NewPoint(0, 0, WGS84)
	p2 := NewPoint(0.001, 0.001, WGS84) // Very close

	if !WithinDistance(p1, p2, 1000) { // 1km
		t.Error("points should be within 1km")
	}

	p3 := NewPoint(10, 10, WGS84) // Far
	if WithinDistance(p1, p3, 1000) {
		t.Error("points should not be within 1km")
	}
}

func TestEuclideanDistance(t *testing.T) {
	p1 := NewPoint(0, 0, WGS84)
	p2 := NewPoint(3, 4, WGS84)

	dist := EuclideanDistance(p1, p2)
	if dist != 5.0 {
		t.Errorf("expected 5.0, got %f", dist)
	}
}

func TestGeoFields(t *testing.T) {
	pf := PointField(WGS84)
	if pf.FieldType != "PointField" || pf.SRID != WGS84 {
		t.Errorf("unexpected PointField: %+v", pf)
	}

	lf := LineStringField(WebMercator)
	if lf.FieldType != "LineStringField" || lf.SRID != WebMercator {
		t.Errorf("unexpected LineStringField: %+v", lf)
	}

	polyf := PolygonField(WGS84)
	if polyf.FieldType != "PolygonField" {
		t.Errorf("unexpected PolygonField: %+v", polyf)
	}
}

func TestPolygonHole(t *testing.T) {
	poly := NewPolygon(WGS84,
		NewPoint(0, 0, WGS84),
		NewPoint(10, 0, WGS84),
		NewPoint(10, 10, WGS84),
		NewPoint(0, 10, WGS84),
		NewPoint(0, 0, WGS84),
	)

	hole := []*Point{
		NewPoint(3, 3, WGS84),
		NewPoint(7, 3, WGS84),
		NewPoint(7, 7, WGS84),
		NewPoint(3, 7, WGS84),
		NewPoint(3, 3, WGS84),
	}
	poly.AddHole(hole)

	if len(poly.Holes) != 1 {
		t.Errorf("expected 1 hole, got %d", len(poly.Holes))
	}
}
