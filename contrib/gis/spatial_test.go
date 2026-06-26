package gis

import (
	"strings"
	"testing"

	"github.com/pkshahid/JanGo/orm"
	"github.com/pkshahid/JanGo/orm/queryset"
)

// TestPlace is a model with a PointField for testing spatial query SQL generation.
type TestPlace struct {
	orm.Model
	Name     string  `gd:"CharField"`
	Location *Point  `gd:"PointField,srid=4326"`
	Region   *Polygon `gd:"PolygonField,srid=4326"`
}

func (p *TestPlace) ModelMeta() *orm.Meta {
	return &orm.Meta{
		DbTable: "test_place",
		SpatialIndexes: []orm.SpatialIndex{
			{Fields: []string{"location"}},
		},
	}
}

func setupTestPlace(t *testing.T) *orm.ModelInfo {
	t.Helper()
	orm.ClearRegistry()
	if err := orm.Register(&TestPlace{}); err != nil {
		t.Fatalf("failed to register TestPlace: %v", err)
	}
	info, err := orm.GetModelInfo(&TestPlace{})
	if err != nil {
		t.Fatalf("failed to get model info: %v", err)
	}
	return info
}

func TestSpatialIntersectsSQL(t *testing.T) {
	info := setupTestPlace(t)
	query := queryset.NewQuery(info)

	pt := NewPoint(-73.9857, 40.7484, WGS84)
	query.Where = queryset.Q(queryset.Lookup{"Location__intersects": pt})
	sql, params := query.ToSQL()

	if !strings.Contains(sql, "&&") {
		t.Errorf("expected && operator for intersects, got: %s", sql)
	}
	if !strings.Contains(sql, "ST_GeomFromText") {
		t.Errorf("expected ST_GeomFromText, got: %s", sql)
	}
	if !strings.Contains(sql, "4326") {
		t.Errorf("expected SRID 4326, got: %s", sql)
	}
	if len(params) != 1 {
		t.Errorf("expected 1 param, got %d", len(params))
	}
	wkt, ok := params[0].(string)
	if !ok || !strings.HasPrefix(wkt, "POINT(") {
		t.Errorf("expected WKT string param, got: %v", params[0])
	}
}

func TestSpatialWithinSQL(t *testing.T) {
	info := setupTestPlace(t)
	query := queryset.NewQuery(info)

	poly := NewPolygon(WGS84,
		NewPoint(0, 0, WGS84),
		NewPoint(10, 0, WGS84),
		NewPoint(10, 10, WGS84),
		NewPoint(0, 10, WGS84),
		NewPoint(0, 0, WGS84),
	)
	query.Where = queryset.Q(queryset.Lookup{"Location__within": poly})
	sql, params := query.ToSQL()

	if !strings.Contains(sql, "ST_Within") {
		t.Errorf("expected ST_Within, got: %s", sql)
	}
	if !strings.Contains(sql, "ST_GeomFromText") {
		t.Errorf("expected ST_GeomFromText, got: %s", sql)
	}
	if len(params) != 1 {
		t.Errorf("expected 1 param, got %d", len(params))
	}
}

func TestSpatialContainsSQL(t *testing.T) {
	info := setupTestPlace(t)
	query := queryset.NewQuery(info)

	pt := NewPoint(5, 5, WGS84)
	query.Where = queryset.Q(queryset.Lookup{"Region__contains": pt})
	sql, params := query.ToSQL()

	if !strings.Contains(sql, "ST_Contains") {
		t.Errorf("expected ST_Contains, got: %s", sql)
	}
	if len(params) != 1 {
		t.Errorf("expected 1 param, got %d", len(params))
	}
}

func TestSpatialDistanceLTESQL(t *testing.T) {
	info := setupTestPlace(t)
	query := queryset.NewQuery(info)

	pt := NewPoint(-73.9857, 40.7484, WGS84)
	dq := NewDistanceQuery(pt, 5000)
	query.Where = queryset.Q(queryset.Lookup{"Location__distance_lte": dq})
	sql, params := query.ToSQL()

	if !strings.Contains(sql, "ST_DWithin") {
		t.Errorf("expected ST_DWithin, got: %s", sql)
	}
	if !strings.Contains(sql, "ST_GeomFromText") {
		t.Errorf("expected ST_GeomFromText, got: %s", sql)
	}
	if len(params) != 2 {
		t.Errorf("expected 2 params (WKT + distance), got %d", len(params))
	}
	if params[1].(float64) != 5000 {
		t.Errorf("expected distance 5000, got %v", params[1])
	}
}

func TestSpatialDistanceGTSQL(t *testing.T) {
	info := setupTestPlace(t)
	query := queryset.NewQuery(info)

	pt := NewPoint(-73.9857, 40.7484, WGS84)
	dq := NewDistanceQuery(pt, 10000)
	query.Where = queryset.Q(queryset.Lookup{"Location__distance_gt": dq})
	sql, params := query.ToSQL()

	if !strings.Contains(sql, "NOT ST_DWithin") {
		t.Errorf("expected NOT ST_DWithin, got: %s", sql)
	}
	if len(params) != 2 {
		t.Errorf("expected 2 params, got %d", len(params))
	}
}

func TestSpatialDisjointSQL(t *testing.T) {
	info := setupTestPlace(t)
	query := queryset.NewQuery(info)

	pt := NewPoint(-73.9857, 40.7484, WGS84)
	query.Where = queryset.Q(queryset.Lookup{"Location__disjoint": pt})
	sql, _ := query.ToSQL()

	if !strings.Contains(sql, "ST_Disjoint") {
		t.Errorf("expected ST_Disjoint, got: %s", sql)
	}
}

func TestSpatialOverlapsSQL(t *testing.T) {
	info := setupTestPlace(t)
	query := queryset.NewQuery(info)

	poly := NewPolygon(WGS84,
		NewPoint(0, 0, WGS84),
		NewPoint(10, 0, WGS84),
		NewPoint(10, 10, WGS84),
		NewPoint(0, 10, WGS84),
		NewPoint(0, 0, WGS84),
	)
	query.Where = queryset.Q(queryset.Lookup{"Region__overlaps": poly})
	sql, _ := query.ToSQL()

	if !strings.Contains(sql, "ST_Overlaps") {
		t.Errorf("expected ST_Overlaps, got: %s", sql)
	}
}

func TestSpatialTouchesSQL(t *testing.T) {
	info := setupTestPlace(t)
	query := queryset.NewQuery(info)

	poly := NewPolygon(WGS84,
		NewPoint(0, 0, WGS84),
		NewPoint(10, 0, WGS84),
		NewPoint(10, 10, WGS84),
		NewPoint(0, 10, WGS84),
		NewPoint(0, 0, WGS84),
	)
	query.Where = queryset.Q(queryset.Lookup{"Region__touches": poly})
	sql, _ := query.ToSQL()

	if !strings.Contains(sql, "ST_Touches") {
		t.Errorf("expected ST_Touches, got: %s", sql)
	}
}

func TestSpatialCoversSQL(t *testing.T) {
	info := setupTestPlace(t)
	query := queryset.NewQuery(info)

	pt := NewPoint(5, 5, WGS84)
	query.Where = queryset.Q(queryset.Lookup{"Region__covers": pt})
	sql, _ := query.ToSQL()

	if !strings.Contains(sql, "ST_Covers") {
		t.Errorf("expected ST_Covers, got: %s", sql)
	}
}

func TestSpatialEqualsSQL(t *testing.T) {
	info := setupTestPlace(t)
	query := queryset.NewQuery(info)

	pt := NewPoint(-73.9857, 40.7484, WGS84)
	query.Where = queryset.Q(queryset.Lookup{"Location__equals": pt})
	sql, _ := query.ToSQL()

	if !strings.Contains(sql, "ST_Equals") {
		t.Errorf("expected ST_Equals, got: %s", sql)
	}
}

func TestSpatialBBContainsSQL(t *testing.T) {
	info := setupTestPlace(t)
	query := queryset.NewQuery(info)

	poly := NewPolygon(WGS84,
		NewPoint(0, 0, WGS84),
		NewPoint(10, 0, WGS84),
		NewPoint(10, 10, WGS84),
		NewPoint(0, 10, WGS84),
		NewPoint(0, 0, WGS84),
	)
	query.Where = queryset.Q(queryset.Lookup{"Region__bbcontains": poly})
	sql, _ := query.ToSQL()

	if !strings.Contains(sql, "@") {
		t.Errorf("expected @ operator for bbcontains, got: %s", sql)
	}
}

func TestSpatialHelperFunctions(t *testing.T) {
	pt := NewPoint(1, 2, WGS84)

	lookup := IntersectsLookup("location", pt)
	if _, ok := lookup["location__intersects"]; !ok {
		t.Error("IntersectsLookup should produce location__intersects key")
	}

	lookup = WithinLookup("location", pt)
	if _, ok := lookup["location__within"]; !ok {
		t.Error("WithinLookup should produce location__within key")
	}

	lookup = ContainsLookup("region", pt)
	if _, ok := lookup["region__contains"]; !ok {
		t.Error("ContainsLookup should produce region__contains key")
	}

	dq := NewDistanceQuery(pt, 3000)
	lookup = DistanceLTELookup("location", dq)
	if _, ok := lookup["location__distance_lte"]; !ok {
		t.Error("DistanceLTELookup should produce location__distance_lte key")
	}
}

func TestSpatialQ(t *testing.T) {
	pt := NewPoint(1, 2, WGS84)
	q := SpatialQ("location", "intersects", pt)
	if q == nil {
		t.Fatal("SpatialQ should return non-nil QNode")
	}
	if _, ok := q.Filters["location__intersects"]; !ok {
		t.Error("SpatialQ should have location__intersects in filters")
	}
}

func TestSpatialQChaining(t *testing.T) {
	info := setupTestPlace(t)
	query := queryset.NewQuery(info)

	pt1 := NewPoint(1, 2, WGS84)
	pt2 := NewPoint(3, 4, WGS84)

	q := SpatialQ("Location", "intersects", pt1).Or(SpatialQ("Location", "within", pt2))
	query.Where = q
	sql, _ := query.ToSQL()

	if !strings.Contains(sql, "OR") {
		t.Errorf("expected OR in chained spatial Q, got: %s", sql)
	}
	if !strings.Contains(sql, "ST_Within") {
		t.Errorf("expected ST_Within in chained query, got: %s", sql)
	}
}

func TestDistanceQueryImplementsInterfaces(t *testing.T) {
	pt := NewPoint(1, 2, WGS84)
	dq := NewDistanceQuery(pt, 5000)

	if dq.WKT() != pt.WKT() {
		t.Error("DistanceQuery WKT should delegate to geometry")
	}
	if dq.GetSRID() != pt.GetSRID() {
		t.Error("DistanceQuery GetSRID should delegate to geometry")
	}
	if dq.DistanceMeters() != 5000 {
		t.Errorf("expected 5000, got %f", dq.DistanceMeters())
	}
}

func TestPointImplementsGeometryValue(t *testing.T) {
	pt := NewPoint(1, 2, WGS84)
	if pt.WKT() != pt.String() {
		t.Error("Point WKT should match String")
	}
	if pt.GetSRID() != WGS84 {
		t.Errorf("expected SRID %d, got %d", WGS84, pt.GetSRID())
	}
}

func TestLineStringImplementsGeometryValue(t *testing.T) {
	ls := NewLineString(WGS84,
		NewPoint(0, 0, WGS84),
		NewPoint(1, 1, WGS84),
	)
	if ls.WKT() != ls.String() {
		t.Error("LineString WKT should match String")
	}
	if ls.GetSRID() != WGS84 {
		t.Errorf("expected SRID %d, got %d", WGS84, ls.GetSRID())
	}
}

func TestPolygonImplementsGeometryValue(t *testing.T) {
	poly := NewPolygon(WGS84,
		NewPoint(0, 0, WGS84),
		NewPoint(10, 0, WGS84),
		NewPoint(10, 10, WGS84),
		NewPoint(0, 10, WGS84),
		NewPoint(0, 0, WGS84),
	)
	if poly.WKT() != poly.String() {
		t.Error("Polygon WKT should match String")
	}
	if poly.GetSRID() != WGS84 {
		t.Errorf("expected SRID %d, got %d", WGS84, poly.GetSRID())
	}
}

func TestSpatialIndexInMeta(t *testing.T) {
	info := setupTestPlace(t)
	if len(info.Meta.SpatialIndexes) != 1 {
		t.Fatalf("expected 1 spatial index, got %d", len(info.Meta.SpatialIndexes))
	}
	si := info.Meta.SpatialIndexes[0]
	if len(si.Fields) != 1 || si.Fields[0] != "location" {
		t.Errorf("expected spatial index on 'location', got %v", si.Fields)
	}
}

func TestPointFieldSRIDParsing(t *testing.T) {
	info := setupTestPlace(t)

	locField, ok := info.FieldByName["Location"]
	if !ok {
		t.Fatal("expected Location field")
	}
	if locField.Type != orm.PointField {
		t.Errorf("expected PointField, got %s", locField.Type)
	}
	if locField.Options.SRID != 4326 {
		t.Errorf("expected SRID 4326, got %d", locField.Options.SRID)
	}

	regionField, ok := info.FieldByName["Region"]
	if !ok {
		t.Fatal("expected Region field")
	}
	if regionField.Type != orm.PolygonField {
		t.Errorf("expected PolygonField, got %s", regionField.Type)
	}
}
