package queryset

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkshahid/JanGo/orm"
	"github.com/pkshahid/JanGo/orm/backends"
	"github.com/pkshahid/JanGo/orm/signals"
)

// RefreshOptions controls the behaviour of RefreshFromDB.
type RefreshOptions struct {
	// Using specifies the database alias to read from.
	// If empty, the router determines the read database.
	Using string

	// Fields restricts the refresh to the named fields.
	// If empty, all database-backed fields are reloaded.
	Fields []string
}

// RefreshFromDB reloads field values from the database for an existing model
// instance, mirroring Django's Model.refresh_from_db().
//
// The instance must be a pointer to a registered model struct with a non-zero
// primary key. When opts.Fields is empty, every database-backed field is
// overwritten with the current database value. When opts.Fields lists specific
// field names, only those fields are refreshed; all other fields retain their
// in-memory values.
//
// The PostInit signal is sent after a successful refresh.
func RefreshFromDB(obj any, opts ...RefreshOptions) error {
	if obj == nil {
		return fmt.Errorf("orm: RefreshFromDB requires a non-nil pointer")
	}

	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("orm: RefreshFromDB requires a non-nil pointer to a struct, got %T", obj)
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("orm: RefreshFromDB requires a pointer to a struct, got %T", obj)
	}

	info, err := orm.GetModelInfo(obj)
	if err != nil {
		return err
	}

	if info.PrimaryKey == nil {
		return fmt.Errorf("orm: model %s has no primary key", info.Name)
	}

	// Read the PK value from the struct.
	pkField, ok := info.Type.FieldByName(info.PrimaryKey.Name)
	if !ok {
		return fmt.Errorf("orm: primary key field %s not found on %s", info.PrimaryKey.Name, info.Name)
	}
	pkVal := v.FieldByIndex(pkField.Index)
	if pkVal.IsZero() {
		return fmt.Errorf("orm: cannot refresh %s with zero primary key", info.Name)
	}

	// Determine which fields to refresh.
	var refreshFields []*orm.Field
	if len(opts) > 0 && len(opts[0].Fields) > 0 {
		for _, name := range opts[0].Fields {
			f, ok := info.FieldByName[name]
			if !ok {
				return fmt.Errorf("orm: unknown field %q on model %s", name, info.Name)
			}
			refreshFields = append(refreshFields, f)
		}
	} else {
		refreshFields = info.Fields
	}

	// Build SELECT clause.
	var selectCols []string
	for _, f := range refreshFields {
		selectCols = append(selectCols, f.Column)
	}

	// Build the SQL.
	table := info.Meta.DbTable
	pkCol := info.PrimaryKey.Column
	sqlStr := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?",
		strings.Join(selectCols, ", "), table, pkCol)

	// Resolve the database alias.
	dbAlias := ""
	if len(opts) > 0 {
		dbAlias = opts[0].Using
	}
	if dbAlias == "" {
		dbAlias = backends.RouteForRead(info)
	}

	backend, err := backends.GetBackend(dbAlias)
	if err != nil {
		return err
	}

	ctx := context.Background()
	row := backend.QueryRow(ctx, sqlStr, pkVal.Interface())

	// Build scan targets: one pointer per refreshed field.
	scanTargets := make([]any, len(refreshFields))
	for i, f := range refreshFields {
		sf, ok := info.Type.FieldByName(f.Name)
		if !ok {
			return fmt.Errorf("orm: field %s not found on struct %s", f.Name, info.Name)
		}
		scanTargets[i] = v.FieldByIndex(sf.Index).Addr().Interface()
	}

	if err := row.Scan(scanTargets...); err != nil {
		return fmt.Errorf("orm: refresh_from_db failed for %s (pk=%v): %w", info.Name, pkVal.Interface(), err)
	}

	// Fire PostInit, mirroring Django which sends post_init after refresh.
	signals.PostInit.Send(obj, map[string]any{"instance": obj})

	return nil
}
