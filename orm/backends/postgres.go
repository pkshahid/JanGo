package backends

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/orm"

	// Import PostgreSQL driver
	_ "github.com/lib/pq"
)

func init() {
	Register("postgres", func() Backend {
		return &PostgresBackend{}
	})
}

type PostgresBackend struct {
	db *sql.DB
}

func (b *PostgresBackend) Connect(config settings.DatabaseConfig) error {
	dsn := config.DSN
	if dsn == "" {
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.User, config.Password, config.Name)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	b.db = db
	return b.db.Ping()
}

func (b *PostgresBackend) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

func (b *PostgresBackend) DB() *sql.DB {
	return b.db
}

func (b *PostgresBackend) Execute(ctx context.Context, sqlStr string, args ...any) (sql.Result, error) {
	return b.db.ExecContext(ctx, sqlStr, args...)
}

func (b *PostgresBackend) Query(ctx context.Context, sqlStr string, args ...any) (*sql.Rows, error) {
	return b.db.QueryContext(ctx, sqlStr, args...)
}

func (b *PostgresBackend) QueryRow(ctx context.Context, sqlStr string, args ...any) *sql.Row {
	return b.db.QueryRowContext(ctx, sqlStr, args...)
}

func (b *PostgresBackend) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return b.db.BeginTx(ctx, opts)
}

func (b *PostgresBackend) DatabaseName() string {
	return "postgres"
}

func (b *PostgresBackend) Features() DatabaseFeatures {
	return DatabaseFeatures{
		SupportsReturning:  true,
		SupportsJSON:       true,
		SupportsSavepoints: true,
	}
}

func (b *PostgresBackend) SchemaEditor() SchemaEditor {
	return &PostGISSchemaEditor{backend: b}
}

// PostGISSchemaEditor handles schema operations for PostgreSQL/PostGIS.
type PostGISSchemaEditor struct {
	backend *PostgresBackend
}

func (s *PostGISSchemaEditor) typeMapping(t orm.FieldType, opts orm.FieldOptions) string {
	switch t {
	case orm.CharField, orm.SlugField, orm.URLField, orm.EmailField:
		maxLen := opts.MaxLength
		if maxLen == 0 {
			maxLen = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", maxLen)
	case orm.TextField:
		return "TEXT"
	case orm.IntegerField, orm.ForeignKey, orm.OneToOneField:
		return "INTEGER"
	case orm.SmallIntegerField:
		return "SMALLINT"
	case orm.BigIntegerField:
		return "BIGINT"
	case orm.BigAutoField:
		return "SERIAL"
	case orm.BooleanField, orm.NullBooleanField:
		return "BOOLEAN"
	case orm.FloatField:
		return "DOUBLE PRECISION"
	case orm.DecimalField:
		digits := opts.MaxDigits
		places := opts.DecimalPlaces
		if digits == 0 {
			digits = 10
		}
		return fmt.Sprintf("NUMERIC(%d, %d)", digits, places)
	case orm.DateTimeField:
		return "TIMESTAMP WITH TIME ZONE"
	case orm.DateField:
		return "DATE"
	case orm.TimeField:
		return "TIME"
	case orm.DurationField:
		return "INTERVAL"
	case orm.JSONField:
		return "JSONB"
	case orm.UUIDField:
		return "UUID"
	case orm.IPAddressField:
		return "INET"
	case orm.BinaryField:
		return "BYTEA"
	// ── GIS geometry types (PostGIS) ──
	case orm.PointField:
		return s.geometryType("POINT", opts)
	case orm.LineStringField:
		return s.geometryType("LINESTRING", opts)
	case orm.PolygonField:
		return s.geometryType("POLYGON", opts)
	case orm.MultiPointField:
		return s.geometryType("MULTIPOINT", opts)
	// ── PostgreSQL-specific fields (contrib.postgres) ──
	case orm.ArrayField:
		baseType := s.arrayBaseType(opts.BaseField, opts)
		return baseType + "[]"
	case orm.HStoreField:
		return "HSTORE"
	case orm.IntegerRangeField:
		return "INT4RANGE"
	case orm.BigIntegerRangeField:
		return "INT8RANGE"
	case orm.DecimalRangeField:
		return "NUMRANGE"
	case orm.DateTimeRangeField:
		return "TSTZRANGE"
	case orm.DateRangeField:
		return "DATERANGE"
	case orm.CICharField:
		maxLen := opts.MaxLength
		if maxLen == 0 {
			maxLen = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", maxLen) // stored as CITEXT via citext extension
	case orm.CITextField:
		return "TEXT"
	case orm.CIEmailField:
		return "VARCHAR(254)"
	}
	return "TEXT"
}

// geometryType returns the PostGIS column type definition. When opts.SRID is
// non-zero the column is created with an explicit SRID constraint. When
// opts.SRID is zero it defaults to 4326 (WGS84).
func (s *PostGISSchemaEditor) geometryType(geomType string, opts orm.FieldOptions) string {
	srid := opts.SRID
	if srid == 0 {
		srid = 4326
	}
	return fmt.Sprintf("geometry(%s,%d)", geomType, srid)
}

// arrayBaseType maps the base_field option string to a PostgreSQL column type
// for use inside an ARRAY column definition.
func (s *PostGISSchemaEditor) arrayBaseType(baseField string, opts orm.FieldOptions) string {
	switch orm.FieldType(baseField) {
	case orm.IntegerField:
		return "INTEGER"
	case orm.SmallIntegerField:
		return "SMALLINT"
	case orm.BigIntegerField:
		return "BIGINT"
	case orm.FloatField:
		return "DOUBLE PRECISION"
	case orm.DecimalField:
		digits := opts.MaxDigits
		places := opts.DecimalPlaces
		if digits == 0 {
			digits = 10
		}
		return fmt.Sprintf("NUMERIC(%d, %d)", digits, places)
	case orm.BooleanField:
		return "BOOLEAN"
	case orm.CharField, orm.SlugField, orm.URLField, orm.EmailField:
		maxLen := opts.MaxLength
		if maxLen == 0 {
			maxLen = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", maxLen)
	case orm.TextField:
		return "TEXT"
	case orm.UUIDField:
		return "UUID"
	case orm.JSONField:
		return "JSONB"
	default:
		return "TEXT"
	}
}

func (s *PostGISSchemaEditor) CreateTable(model *orm.ModelInfo) error {
	var cols []string
	for _, f := range model.Fields {
		if f.Type == orm.ManyToManyField {
			continue
		}
		sqlType := s.typeMapping(f.Type, f.Options)
		if f.PrimaryKey {
			if f.Type == orm.BigAutoField {
				sqlType = "BIGSERIAL"
			}
			sqlType += " PRIMARY KEY"
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

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", model.Meta.DbTable, strings.Join(cols, ", "))
	if _, err := s.backend.Execute(context.Background(), query); err != nil {
		return err
	}

	for _, idx := range model.Meta.Indexes {
		if err := s.CreateIndex(model, idx); err != nil {
			return err
		}
	}

	for _, c := range model.Meta.Constraints {
		if uc, ok := c.(orm.UniqueConstraint); ok && uc.Condition != "" {
			if err := s.AddConstraint(model, uc); err != nil {
				return err
			}
		}
	}

	for _, si := range model.Meta.SpatialIndexes {
		if err := s.CreateSpatialIndex(model, si); err != nil {
			return err
		}
	}

	return nil
}

func (s *PostGISSchemaEditor) DeleteTable(model *orm.ModelInfo) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", model.Meta.DbTable)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *PostGISSchemaEditor) AddColumn(model *orm.ModelInfo, field *orm.Field) error {
	sqlType := s.typeMapping(field.Type, field.Options)
	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", model.Meta.DbTable, field.Column, sqlType)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *PostGISSchemaEditor) RemoveColumn(model *orm.ModelInfo, fieldName string) error {
	query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", model.Meta.DbTable, fieldName)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *PostGISSchemaEditor) AlterColumn(model *orm.ModelInfo, oldField, newField *orm.Field) error {
	newType := s.typeMapping(newField.Type, newField.Options)
	query := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s;", model.Meta.DbTable, newField.Column, newType)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *PostGISSchemaEditor) CreateIndex(model *orm.ModelInfo, index orm.Index) error {
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

// CreateSpatialIndex creates a GiST spatial index on a geometry column.
func (s *PostGISSchemaEditor) CreateSpatialIndex(model *orm.ModelInfo, index orm.SpatialIndex) error {
	name := index.Name
	if name == "" {
		name = fmt.Sprintf("sidx_%s_%s", model.Meta.DbTable, strings.Join(index.Fields, "_"))
	}
	fields := strings.Join(index.Fields, ", ")
	query := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s USING GIST (%s);", name, model.Meta.DbTable, fields)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *PostGISSchemaEditor) DeleteIndex(model *orm.ModelInfo, indexName string) error {
	query := fmt.Sprintf("DROP INDEX IF EXISTS %s;", indexName)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *PostGISSchemaEditor) AddForeignKey(model *orm.ModelInfo, field *orm.Field) error {
	if field.Options.To == "" {
		return fmt.Errorf("AddForeignKey: missing 'to' reference for field %s", field.Name)
	}
	query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT fk_%s_%s FOREIGN KEY (%s) REFERENCES %s(id) ON DELETE %s;",
		model.Meta.DbTable, model.Meta.DbTable, field.Column, field.Column,
		field.Options.To, field.Options.OnDelete)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *PostGISSchemaEditor) RemoveForeignKey(model *orm.ModelInfo, fieldName string) error {
	constraintName := fmt.Sprintf("fk_%s_%s", model.Meta.DbTable, fieldName)
	query := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s;", model.Meta.DbTable, constraintName)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *PostGISSchemaEditor) AddConstraint(model *orm.ModelInfo, constraint orm.Constraint) error {
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

func (s *PostGISSchemaEditor) RemoveConstraint(model *orm.ModelInfo, constraintName string) error {
	query := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s;", model.Meta.DbTable, constraintName)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}
