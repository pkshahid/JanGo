package queryset

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkshahid/JanGo/orm"
)

// knownLookupTypes is the set of Django-style lookup suffixes that can appear
// as the last segment in a __-separated lookup key.
var knownLookupTypes = map[string]bool{
	"exact": true, "iexact": true, "contains": true, "icontains": true,
	"startswith": true, "istartswith": true, "endswith": true, "iendswith": true,
	"gt": true, "gte": true, "lt": true, "lte": true,
	"in": true, "isnull": true, "range": true, "regex": true, "iregex": true,
	"date": true, "year": true, "month": true, "day": true,
	"week": true, "week_day": true, "time": true, "hour": true,
	"minute": true, "second": true,
	// Spatial lookups (PostGIS / SpatiaLite)
	"contains_properly": true, "coveredby": true, "covers": true,
	"crosses": true, "disjoint": true, "distance_gt": true,
	"distance_gte": true, "distance_lt": true, "distance_lte": true,
	"equals": true, "intersects": true, "overlaps": true,
	"relate": true, "same_as": true, "touches": true, "within": true,
	"left": true, "right": true, "above": true, "below": true,
	"bbcontains": true, "bboverlaps": true, "strictly_above": true,
	"strictly_below": true,
}

// joinInfo represents a single LEFT JOIN in a query.
type joinInfo struct {
	alias       string         // table alias (e.g. "T1")
	table       string         // actual table name
	fromAlias   string         // alias of the source table
	fromColumn  string         // FK column on the source table
	toColumn    string         // PK column on the target table
	relatedInfo *orm.ModelInfo // ModelInfo of the joined model
	fkField     *orm.Field     // the FK field that created this join
	path        string         // full path from root (e.g. "author" or "author__profile")
}

// joinManager tracks JOINs for a single query, ensuring each relationship
// path is joined only once and producing consistent table aliases.
type joinManager struct {
	joins     []*joinInfo
	pathToJoin map[string]*joinInfo
	counter   int
	baseAlias string
	baseTable string
}

// newJoinManager creates a joinManager rooted at the given table.
func newJoinManager(baseTable string) *joinManager {
	return &joinManager{
		joins:      nil,
		pathToJoin: make(map[string]*joinInfo),
		baseAlias:  "T0",
		baseTable:  baseTable,
	}
}

// hasJoins returns true if at least one JOIN has been registered.
func (jm *joinManager) hasJoins() bool {
	return len(jm.joins) > 0
}

// resolveFKTarget resolves the target ModelInfo for a ForeignKey or
// OneToOneField by looking up the field's GoType in the registry.
func resolveFKTarget(field *orm.Field) (*orm.ModelInfo, error) {
	t := field.GoType
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("FK field %s has non-struct type %s", field.Name, t)
	}
	return orm.GetModelInfoByType(t)
}

// getOrCreateJoin creates a JOIN for the given path if it doesn't already
// exist, or returns the existing one. The path uniquely identifies the
// relationship chain from the base model (e.g. "author", "author__profile").
func (jm *joinManager) getOrCreateJoin(path, fromAlias string, fkField *orm.Field, fromInfo *orm.ModelInfo) *joinInfo {
	if existing, ok := jm.pathToJoin[path]; ok {
		return existing
	}
	relatedInfo, err := resolveFKTarget(fkField)
	if err != nil {
		return nil
	}
	fromColumn := fkField.Column
	toColumn := "id"
	if relatedInfo.PrimaryKey != nil {
		toColumn = relatedInfo.PrimaryKey.Column
	}
	jm.counter++
	alias := fmt.Sprintf("T%d", jm.counter)
	ji := &joinInfo{
		alias:       alias,
		table:       relatedInfo.Meta.DbTable,
		fromAlias:   fromAlias,
		fromColumn:  fromColumn,
		toColumn:    toColumn,
		relatedInfo: relatedInfo,
		fkField:     fkField,
		path:        path,
	}
	jm.joins = append(jm.joins, ji)
	jm.pathToJoin[path] = ji
	return ji
}

// fromClause generates the FROM clause including all JOINs.
// When no joins exist, returns just the base table name (no alias).
func (jm *joinManager) fromClause() string {
	if !jm.hasJoins() {
		return jm.baseTable
	}
	var sb strings.Builder
	sb.WriteString(jm.baseTable)
	sb.WriteString(" AS ")
	sb.WriteString(jm.baseAlias)
	for _, j := range jm.joins {
		sb.WriteString(" LEFT JOIN ")
		sb.WriteString(j.table)
		sb.WriteString(" AS ")
		sb.WriteString(j.alias)
		sb.WriteString(" ON ")
		sb.WriteString(j.fromAlias)
		sb.WriteByte('.')
		sb.WriteString(j.fromColumn)
		sb.WriteString(" = ")
		sb.WriteString(j.alias)
		sb.WriteByte('.')
		sb.WriteString(j.toColumn)
	}
	return sb.String()
}

// resolveFieldPath resolves a __-separated lookup key (e.g.
// "author__profile__bio__icontains") into a qualified column reference and
// the lookup type. It creates JOINs as needed for FK traversals.
//
// Returns (columnRef, lookup). When no joins are needed, columnRef is an
// unqualified column name (for backward compatibility).
func (jm *joinManager) resolveFieldPath(parts []string, info *orm.ModelInfo) (string, string) {
	fieldParts := parts
	lookup := "exact"
	if len(parts) > 1 && knownLookupTypes[parts[len(parts)-1]] {
		lookup = parts[len(parts)-1]
		fieldParts = parts[:len(parts)-1]
	}

	currentInfo := info
	currentAlias := jm.baseAlias
	traversed := false

	// Walk all parts except the last — each must be a FK/OneToOne.
	for i := 0; i < len(fieldParts)-1; i++ {
		field, ok := currentInfo.FieldByName[fieldParts[i]]
		if !ok || (field.Type != orm.ForeignKey && field.Type != orm.OneToOneField) {
			break
		}
		path := strings.Join(fieldParts[:i+1], "__")
		ji := jm.getOrCreateJoin(path, currentAlias, field, currentInfo)
		if ji == nil {
			break
		}
		currentAlias = ji.alias
		currentInfo = ji.relatedInfo
		traversed = true
	}

	// Resolve the final field part.
	colName := fieldParts[len(fieldParts)-1]
	if colName == "pk" && currentInfo != nil && currentInfo.PrimaryKey != nil {
		colName = currentInfo.PrimaryKey.Column
	} else if currentInfo != nil {
		if f, ok := currentInfo.FieldByName[colName]; ok {
			colName = f.Column
		}
	}

	if traversed || jm.hasJoins() {
		return currentAlias + "." + colName, lookup
	}
	return colName, lookup
}

// resolveOrderByField resolves an order-by field (possibly a __-separated
// path) into a qualified column reference. It creates JOINs as needed.
func (jm *joinManager) resolveOrderByField(field string, info *orm.ModelInfo) string {
	parts := strings.Split(field, "__")
	if len(parts) == 1 {
		// Simple field name — resolve to column, qualify if joins exist.
		colName := field
		if f, ok := info.FieldByName[field]; ok {
			colName = f.Column
		}
		if jm.hasJoins() {
			return jm.baseAlias + "." + colName
		}
		return colName
	}
	// Multi-part path — use resolveFieldPath which handles FK traversal.
	colRef, _ := jm.resolveFieldPath(parts, info)
	return colRef
}

// selectCol represents one entry in the SELECT column list.
type selectCol struct {
	expr string
}

// collectSelectRelated walks the model's FK/OneToOne fields up to the given
// depth and returns SELECT column expressions for joined tables.
func (jm *joinManager) collectSelectRelated(info *orm.ModelInfo, fromAlias, path string, depth int) []selectCol {
	if depth <= 0 {
		return nil
	}
	var cols []selectCol
	for _, field := range info.Fields {
		if field.Type != orm.ForeignKey && field.Type != orm.OneToOneField {
			continue
		}
		relatedInfo, err := resolveFKTarget(field)
		if err != nil {
			continue
		}
		joinPath := field.Name
		if path != "" {
			joinPath = path + "__" + field.Name
		}
		ji := jm.getOrCreateJoin(joinPath, fromAlias, field, info)
		if ji == nil {
			continue
		}
		// Add all non-M2M columns from the related table.
		for _, relField := range relatedInfo.Fields {
			if relField.Type == orm.ManyToManyField {
				continue
			}
			alias := joinPath + "__" + relField.Column
			cols = append(cols, selectCol{
				expr: fmt.Sprintf("%s.%s AS %s", ji.alias, relField.Column, alias),
			})
		}
		// Recurse into the related model.
		cols = append(cols, jm.collectSelectRelated(relatedInfo, ji.alias, joinPath, depth-1)...)
	}
	return cols
}
