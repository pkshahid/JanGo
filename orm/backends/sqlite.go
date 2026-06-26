package backends

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/orm"

	// Import the modernc.org/sqlite pure Go driver
	_ "modernc.org/sqlite"
)

func init() {
	Register("sqlite", func() Backend {
		return &SQLiteBackend{}
	})
	Register("sqlite3", func() Backend {
		return &SQLiteBackend{}
	})
}

// SQLiteBackend implements the Backend interface for SQLite.
type SQLiteBackend struct {
	db *sql.DB
}

func (b *SQLiteBackend) Connect(config settings.DatabaseConfig) error {
	dsn := config.DSN
	if dsn == "" {
		dsn = config.Name // SQLite typically just needs the filename
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	b.db = db
	return b.db.Ping()
}

func (b *SQLiteBackend) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

func (b *SQLiteBackend) DB() *sql.DB {
	return b.db
}

func (b *SQLiteBackend) Execute(ctx context.Context, sqlStr string, args ...any) (sql.Result, error) {
	return b.db.ExecContext(ctx, sqlStr, args...)
}

func (b *SQLiteBackend) Query(ctx context.Context, sqlStr string, args ...any) (*sql.Rows, error) {
	return b.db.QueryContext(ctx, sqlStr, args...)
}

func (b *SQLiteBackend) QueryRow(ctx context.Context, sqlStr string, args ...any) *sql.Row {
	return b.db.QueryRowContext(ctx, sqlStr, args...)
}

func (b *SQLiteBackend) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return b.db.BeginTx(ctx, opts)
}

func (b *SQLiteBackend) DatabaseName() string {
	return "sqlite"
}

func (b *SQLiteBackend) Features() DatabaseFeatures {
	return DatabaseFeatures{
		SupportsReturning:  true,
		SupportsJSON:       true,
		SupportsSavepoints: true,
	}
}

func (b *SQLiteBackend) SchemaEditor() SchemaEditor {
	return &SQLiteSchemaEditor{backend: b}
}

// SQLiteSchemaEditor handles schema operations for SQLite.
type SQLiteSchemaEditor struct {
	backend *SQLiteBackend
}

func (s *SQLiteSchemaEditor) CreateTable(model *orm.ModelInfo) error {
	var cols []string
	for _, f := range model.Fields {
		// ManyToMany fields are resolved through junction tables, not columns.
		if f.Type == orm.ManyToManyField {
			continue
		}
		sqlType := s.typeMapping(f.Type, f.Options)
		if f.PrimaryKey {
			sqlType += " PRIMARY KEY AUTOINCREMENT"
		} else {
			if !f.Options.Null && !f.Options.Blank {
				sqlType += " NOT NULL"
			}
			if f.Options.Unique {
				sqlType += " UNIQUE"
			}
		}
		cols = append(cols, fmt.Sprintf("%s %s", f.Column, sqlType))
	}

	// Add inline constraints (CHECK and non-conditional UNIQUE)
	for _, c := range model.Meta.Constraints {
		switch ct := c.(type) {
		case orm.CheckConstraint:
			cols = append(cols, fmt.Sprintf("CONSTRAINT %s CHECK (%s)", ct.Name, ct.Check))
		case orm.UniqueConstraint:
			if ct.Condition == "" {
				cols = append(cols, fmt.Sprintf("CONSTRAINT %s UNIQUE (%s)", ct.Name, strings.Join(ct.Fields, ", ")))
			}
		}
	}

	// Create table
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", model.Meta.DbTable, strings.Join(cols, ", "))
	_, err := s.backend.Execute(context.Background(), query)
	if err != nil {
		return err
	}

	// Create explicit indexes
	for _, idx := range model.Meta.Indexes {
		if err := s.CreateIndex(model, idx); err != nil {
			return err
		}
	}

	// Create partial unique indexes for conditional UniqueConstraints
	for _, c := range model.Meta.Constraints {
		if uc, ok := c.(orm.UniqueConstraint); ok && uc.Condition != "" {
			if err := s.AddConstraint(model, uc); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SQLiteSchemaEditor) typeMapping(t orm.FieldType, opts orm.FieldOptions) string {
	switch t {
	case orm.CharField, orm.SlugField, orm.URLField, orm.EmailField:
		return fmt.Sprintf("VARCHAR(%d)", opts.MaxLength)
	case orm.TextField:
		return "TEXT"
	case orm.IntegerField, orm.ForeignKey, orm.OneToOneField:
		return "INTEGER"
	case orm.BigAutoField:
		return "INTEGER" // SQLite AUTOINCREMENT must be INTEGER
	case orm.BooleanField, orm.NullBooleanField:
		return "BOOLEAN"
	case orm.FloatField, orm.DecimalField:
		return "REAL"
	case orm.DateTimeField, orm.DateField, orm.TimeField:
		return "DATETIME"
	case orm.JSONField:
		return "JSON" // Supported by SQLite JSON1 extension natively in modernc
	}
	return "TEXT" // Fallback
}

func (s *SQLiteSchemaEditor) DeleteTable(model *orm.ModelInfo) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s;", model.Meta.DbTable)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *SQLiteSchemaEditor) AddColumn(model *orm.ModelInfo, field *orm.Field) error {
	sqlType := s.typeMapping(field.Type, field.Options)
	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", model.Meta.DbTable, field.Column, sqlType)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *SQLiteSchemaEditor) RemoveColumn(model *orm.ModelInfo, fieldName string) error {
	// SQLite ALTER TABLE DROP COLUMN is supported in newer versions (3.35.0+)
	// Modernc supports this.
	query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", model.Meta.DbTable, fieldName)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *SQLiteSchemaEditor) AlterColumn(model *orm.ModelInfo, oldField, newField *orm.Field) error {
	return fmt.Errorf("AlterColumn not directly supported in basic SQLite without table recreation")
}

func (s *SQLiteSchemaEditor) CreateIndex(model *orm.ModelInfo, index orm.Index) error {
	unique := ""
	if index.Unique {
		unique = "UNIQUE "
	}
	name := index.Name
	if name == "" {
		name = fmt.Sprintf("idx_%s_%s", model.Meta.DbTable, strings.Join(index.Fields, "_"))
	}
	query := fmt.Sprintf("CREATE %sINDEX IF NOT EXISTS %s ON %s (%s);", unique, name, model.Meta.DbTable, strings.Join(index.Fields, ", "))
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *SQLiteSchemaEditor) DeleteIndex(model *orm.ModelInfo, indexName string) error {
	query := fmt.Sprintf("DROP INDEX IF EXISTS %s;", indexName)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *SQLiteSchemaEditor) AddForeignKey(model *orm.ModelInfo, field *orm.Field) error {
	return fmt.Errorf("AddForeignKey unsupported in SQLite via ALTER TABLE")
}

func (s *SQLiteSchemaEditor) RemoveForeignKey(model *orm.ModelInfo, fieldName string) error {
	return fmt.Errorf("RemoveForeignKey unsupported in SQLite via ALTER TABLE")
}

func (s *SQLiteSchemaEditor) AddConstraint(model *orm.ModelInfo, constraint orm.Constraint) error {
	switch ct := constraint.(type) {
	case orm.CheckConstraint:
		query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s CHECK (%s)",
			model.Meta.DbTable, ct.Name, ct.Check)
		_, err := s.backend.Execute(context.Background(), query)
		return err
	case orm.UniqueConstraint:
		if ct.Condition != "" {
			query := fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s (%s) WHERE %s",
				ct.Name, model.Meta.DbTable, strings.Join(ct.Fields, ", "), ct.Condition)
			_, err := s.backend.Execute(context.Background(), query)
			return err
		}
		query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)",
			model.Meta.DbTable, ct.Name, strings.Join(ct.Fields, ", "))
		_, err := s.backend.Execute(context.Background(), query)
		return err
	default:
		return fmt.Errorf("unsupported constraint type: %T", constraint)
	}
}

func (s *SQLiteSchemaEditor) RemoveConstraint(model *orm.ModelInfo, constraintName string) error {
	// Try dropping as a table constraint first; if that fails, try dropping as an index.
	query := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", model.Meta.DbTable, constraintName)
	_, err := s.backend.Execute(context.Background(), query)
	if err == nil {
		return nil
	}
	// Fallback: try dropping as an index (for partial unique constraints)
	indexQuery := fmt.Sprintf("DROP INDEX IF EXISTS %s;", constraintName)
	_, err = s.backend.Execute(context.Background(), indexQuery)
	return err
}
