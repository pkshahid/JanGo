package admin

import (
	"fmt"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/orm"
	"github.com/pkshahid/JanGo/orm/queryset"
)

// AdminAction is a bulk action applied to a queryset of models.
type AdminAction func(admin *ModelAdmin, req *godjangohttp.Request, qs queryset.RawQuerySet[any]) godjangohttp.Response

// InlineModelAdmin represents a model to be edited inline.
type InlineModelAdmin struct {
	Model     any
	FkName    string
	Template  string // "tabular" or "stacked"
	MinNum    int
	MaxNum    int
	Extra     int
	Fields    []string
	Readonly  []string
}

// Fieldset represents a group of fields in a change form.
type Fieldset struct {
	Name        string
	Description string
	Fields      []string
	Classes     []string // e.g., "collapse"
}

// ModelAdmin provides configuration for how a model is presented in the admin interface.
type ModelAdmin struct {
	ModelInfo       *orm.ModelInfo
	ListDisplay     []string
	ListFilter      []string
	SearchFields    []string
	Ordering        []string
	ReadonlyFields  []string
	Fields          []string
	Fieldsets       []Fieldset
	InlineModels    []InlineModelAdmin
	Actions         []AdminAction
	ListPerPage     int
	ListMaxShowAll  int
	ShowFullResultCount bool

	// Hooks
	SaveModel       func(req *godjangohttp.Request, obj any, form any, change bool)
	DeleteModel     func(req *godjangohttp.Request, obj any)
}

// NewModelAdmin creates a new ModelAdmin with sane defaults.
func NewModelAdmin(model any) (*ModelAdmin, error) {
	info, err := orm.GetModelInfo(model)
	if err != nil {
		return nil, fmt.Errorf("admin: cannot register unregistered model %T: %v", model, err)
	}

	ma := &ModelAdmin{
		ModelInfo:           info,
		ListDisplay:         []string{"ID"},
		ListPerPage:         100,
		ListMaxShowAll:      200,
		ShowFullResultCount: true,
	}

	// Default hooks
	ma.SaveModel = func(req *godjangohttp.Request, obj any, form any, change bool) {
		// In a real framework, this would call obj.Save()
		// qs := queryset.NewQuerySet[...]
		// if change { qs.Update(...) } else { qs.Create(obj) }
	}

	ma.DeleteModel = func(req *godjangohttp.Request, obj any) {
		// In a real framework, this would call obj.Delete()
	}

	// Add built-in actions
	ma.Actions = append(ma.Actions, deleteSelectedAction)

	return ma, nil
}

func deleteSelectedAction(admin *ModelAdmin, req *godjangohttp.Request, qs queryset.RawQuerySet[any]) godjangohttp.Response {
	// A real implementation queries and deletes the selected IDs.
	// We'll mock the action execution via form POST handling.
	return nil
}
