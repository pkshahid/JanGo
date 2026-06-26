package queryset

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/pkshahid/JanGo/orm"
	"github.com/pkshahid/JanGo/orm/backends"
	ormsignals "github.com/pkshahid/JanGo/orm/signals"
)

// getBackend resolves the database backend for this query, using the
// explicitly set Database alias or the router.
func (qs QuerySet[T]) getBackend(forWrite bool) (backends.Backend, error) {
	dbAlias := qs.query.Database
	if dbAlias == "" {
		if forWrite {
			dbAlias = backends.RouteForWrite(qs.query.ModelInfo)
		} else {
			dbAlias = backends.RouteForRead(qs.query.ModelInfo)
		}
	}
	return backends.GetBackend(dbAlias)
}

// buildScanDest creates a slice of scan destinations (pointers to struct fields)
// matching the column order returned by the database.
func buildScanDest(obj any, info *orm.ModelInfo, columns []string) []any {
	v := reflect.ValueOf(obj).Elem()
	dest := make([]any, len(columns))

	// Build a lookup from column name to field.
	colToField := make(map[string]*orm.Field, len(info.Fields))
	for _, f := range info.Fields {
		colToField[f.Column] = f
	}

	for i, col := range columns {
		f, ok := colToField[col]
		if !ok {
			var dummy any
			dest[i] = &dummy
			continue
		}
		structField := v.FieldByName(f.Name)
		if !structField.IsValid() || !structField.CanAddr() {
			var dummy any
			dest[i] = &dummy
			continue
		}
		dest[i] = structField.Addr().Interface()
	}
	return dest
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

// setFieldValue assigns val to the struct field via reflection, performing
// type conversion when necessary.
func setFieldValue(field reflect.Value, val any) {
	if !field.CanSet() || val == nil {
		return
	}
	valV := reflect.ValueOf(val)
	if valV.Type() == field.Type() {
		field.Set(valV)
		return
	}
	if valV.Type().ConvertibleTo(field.Type()) {
		field.Set(valV.Convert(field.Type()))
	}
}

// All executes the query and returns all matched records.
func (qs QuerySet[T]) All() ([]T, error) {
	sqlStr, params := qs.query.ToSQL()

	backend, err := qs.getBackend(false)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	rows, err := backend.Query(ctx, sqlStr, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := []T{}
	for rows.Next() {
		var obj T
		scanDest := buildScanDest(&obj, qs.query.ModelInfo, columns)
		if err := rows.Scan(scanDest...); err != nil {
			return nil, err
		}
		results = append(results, obj)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Handle PrefetchRelated concurrently via goroutines.
	// Each prefetch lookup triggers a separate query to fetch related objects.
	if len(qs.query.PrefetchRelated) > 0 && len(results) > 0 {
		var wg sync.WaitGroup
		errChan := make(chan error, len(qs.query.PrefetchRelated))

		for _, prefetch := range qs.query.PrefetchRelated {
			wg.Add(1)
			go func(lookup string) {
				defer wg.Done()
				_ = lookup
				// In a full implementation, this would:
				// 1. Resolve the relationship metadata for the lookup.
				// 2. Collect PKs from results.
				// 3. Execute: SELECT * FROM related_table WHERE fk IN (...).
				// 4. Map related objects back onto results.
			}(prefetch)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			if err != nil {
				return nil, fmt.Errorf("prefetch error: %v", err)
			}
		}
	}

	return results, nil
}

// Get executes the query and returns exactly one record.
// Errors if 0 or >1 records match.
func (qs QuerySet[T]) Get(lookups ...Lookup) (T, error) {
	var zero T
	c := qs.Filter(lookups...)
	c = c.Limit(2) // We limit to 2 to check for MultipleObjectsReturned

	results, err := c.All()
	if err != nil {
		return zero, err
	}

	if len(results) == 0 {
		return zero, fmt.Errorf("orm: DoesNotExist")
	}
	if len(results) > 1 {
		return zero, fmt.Errorf("orm: MultipleObjectsReturned")
	}

	return results[0], nil
}

// First returns the first object matched by the query.
func (qs QuerySet[T]) First() (*T, error) {
	c := qs.clone()
	// If no ordering is specified, Django defaults to PK ASC.
	if len(c.query.OrderBy) == 0 && c.query.ModelInfo.PrimaryKey != nil {
		c = c.OrderBy(c.query.ModelInfo.PrimaryKey.Name)
	}
	c = c.Limit(1)

	results, err := c.All()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil // Return nil, nil for First/Last
	}
	return &results[0], nil
}

// Last returns the last object matched by the query.
func (qs QuerySet[T]) Last() (*T, error) {
	c := qs.clone()
	if len(c.query.OrderBy) == 0 && c.query.ModelInfo.PrimaryKey != nil {
		c = c.OrderBy("-" + c.query.ModelInfo.PrimaryKey.Name)
	} else {
		// Reverse the existing ordering
		var reversed []string
		for _, o := range c.query.OrderBy {
			if strings.HasPrefix(o, "-") {
				reversed = append(reversed, o[1:])
			} else {
				reversed = append(reversed, "-"+o)
			}
		}
		c.query.OrderBy = reversed
	}
	c = c.Limit(1)

	results, err := c.All()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}

// Exists returns true if the QuerySet contains any results.
func (qs QuerySet[T]) Exists() (bool, error) {
	c := qs.clone()
	c.query.OnlyFields = []string{"1"}
	c = c.Limit(1)
	c.query.OrderBy = nil // Order by doesn't matter for exists

	sqlStr, params := c.query.ToSQL()

	backend, err := c.getBackend(false)
	if err != nil {
		return false, err
	}

	ctx := context.Background()
	rows, err := backend.Query(ctx, sqlStr, params...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	hasNext := rows.Next()
	if err := rows.Err(); err != nil {
		return false, err
	}
	return hasNext, nil
}

// Count returns the number of records as an int64.
func (qs QuerySet[T]) Count() (int64, error) {
	c := qs.clone()
	c.query.OrderBy = nil // Remove order by for count

	sqlStr, params := c.query.ToSQL()
	countSQL := "SELECT COUNT(*) FROM (" + sqlStr + ") AS subquery"

	backend, err := c.getBackend(false)
	if err != nil {
		return 0, err
	}

	ctx := context.Background()
	var count int64
	if err := backend.QueryRow(ctx, countSQL, params...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Values returns dictionaries (maps) rather than model instances.
func (qs QuerySet[T]) Values(fields ...string) ([]map[string]any, error) {
	c := qs.Only(fields...)

	sqlStr, params := c.query.ToSQL()

	backend, err := c.getBackend(false)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	rows, err := backend.Query(ctx, sqlStr, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := []map[string]any{}
	for rows.Next() {
		scanDest := make([]any, len(columns))
		for i := range scanDest {
			scanDest[i] = new(any)
		}
		if err := rows.Scan(scanDest...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(columns))
		for i, col := range columns {
			row[col] = *(scanDest[i].(*any))
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// ValuesList returns tuples (slices) rather than model instances.
func (qs QuerySet[T]) ValuesList(fields ...string) ([][]any, error) {
	c := qs.Only(fields...)

	sqlStr, params := c.query.ToSQL()

	backend, err := c.getBackend(false)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	rows, err := backend.Query(ctx, sqlStr, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := [][]any{}
	for rows.Next() {
		scanDest := make([]any, len(columns))
		for i := range scanDest {
			scanDest[i] = new(any)
		}
		if err := rows.Scan(scanDest...); err != nil {
			return nil, err
		}
		row := make([]any, len(columns))
		for i := range scanDest {
			row[i] = *(scanDest[i].(*any))
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// Create inserts a new record into the database.
func (qs QuerySet[T]) Create(obj *T) error {
	// Trigger PreSave
	ormsignals.PreSave.Send(obj, map[string]any{"instance": obj, "created": true})

	info := qs.query.ModelInfo
	v := reflect.ValueOf(obj).Elem()

	// Set auto_now_add / auto_now fields to time.Now() if needed.
	for _, f := range info.Fields {
		if f.Options.AutoNowAdd || f.Options.AutoNow {
			fieldVal := v.FieldByName(f.Name)
			if fieldVal.IsValid() && fieldVal.CanSet() {
				if fieldVal.Kind() == reflect.Ptr {
					if fieldVal.IsNil() {
						fieldVal.Set(reflect.ValueOf(time.Now()))
					}
				} else if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
					if fieldVal.IsZero() {
						fieldVal.Set(reflect.ValueOf(time.Now()))
					}
				}
			}
		}
	}

	var columns []string
	var placeholders []string
	var values []any

	for _, f := range info.Fields {
		// Skip ManyToMany fields — they're resolved through junction tables, not columns.
		if f.Type == orm.ManyToManyField {
			continue
		}
		// Skip auto-increment PK when it has a zero value (let DB assign it).
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

	backend, err := qs.getBackend(true)
	if err != nil {
		return err
	}

	ctx := context.Background()
	result, err := backend.Execute(ctx, sqlStr, values...)
	if err != nil {
		return err
	}

	// Set auto-increment PK back on the struct.
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

	// Trigger PostSave
	ormsignals.PostSave.Send(obj, map[string]any{"instance": obj, "created": true})
	return nil
}

// Update updates the records matched by the query with the given fields.
// Values that implement Expression (e.g. F) are resolved to SQL column references.
func (qs QuerySet[T]) Update(fields map[string]any) (int64, error) {
	sqlStr, params := qs.query.ToUpdateSQL(fields)

	backend, err := qs.getBackend(true)
	if err != nil {
		return 0, err
	}

	ctx := context.Background()
	result, err := backend.Execute(ctx, sqlStr, params...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Delete deletes the records matched by the query.
func (qs QuerySet[T]) Delete() (int64, error) {
	var zero T
	ormsignals.PreDelete.Send(zero, map[string]any{"instance": nil})

	info := qs.query.ModelInfo
	sqlStr := fmt.Sprintf("DELETE FROM %s", info.Meta.DbTable)

	var whereClauses []string
	var params []any

	if qs.query.Where != nil {
		clause, p := qs.query.Where.toSQL(info)
		if clause != "" {
			whereClauses = append(whereClauses, "("+clause+")")
			params = append(params, p...)
		}
	}

	if qs.query.Exclude != nil {
		clause, p := qs.query.Exclude.toSQL(info)
		if clause != "" {
			whereClauses = append(whereClauses, "NOT ("+clause+")")
			params = append(params, p...)
		}
	}

	if len(whereClauses) > 0 {
		sqlStr += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	backend, err := qs.getBackend(true)
	if err != nil {
		ormsignals.PostDelete.Send(zero, map[string]any{"instance": nil})
		return 0, err
	}

	ctx := context.Background()
	result, err := backend.Execute(ctx, sqlStr, params...)
	if err != nil {
		ormsignals.PostDelete.Send(zero, map[string]any{"instance": nil})
		return 0, err
	}

	ormsignals.PostDelete.Send(zero, map[string]any{"instance": nil})
	return result.RowsAffected()
}

// GetOrCreate gets an object, or creates it if it doesn't exist.
func (qs QuerySet[T]) GetOrCreate(kwargs Lookup, defaults map[string]any) (T, bool, error) {
	var zero T
	obj, err := qs.Get(kwargs)
	if err == nil {
		return obj, false, nil // Found
	}

	if err.Error() == "orm: DoesNotExist" {
		var newObj T
		v := reflect.ValueOf(&newObj).Elem()
		info := qs.query.ModelInfo

		for key, val := range kwargs {
			if field, ok := info.FieldByName[key]; ok {
				setFieldValue(v.FieldByName(field.Name), val)
			}
		}
		for key, val := range defaults {
			if field, ok := info.FieldByName[key]; ok {
				setFieldValue(v.FieldByName(field.Name), val)
			}
		}

		if err := qs.Create(&newObj); err != nil {
			return zero, false, err
		}
		return newObj, true, nil
	}

	return zero, false, err
}

// UpdateOrCreate updates an object, or creates it if it doesn't exist.
func (qs QuerySet[T]) UpdateOrCreate(kwargs Lookup, defaults map[string]any) (T, bool, error) {
	var zero T
	obj, err := qs.Get(kwargs)
	if err == nil {
		// Update existing
		_, err = qs.Filter(kwargs).Update(defaults)
		if err != nil {
			return zero, false, err
		}
		// Re-fetch the updated object
		obj, err = qs.Get(kwargs)
		if err != nil {
			return zero, false, err
		}
		return obj, false, nil
	}

	if err.Error() == "orm: DoesNotExist" {
		var newObj T
		v := reflect.ValueOf(&newObj).Elem()
		info := qs.query.ModelInfo

		for key, val := range kwargs {
			if field, ok := info.FieldByName[key]; ok {
				setFieldValue(v.FieldByName(field.Name), val)
			}
		}
		for key, val := range defaults {
			if field, ok := info.FieldByName[key]; ok {
				setFieldValue(v.FieldByName(field.Name), val)
			}
		}

		if err := qs.Create(&newObj); err != nil {
			return zero, false, err
		}
		return newObj, true, nil
	}

	return zero, false, err
}

// BulkCreate inserts multiple records efficiently.
func (qs QuerySet[T]) BulkCreate(objs []T) error {
	if len(objs) == 0 {
		return nil
	}

	info := qs.query.ModelInfo

	var columns []string
	var placeholders []string
	for _, f := range info.Fields {
		if f.Type == orm.ManyToManyField {
			continue
		}
		if f.PrimaryKey && f.Options.AutoCreated {
			continue // Skip auto-increment PK
		}
		columns = append(columns, f.Column)
		placeholders = append(placeholders, "?")
	}

	placeholderGroup := "(" + strings.Join(placeholders, ", ") + ")"
	var valueStrings []string
	var values []any

	for i := range objs {
		v := reflect.ValueOf(&objs[i]).Elem()

		// Set auto_now_add / auto_now fields.
		for _, f := range info.Fields {
			if f.Options.AutoNowAdd || f.Options.AutoNow {
				fieldVal := v.FieldByName(f.Name)
				if fieldVal.IsValid() && fieldVal.CanSet() {
					if fieldVal.Kind() == reflect.Ptr {
						if fieldVal.IsNil() {
							fieldVal.Set(reflect.ValueOf(time.Now()))
						}
					} else if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
						if fieldVal.IsZero() {
							fieldVal.Set(reflect.ValueOf(time.Now()))
						}
					}
				}
			}
		}

		valueStrings = append(valueStrings, placeholderGroup)
		for _, f := range info.Fields {
			if f.Type == orm.ManyToManyField {
				continue
			}
			if f.PrimaryKey && f.Options.AutoCreated {
				continue
			}
			values = append(values, fieldInterface(v, f.Name))
		}
	}

	sqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		info.Meta.DbTable,
		strings.Join(columns, ", "),
		strings.Join(valueStrings, ", "))

	backend, err := qs.getBackend(true)
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = backend.Execute(ctx, sqlStr, values...)
	return err
}

// BulkUpdate updates multiple specific fields on a slice of objects efficiently.
func (qs QuerySet[T]) BulkUpdate(objs []T, fields []string) error {
	if len(objs) == 0 {
		return nil
	}

	info := qs.query.ModelInfo
	if info.PrimaryKey == nil {
		return fmt.Errorf("orm: model %s has no primary key", info.Name)
	}

	backend, err := qs.getBackend(true)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Build the SET clause column names.
	var setClauses []string
	var updateFields []*orm.Field
	for _, fieldName := range fields {
		field, ok := info.FieldByName[fieldName]
		if !ok {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", field.Column))
		updateFields = append(updateFields, field)
	}
	if len(updateFields) == 0 {
		return nil
	}

	for i := range objs {
		v := reflect.ValueOf(&objs[i]).Elem()

		var params []any
		for _, f := range updateFields {
			params = append(params, fieldInterface(v, f.Name))
		}
		params = append(params, fieldInterface(v, info.PrimaryKey.Name))

		sqlStr := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?",
			info.Meta.DbTable,
			strings.Join(setClauses, ", "),
			info.PrimaryKey.Column)

		if _, err := backend.Execute(ctx, sqlStr, params...); err != nil {
			return err
		}
	}

	return nil
}
