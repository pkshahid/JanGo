// Package commands provides Django-style management commands.
// inspectdb introspects an existing database and generates Go model code.
package commands

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkshahid/JanGo/orm/backends"
)

// InspectDBConfig holds configuration for the inspectdb command.
type InspectDBConfig struct {
	Database     string
	TableNames   []string // empty = all tables
	IncludeViews bool
}

// TableInfo represents metadata about a database table.
type TableInfo struct {
	Name    string
	Columns []ColumnInfo
}

// ColumnInfo represents metadata about a database column.
type ColumnInfo struct {
	Name       string
	DataType   string
	Nullable   bool
	PrimaryKey bool
	MaxLength  int
	Default    string
}

// InspectDB introspects the database and returns generated Go source code for models.
func InspectDB(config InspectDBConfig) (string, error) {
	if config.Database == "" {
		config.Database = "default"
	}

	backend, err := backends.GetBackend(config.Database)
	if err != nil {
		return "", fmt.Errorf("inspectdb: %v", err)
	}

	db := backend.DB()
	if db == nil {
		return "", fmt.Errorf("inspectdb: no database connection")
	}

	tables, err := getTablesFromDB(db, config)
	if err != nil {
		return "", fmt.Errorf("inspectdb: %v", err)
	}

	return generateModels(tables), nil
}

func getTablesFromDB(db *sql.DB, config InspectDBConfig) ([]TableInfo, error) {
	ctx := context.Background()

	// Query SQLite master table for table names
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %v", err)
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		if len(config.TableNames) > 0 {
			found := false
			for _, t := range config.TableNames {
				if t == name {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		tableNames = append(tableNames, name)
	}

	var tables []TableInfo
	for _, tableName := range tableNames {
		columns, err := getColumnsForTable(db, ctx, tableName)
		if err != nil {
			return nil, err
		}
		tables = append(tables, TableInfo{Name: tableName, Columns: columns})
	}

	return tables, nil
}

func getColumnsForTable(db *sql.DB, ctx context.Context, tableName string) ([]ColumnInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var dfltValue *string
		var pk int
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		col := ColumnInfo{
			Name:       name,
			DataType:   dataType,
			Nullable:   notNull == 0,
			PrimaryKey: pk > 0,
		}
		if dfltValue != nil {
			col.Default = *dfltValue
		}
		columns = append(columns, col)
	}
	return columns, nil
}

// InspectDBFromTableInfo generates model code from pre-collected table information.
// This is useful for testing and when table info is available from other sources.
func InspectDBFromTableInfo(tables []TableInfo) string {
	return generateModels(tables)
}

func generateModels(tables []TableInfo) string {
	var sb strings.Builder

	sb.WriteString("package models\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"github.com/pkshahid/JanGo/orm\"\n")
	sb.WriteString(")\n\n")

	for _, table := range tables {
		modelName := toGoName(table.Name)
		sb.WriteString(fmt.Sprintf("// %s is auto-generated from table %q.\n", modelName, table.Name))
		sb.WriteString(fmt.Sprintf("type %s struct {\n", modelName))
		sb.WriteString("\torm.Model\n")

		for _, col := range table.Columns {
			if isBaseModelField(col.Name) {
				continue
			}
			fieldName := toGoName(col.Name)
			goType := sqlTypeToGoType(col.DataType, col.Nullable)
			tag := generateFieldTag(col)
			sb.WriteString(fmt.Sprintf("\t%s %s `gd:\"%s\"`\n", fieldName, goType, tag))
		}

		sb.WriteString("}\n\n")

		// Generate Meta method
		sb.WriteString(fmt.Sprintf("func (*%s) ModelMeta() *orm.Meta {\n", modelName))
		sb.WriteString(fmt.Sprintf("\treturn &orm.Meta{DbTable: %q}\n", table.Name))
		sb.WriteString("}\n\n")
	}

	return sb.String()
}

func toGoName(name string) string {
	parts := strings.Split(name, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}
	return result.String()
}

func sqlTypeToGoType(sqlType string, nullable bool) string {
	sqlType = strings.ToUpper(sqlType)

	var goType string
	switch {
	case strings.Contains(sqlType, "INT"):
		goType = "int"
	case strings.Contains(sqlType, "FLOAT") || strings.Contains(sqlType, "REAL") || strings.Contains(sqlType, "DOUBLE"):
		goType = "float64"
	case strings.Contains(sqlType, "DECIMAL") || strings.Contains(sqlType, "NUMERIC"):
		goType = "float64"
	case strings.Contains(sqlType, "BOOL"):
		goType = "bool"
	case strings.Contains(sqlType, "TEXT") || strings.Contains(sqlType, "CHAR") || strings.Contains(sqlType, "VARCHAR"):
		goType = "string"
	case strings.Contains(sqlType, "BLOB") || strings.Contains(sqlType, "BINARY"):
		goType = "[]byte"
	case strings.Contains(sqlType, "DATE") || strings.Contains(sqlType, "TIME"):
		goType = "time.Time"
	case strings.Contains(sqlType, "JSON"):
		goType = "json.RawMessage"
	default:
		goType = "string"
	}

	if nullable && goType != "string" && goType != "[]byte" {
		return "*" + goType
	}
	return goType
}

func generateFieldTag(col ColumnInfo) string {
	var parts []string

	fieldType := sqlTypeToFieldType(col.DataType)
	parts = append(parts, fieldType)

	if col.MaxLength > 0 {
		parts = append(parts, fmt.Sprintf("max_length=%d", col.MaxLength))
	}
	if col.Nullable {
		parts = append(parts, "null=true")
	}
	if col.PrimaryKey {
		parts = append(parts, "primary_key=true")
	}

	return strings.Join(parts, ",")
}

func sqlTypeToFieldType(sqlType string) string {
	sqlType = strings.ToUpper(sqlType)
	switch {
	case strings.Contains(sqlType, "INT"):
		return "IntegerField"
	case strings.Contains(sqlType, "FLOAT") || strings.Contains(sqlType, "REAL") || strings.Contains(sqlType, "DOUBLE"):
		return "FloatField"
	case strings.Contains(sqlType, "DECIMAL") || strings.Contains(sqlType, "NUMERIC"):
		return "DecimalField"
	case strings.Contains(sqlType, "BOOL"):
		return "BooleanField"
	case strings.Contains(sqlType, "TEXT"):
		return "TextField"
	case strings.Contains(sqlType, "CHAR") || strings.Contains(sqlType, "VARCHAR"):
		return "CharField"
	case strings.Contains(sqlType, "DATETIME") || strings.Contains(sqlType, "TIMESTAMP"):
		return "DateTimeField"
	case strings.Contains(sqlType, "DATE"):
		return "DateField"
	case strings.Contains(sqlType, "TIME"):
		return "TimeField"
	case strings.Contains(sqlType, "JSON"):
		return "JSONField"
	case strings.Contains(sqlType, "BLOB") || strings.Contains(sqlType, "BINARY"):
		return "BinaryField"
	default:
		return "CharField"
	}
}

func isBaseModelField(name string) bool {
	switch strings.ToLower(name) {
	case "id", "created_at", "updated_at", "deleted_at":
		return true
	}
	return false
}
