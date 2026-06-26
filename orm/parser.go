package orm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// ModelInfo holds all metadata and parsed fields for a registered model.
type ModelInfo struct {
	Name        string
	AppLabel    string
	Type        reflect.Type
	Fields      []*Field
	FieldByName map[string]*Field
	Meta        *Meta
	PrimaryKey  *Field
}

// parseModel inspects a struct type and generates a ModelInfo.
func parseModel(m any) (*ModelInfo, error) {
	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("orm: model must be a struct, got %s", t.Kind())
	}

	info := &ModelInfo{
		Name:        t.Name(),
		Type:        t,
		Fields:      make([]*Field, 0),
		FieldByName: make(map[string]*Field),
		Meta:        &Meta{},
	}

	// Parse custom Meta if implemented
	if mi, ok := m.(ModelInterface); ok {
		info.Meta = mi.ModelMeta()
	}

	// Handle proxy models: use the parent model's table, no own table is created.
	if info.Meta.Proxy {
		if info.Meta.Abstract {
			return nil, fmt.Errorf("orm: model %s cannot be both proxy and abstract", info.Name)
		}
		parentType := findParentModelType(t)
		if parentType == nil {
			return nil, fmt.Errorf("orm: proxy model %s must embed a concrete parent model", info.Name)
		}
		parentTable, err := resolveModelDbTable(parentType)
		if err != nil {
			return nil, fmt.Errorf("orm: proxy model %s: %w", info.Name, err)
		}
		info.Meta.DbTable = parentTable
	} else if info.Meta.DbTable == "" {
		// Set default table name if not provided (e.g. app_modelname)
		info.Meta.DbTable = strings.ToLower(info.Name)
	}

	// Recursively parse fields
	err := parseStructFields(t, info)
	if err != nil {
		return nil, err
	}

	// Find PK. If a model defines its own primary key, it overrides the
	// auto_created PK inherited from the base Model (mirrors Django behavior).
	for _, f := range info.Fields {
		if f.PrimaryKey {
			if info.PrimaryKey != nil {
				// Allow override: non-auto-created PK supersedes auto-created PK.
				if info.PrimaryKey.Options.AutoCreated && !f.Options.AutoCreated {
					info.PrimaryKey.PrimaryKey = false
					info.PrimaryKey = f
				} else if !info.PrimaryKey.Options.AutoCreated && f.Options.AutoCreated {
					// Keep existing non-auto PK, skip the auto-created one.
					f.PrimaryKey = false
				} else {
					return nil, fmt.Errorf("orm: multiple primary keys defined on %s", info.Name)
				}
			} else {
				info.PrimaryKey = f
			}
		}
	}

	// Fallback to searching for ID if no explicit PK
	if info.PrimaryKey == nil {
		if idField, ok := info.FieldByName["ID"]; ok {
			idField.PrimaryKey = true
			info.PrimaryKey = idField
		}
	}

	return info, nil
}

func parseStructFields(t reflect.Type, info *ModelInfo) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}

		// Handle embedded structs
		if field.Anonymous {
			fieldType := field.Type
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() == reflect.Struct {
				if err := parseStructFields(fieldType, info); err != nil {
					return err
				}
			}
			continue
		}

		// Parse gd tag
		gdTag := field.Tag.Get("gd")
		if gdTag == "-" {
			continue // explicitly ignored
		}

		parsedField, err := parseFieldTag(field, gdTag)
		if err != nil {
			return fmt.Errorf("orm: error parsing field %s on %s: %v", field.Name, info.Name, err)
		}

		if parsedField != nil {
			info.Fields = append(info.Fields, parsedField)
			info.FieldByName[parsedField.Name] = parsedField
		}
	}
	return nil
}

func parseFieldTag(sf reflect.StructField, tag string) (*Field, error) {
	// If no tag is provided but it's exported, we don't automatically map it
	// unless we want to enforce standard Django defaults. We'll skip empty tags.
	if tag == "" {
		return nil, nil
	}

	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return nil, nil
	}

	fieldTypeStr := strings.TrimSpace(parts[0])

	f := &Field{
		Name:    sf.Name,
		Column:  toSnakeCase(sf.Name),
		Type:    FieldType(fieldTypeStr),
		GoType:  sf.Type,
		Options: FieldOptions{},
	}

	// Parse kwargs
	for _, part := range parts[1:] {
		kv := strings.SplitN(part, "=", 2)
		key := strings.TrimSpace(kv[0])
		var val string
		if len(kv) == 2 {
			val = strings.TrimSpace(kv[1])
		}

		switch key {
		case "max_length":
			f.Options.MaxLength, _ = strconv.Atoi(val)
		case "max_digits":
			f.Options.MaxDigits, _ = strconv.Atoi(val)
		case "decimal_places":
			f.Options.DecimalPlaces, _ = strconv.Atoi(val)
		case "blank":
			f.Options.Blank = val == "true"
		case "null":
			f.Options.Null = val == "true"
		case "default":
			f.Options.Default = val // Stored as string, casted during execution
		case "unique":
			f.Options.Unique = val == "true"
		case "db_index":
			f.Options.DbIndex = val == "true"
		case "auto_now":
			f.Options.AutoNow = val == "true"
		case "auto_now_add":
			f.Options.AutoNowAdd = val == "true"
		case "upload_to":
			f.Options.UploadTo = val
		case "to":
			f.Options.To = val
		case "on_delete":
			f.Options.OnDelete = val
		case "related_name":
			f.Options.RelatedName = val
		case "db_column":
			f.Options.DbColumn = val
			f.Column = val
		case "through":
			f.Options.Through = val
		case "primary_key":
			f.PrimaryKey = val == "true"
		case "auto_created":
			f.Options.AutoCreated = val == "true"
		case "srid":
			f.Options.SRID, _ = strconv.Atoi(val)
		case "size":
			f.Options.Size, _ = strconv.Atoi(val)
		case "base_field":
			f.Options.BaseField = val
		}
	}

	return f, nil
}

// toSnakeCase converts camel case to snake case
func toSnakeCase(s string) string {
	var buf strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && unicode.IsLower(rune(s[i-1])) {
				buf.WriteRune('_')
			}
			buf.WriteRune(unicode.ToLower(r))
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// findParentModelType returns the reflect.Type of the first embedded struct
// that is not the base Model. Used to identify the parent of a proxy model.
func findParentModelType(t reflect.Type) reflect.Type {
	modelType := reflect.TypeOf(Model{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.Anonymous {
			continue
		}
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() != reflect.Struct {
			continue
		}
		if fieldType == modelType {
			continue
		}
		return fieldType
	}
	return nil
}

// resolveModelDbTable resolves the DbTable for a given model type by checking
// the registry first, then falling back to parsing the model on-the-fly.
func resolveModelDbTable(t reflect.Type) (string, error) {
	if info, ok := globalRegistry.Load(t); ok {
		return info.(*ModelInfo).Meta.DbTable, nil
	}
	// Parent not registered yet — parse it on-the-fly to determine its table.
	instance := reflect.New(t).Interface()
	parentInfo, err := parseModel(instance)
	if err != nil {
		return "", fmt.Errorf("failed to resolve table for parent model %s: %w", t.Name(), err)
	}
	return parentInfo.Meta.DbTable, nil
}
