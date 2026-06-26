package admin

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/orm"
	"github.com/pkshahid/JanGo/orm/backends"
	"github.com/pkshahid/JanGo/orm/queryset"
	ormsignals "github.com/pkshahid/JanGo/orm/signals"
)

// AdminAction is a bulk action applied to a queryset of models.
type AdminAction func(admin *ModelAdmin, req *godjangohttp.Request, qs queryset.RawQuerySet[any]) godjangohttp.Response

// InlineModelAdmin represents a model to be edited inline.
type InlineModelAdmin struct {
	Model    any
	FkName   string
	Template string // "tabular" or "stacked"
	MinNum   int
	MaxNum   int
	Extra    int
	Fields   []string
	Readonly []string
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
	ModelInfo           *orm.ModelInfo
	ListDisplay         []string
	ListFilter          []any // field name strings or ListFilterer implementations
	SearchFields        []string
	Ordering            []string
	ReadonlyFields      []string
	Fields              []string
	Fieldsets           []Fieldset
	InlineModels        []InlineModelAdmin
	Actions             []AdminAction
	ListPerPage         int
	ListMaxShowAll      int
	ShowFullResultCount bool

	// ViewOnSite controls whether the "View on site" link appears in the
	// admin change form. When nil, the link is shown automatically if the
	// model implements orm.GetAbsoluteURLer. When set to a bool, it
	// overrides the auto-detection. When set to a func(any) string, the
	// function is called with the object to produce the URL.
	ViewOnSite any

	// Hooks
	SaveModel   func(req *godjangohttp.Request, obj any, form any, change bool)
	DeleteModel func(req *godjangohttp.Request, obj any)
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
		if change {
			updateObject(ma.ModelInfo, obj)
		} else {
			saveObject(ma.ModelInfo, obj)
		}
	}

	ma.DeleteModel = func(req *godjangohttp.Request, obj any) {
		deleteObject(ma.ModelInfo, obj)
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

// fieldInterface returns the value of a struct field suitable for passing as
// a SQL parameter. Nil pointers are returned as nil so the driver inserts NULL.
func fieldInterface(v reflect.Value, fieldName string) any {
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return nil
	}
	return field.Interface()
}

// setAutoTimestampFields sets auto_now_add / auto_now fields to time.Now()
// when appropriate.
func setAutoTimestampFields(v reflect.Value, info *orm.ModelInfo, forCreate bool) {
	for _, f := range info.Fields {
		if f.Options.AutoNowAdd && !forCreate {
			continue
		}
		if !f.Options.AutoNow && !f.Options.AutoNowAdd {
			continue
		}
		fieldVal := v.FieldByName(f.Name)
		if !fieldVal.IsValid() || !fieldVal.CanSet() {
			continue
		}
		if fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				fieldVal.Set(reflect.ValueOf(time.Now()))
			}
		} else if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
			if forCreate && fieldVal.IsZero() {
				fieldVal.Set(reflect.ValueOf(time.Now()))
			} else if f.Options.AutoNow {
				fieldVal.Set(reflect.ValueOf(time.Now()))
			}
		}
	}
}

// saveObject inserts a new record into the database using reflection,
// mirroring queryset.QuerySet[T].Create.
func saveObject(info *orm.ModelInfo, obj any) error {
	ormsignals.PreSave.Send(obj, map[string]any{"instance": obj, "created": true})

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	setAutoTimestampFields(v, info, true)

	var columns []string
	var placeholders []string
	var values []any

	for _, f := range info.Fields {
		if f.Type == orm.ManyToManyField {
			continue
		}
		if f.PrimaryKey && f.Options.AutoCreated {
			pkField := v.FieldByName(f.Name)
			if pkField.IsValid() && pkField.IsZero() {
				continue
			}
		}
		columns = append(columns, f.Column)
		placeholders = append(placeholders, "?")
		values = append(values, fieldInterface(v, f.Name))
	}

	sqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		info.Meta.DbTable,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	dbAlias := backends.RouteForWrite(info)
	backend, err := backends.GetBackend(dbAlias)
	if err != nil {
		return err
	}

	ctx := context.Background()
	result, err := backend.Execute(ctx, sqlStr, values...)
	if err != nil {
		return err
	}

	if info.PrimaryKey != nil && info.PrimaryKey.Options.AutoCreated {
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		pkField := v.FieldByName(info.PrimaryKey.Name)
		if pkField.IsValid() && pkField.CanSet() && pkField.Kind() == reflect.Uint64 {
			pkField.SetUint(uint64(id))
		}
	}

	ormsignals.PostSave.Send(obj, map[string]any{"instance": obj, "created": true})
	return nil
}

// updateObject updates an existing record by its primary key using reflection.
func updateObject(info *orm.ModelInfo, obj any) error {
	ormsignals.PreSave.Send(obj, map[string]any{"instance": obj, "created": false})

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	setAutoTimestampFields(v, info, false)

	var setClauses []string
	var values []any

	for _, f := range info.Fields {
		if f.Type == orm.ManyToManyField {
			continue
		}
		if f.PrimaryKey {
			continue
		}
		setClauses = append(setClauses, f.Column+" = ?")
		values = append(values, fieldInterface(v, f.Name))
	}

	pkField := info.PrimaryKey
	pkVal := fieldInterface(v, pkField.Name)

	sqlStr := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?",
		info.Meta.DbTable,
		strings.Join(setClauses, ", "),
		pkField.Column)
	values = append(values, pkVal)

	dbAlias := backends.RouteForWrite(info)
	backend, err := backends.GetBackend(dbAlias)
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = backend.Execute(ctx, sqlStr, values...)
	if err != nil {
		return err
	}

	ormsignals.PostSave.Send(obj, map[string]any{"instance": obj, "created": false})
	return nil
}

// deleteObject deletes a record by its primary key using reflection.
func deleteObject(info *orm.ModelInfo, obj any) error {
	ormsignals.PreDelete.Send(obj, map[string]any{"instance": obj})

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	pkField := info.PrimaryKey
	pkVal := fieldInterface(v, pkField.Name)

	sqlStr := fmt.Sprintf("DELETE FROM %s WHERE %s = ?",
		info.Meta.DbTable,
		pkField.Column)

	dbAlias := backends.RouteForWrite(info)
	backend, err := backends.GetBackend(dbAlias)
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = backend.Execute(ctx, sqlStr, pkVal)
	if err != nil {
		return err
	}

	ormsignals.PostDelete.Send(obj, map[string]any{"instance": obj})
	return nil
}

// ViewOnSiteURL returns the "view on site" URL for the given object, or
// empty string if the link should not be shown. The resolution follows
// Django's semantics:
//   - If ViewOnSite is a func(any) string, it is called with obj.
//   - If ViewOnSite is a bool, it enables/disables auto-detection via
//     orm.GetAbsoluteURLer.
//   - If ViewOnSite is nil, auto-detection via orm.GetAbsoluteURLer is used.
func (ma *ModelAdmin) ViewOnSiteURL(obj any) string {
	switch v := ma.ViewOnSite.(type) {
	case func(any) string:
		return v(obj)
	case bool:
		if !v {
			return ""
		}
	}
	// nil or bool true → auto-detect
	url, ok := orm.GetAbsoluteURL(obj)
	if !ok {
		return ""
	}
	return url
}
