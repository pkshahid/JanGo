package admin

import (
	"fmt"
	"strings"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/orm"
)

// FilterChoice represents a single clickable filter option in the admin
// changelist sidebar. Django calls these "choices" or "lookups".
type FilterChoice struct {
	// Display is the human-readable label shown in the sidebar.
	Display string
	// Query is the query string value appended to the URL when this
	// choice is selected (e.g. "?is_published=true").
	Query string
	// Selected is true when the current request matches this choice.
	Selected bool
}

// ListFilterer is the interface that all list filters implement. It is
// satisfied by SimpleListFilter and any custom filter type.
//
// Django's equivalent is the admin.SimpleListFilter / admin.ListFilter
// class hierarchy.
type ListFilterer interface {
	// Title returns the human-readable heading for the filter section.
	Title() string
	// Choices returns the available filter options for the current request.
	Choices(req *godjangohttp.Request) []FilterChoice
	// Queryset applies the filter to the given query parameters and
	// returns a modified WHERE clause fragment and args. If the filter
	// is not active (no matching query param), it returns empty strings
	// and nil args.
	Queryset(req *godjangohttp.Request, info *orm.ModelInfo) (string, []any)
}

// SimpleListFilter is a convenience type for creating custom filters
// without defining a full struct. It mirrors Django's SimpleListFilter.
//
// Usage:
//
//	filter := &admin.SimpleListFilter{
//	    FilterTitle: "Published status",
//	    ParameterName: "published",
//	    Lookups: []admin.FilterLookup{
//	        {"Published", "yes"},
//	        {"Unpublished", "no"},
//	    },
//	    QuerysetFn: func(val string, info *orm.ModelInfo) (string, []any) {
//	        if val == "yes" {
//	            return "IsPublished = ?", []any{true}
//	        }
//	        if val == "no" {
//	            return "IsPublished = ?", []any{false}
//	        }
//	        return "", nil
//	    },
//	}
type SimpleListFilter struct {
	FilterTitle   string
	ParameterName string
	Lookups       []FilterLookup
	QuerysetFn    func(val string, info *orm.ModelInfo) (string, []any)
}

// FilterLookup defines a display/value pair for SimpleListFilter.
type FilterLookup struct {
	// Display is the human-readable label.
	Display string
	// Value is the query string value used when this lookup is selected.
	Value string
}

func (f *SimpleListFilter) Title() string { return f.FilterTitle }

func (f *SimpleListFilter) Choices(req *godjangohttp.Request) []FilterChoice {
	current := req.GET.Get(f.ParameterName)
	choices := make([]FilterChoice, 0, len(f.Lookups)+1)

	// "All" option (clears the filter)
	allQuery := removeQueryParam(req.URL.RawQuery, f.ParameterName)
	choices = append(choices, FilterChoice{
		Display:  "All",
		Query:    allQuery,
		Selected: current == "",
	})

	for _, lookup := range f.Lookups {
		q := setQueryParam(req.URL.RawQuery, f.ParameterName, lookup.Value)
		choices = append(choices, FilterChoice{
			Display:  lookup.Display,
			Query:    q,
			Selected: current == lookup.Value,
		})
	}
	return choices
}

func (f *SimpleListFilter) Queryset(req *godjangohttp.Request, info *orm.ModelInfo) (string, []any) {
	val := req.GET.Get(f.ParameterName)
	if val == "" || f.QuerysetFn == nil {
		return "", nil
	}
	return f.QuerysetFn(val, info)
}

// FieldListFilter is the automatic filter created when ListFilter contains
// a plain field name string. It introspects the model field to generate
// choices based on distinct values (for boolean and choice fields) or
// date hierarchies (for date fields).
type FieldListFilter struct {
	FieldName string
	fieldInfo *orm.Field
}

// NewFieldListFilter creates a FieldListFilter for the given field name
// and model info. If the field is not found, Title() returns the field
// name as-is and Choices/Queryset are no-ops.
func NewFieldListFilter(fieldName string, info *orm.ModelInfo) *FieldListFilter {
	f := &FieldListFilter{FieldName: fieldName}
	if info != nil {
		if fi, ok := info.FieldByName[fieldName]; ok {
			f.fieldInfo = fi
		}
	}
	return f
}

func (f *FieldListFilter) Title() string {
	return f.FieldName
}

func (f *FieldListFilter) Choices(req *godjangohttp.Request) []FilterChoice {
	param := strings.ToLower(f.FieldName)
	current := req.GET.Get(param)

	choices := make([]FilterChoice, 0, 3)

	// "All" option
	choices = append(choices, FilterChoice{
		Display:  "All",
		Query:    removeQueryParam(req.URL.RawQuery, param),
		Selected: current == "",
	})

	// For boolean fields, show Yes/No
	if f.fieldInfo != nil && f.fieldInfo.Type == orm.BooleanField {
		choices = append(choices,
			FilterChoice{
				Display:  "Yes",
				Query:    setQueryParam(req.URL.RawQuery, param, "true"),
				Selected: current == "true",
			},
			FilterChoice{
				Display:  "No",
				Query:    setQueryParam(req.URL.RawQuery, param, "false"),
				Selected: current == "false",
			},
		)
	}

	return choices
}

func (f *FieldListFilter) Queryset(req *godjangohttp.Request, info *orm.ModelInfo) (string, []any) {
	param := strings.ToLower(f.FieldName)
	val := req.GET.Get(param)
	if val == "" {
		return "", nil
	}

	// Resolve the actual column name
	col := f.FieldName
	if f.fieldInfo != nil {
		col = f.fieldInfo.Column
	}

	// Handle boolean field values
	if f.fieldInfo != nil && f.fieldInfo.Type == orm.BooleanField {
		if val == "true" {
			return fmt.Sprintf("%s = ?", col), []any{true}
		}
		if val == "false" {
			return fmt.Sprintf("%s = ?", col), []any{false}
		}
	}

	return fmt.Sprintf("%s = ?", col), []any{val}
}

// --- query string helpers ---

func setQueryParam(rawQuery, key, val string) string {
	pairs := parseQueryPairs(rawQuery)
	pairs[key] = val
	return encodeQueryPairs(pairs)
}

func removeQueryParam(rawQuery, key string) string {
	pairs := parseQueryPairs(rawQuery)
	delete(pairs, key)
	return encodeQueryPairs(pairs)
}

func parseQueryPairs(rawQuery string) map[string]string {
	result := make(map[string]string)
	if rawQuery == "" {
		return result
	}
	for _, part := range strings.Split(rawQuery, "&") {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		} else {
			result[kv[0]] = ""
		}
	}
	return result
}

func encodeQueryPairs(pairs map[string]string) string {
	if len(pairs) == 0 {
		return ""
	}
	var parts []string
	for k, v := range pairs {
		if v == "" {
			parts = append(parts, k)
		} else {
			parts = append(parts, k+"="+v)
		}
	}
	return strings.Join(parts, "&")
}

// FilterSpec is the template-ready representation of a resolved filter.
type FilterSpec struct {
	Title     string
	Choices   []FilterChoice
	HasActive bool
}

func hasActiveChoice(choices []FilterChoice) bool {
	for _, c := range choices {
		if c.Selected && c.Display != "All" {
			return true
		}
	}
	return false
}

// ResolveFilters converts the ListFilter entries on a ModelAdmin into
// concrete ListFilterer instances. String entries become FieldListFilter;
// entries already implementing ListFilterer are passed through.
func ResolveFilters(ma *ModelAdmin) []ListFilterer {
	if len(ma.ListFilter) == 0 {
		return nil
	}
	filters := make([]ListFilterer, 0, len(ma.ListFilter))
	for _, entry := range ma.ListFilter {
		switch v := entry.(type) {
		case string:
			filters = append(filters, NewFieldListFilter(v, ma.ModelInfo))
		case ListFilterer:
			filters = append(filters, v)
		default:
			// Unknown type — skip
		}
	}
	return filters
}
