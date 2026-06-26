// Package gis provides spatial query integration helpers that bridge the GIS
// geometry types with the ORM queryset layer. These helpers produce
// orm.GeometryValue / orm.DistanceGeometryValue values suitable for use in
// queryset Filter / Exclude / Q lookups.
package gis

import (
	"github.com/pkshahid/JanGo/orm/queryset"
)

// SpatialFilter returns a queryset.Lookup map for a single spatial predicate.
// The field is the model field name, lookup is the spatial lookup type
// (e.g. "intersects", "within", "contains"), and geom is any GeometryValue
// (Point, LineString, Polygon) or a raw WKT string.
//
// Example:
//
//	qs.Filter(gis.SpatialFilter("location", "intersects", gis.NewPoint(lon, lat, gis.WGS84)))
func SpatialFilter(field, lookup string, geom any) queryset.Lookup {
	return queryset.Lookup{field + "__" + lookup: geom}
}

// IntersectsLookup is a convenience for field__intersects=<geom>.
func IntersectsLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "intersects", geom)
}

// WithinLookup is a convenience for field__within=<geom>.
func WithinLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "within", geom)
}

// ContainsLookup is a convenience for field__contains=<geom>.
func ContainsLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "contains", geom)
}

// DisjointLookup is a convenience for field__disjoint=<geom>.
func DisjointLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "disjoint", geom)
}

// OverlapsLookup is a convenience for field__overlaps=<geom>.
func OverlapsLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "overlaps", geom)
}

// TouchesLookup is a convenience for field__touches=<geom>.
func TouchesLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "touches", geom)
}

// CrossesLookup is a convenience for field__crosses=<geom>.
func CrossesLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "crosses", geom)
}

// CoversLookup is a convenience for field__covers=<geom>.
func CoversLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "covers", geom)
}

// CoveredByLookup is a convenience for field__coveredby=<geom>.
func CoveredByLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "coveredby", geom)
}

// EqualsLookup is a convenience for field__equals=<geom>.
func EqualsLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "equals", geom)
}

// DistanceLTELookup creates a distance_lte spatial lookup. The geom must
// implement orm.DistanceGeometryValue (e.g. *gis.DistanceQuery) or
// orm.GeometryValue (in which case distance defaults to 0).
//
// Example:
//
//	qs.Filter(gis.DistanceLTELookup("location", gis.NewDistanceQuery(gis.NewPoint(lon, lat, gis.WGS84), 5000)))
func DistanceLTELookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "distance_lte", geom)
}

// DistanceLTLookup creates a distance_lt spatial lookup.
func DistanceLTLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "distance_lt", geom)
}

// DistanceGTLookup creates a distance_gt spatial lookup.
func DistanceGTLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "distance_gt", geom)
}

// DistanceGTELookup creates a distance_gte spatial lookup.
func DistanceGTELookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "distance_gte", geom)
}

// BBContainsLookup is a convenience for field__bbcontains=<geom>.
func BBContainsLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "bbcontains", geom)
}

// BBOverlapsLookup is a convenience for field__bboverlaps=<geom>.
func BBOverlapsLookup(field string, geom any) queryset.Lookup {
	return SpatialFilter(field, "bboverlaps", geom)
}

// SpatialQ creates a QNode for spatial filtering, allowing combination with
// other Q objects via And / Or / Not.
//
// Example:
//
//	qs.FilterQ(gis.SpatialQ("location", "intersects", poly).Or(
//	    gis.SpatialQ("location", "within", bbox)))
func SpatialQ(field, lookup string, geom any) *queryset.QNode {
	return queryset.Q(SpatialFilter(field, lookup, geom))
}
