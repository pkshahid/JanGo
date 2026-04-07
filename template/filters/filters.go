package filters

import (
	godjango "github.com/godjango/godjango/template"
)

// RegisterFilters registers all core Django template filters.
func RegisterFilters(lib *godjango.Library) {
	RegisterMathLogicFilters(lib)
}
