package backends

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/orm"

	// Import the pure-Go Oracle driver (no Oracle Instant Client required).
	_ "github.com/sijms/go-ora/v2"
)

func init() {
	Register("oracle", func() Backend {
		return &OracleBackend{}
	})
	Register("oracledb", func() Backend {
		return &OracleBackend{}
	})
}

// OracleBackend implements the Backend interface for Oracle Database.
type OracleBackend struct {
	db *sql.DB
}

func (b *OracleBackend) Connect(config settings.DatabaseConfig) error {
	dsn := config.DSN
	if dsn == "" {
		port := config.Port
		if port == 0 {
			port = 1521 // default Oracle port
		}
		dsn = fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
			config.User, config.Password, config.Host, port, config.Name)
	}
	db, err := sql.Open("oracle", dsn)
	if err != nil {
		return err
	}
	b.db = db
	return b.db.Ping()
}

func (b *OracleBackend) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

func (b *OracleBackend) DB() *sql.DB {
	return b.db
}

func (b *OracleBackend) Execute(ctx context.Context, sqlStr string, args ...any) (sql.Result, error) {
	return b.db.ExecContext(ctx, sqlStr, args...)
}

func (b *OracleBackend) Query(ctx context.Context, sqlStr string, args ...any) (*sql.Rows, error) {
	return b.db.QueryContext(ctx, sqlStr, args...)
}

func (b *OracleBackend) QueryRow(ctx context.Context, sqlStr string, args ...any) *sql.Row {
	return b.db.QueryRowContext(ctx, sqlStr, args...)
}

func (b *OracleBackend) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return b.db.BeginTx(ctx, opts)
}

func (b *OracleBackend) DatabaseName() string {
	return "oracle"
}

func (b *OracleBackend) Features() DatabaseFeatures {
	return DatabaseFeatures{
		SupportsReturning:  false, // Oracle does not support RETURNING clause in the same way as PostgreSQL
		SupportsJSON:       true,  // Oracle 12c+ supports JSON natively
		SupportsSavepoints: true,
	}
}

func (b *OracleBackend) SchemaEditor() SchemaEditor {
	return &OracleSchemaEditor{backend: b}
}

// OracleSchemaEditor handles schema operations for Oracle Database.
type OracleSchemaEditor struct {
	backend *OracleBackend
}

func (s *OracleSchemaEditor) CreateTable(model *orm.ModelInfo) error {
	var cols []string
	for _, f := range model.Fields {
		sqlType := s.typeMapping(f.Type, f.Options)
		if f.PrimaryKey {
			sqlType += " PRIMARY KEY"
			if f.Type == orm.BigAutoField {
				sqlType += " GENERATED ALWAYS AS IDENTITY"
			}
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

	query := fmt.Sprintf("CREATE TABLE %s (%s)", model.Meta.DbTable, strings.Join(cols, ", "))
	_, err := s.backend.Execute(context.Background(), query)
	if err != nil {
		return err
	}

	for _, idx := range model.Meta.Indexes {
		if err := s.CreateIndex(model, idx); err != nil {
			return err
		}
	}
	return nil
}

func (s *OracleSchemaEditor) typeMapping(t orm.FieldType, opts orm.FieldOptions) string {
	switch t {
	case orm.CharField, orm.SlugField, orm.URLField, orm.EmailField:
		maxLen := opts.MaxLength
		if maxLen <= 0 {
			maxLen = 255
		}
		return fmt.Sprintf("VARCHAR2(%d)", maxLen)
	case orm.TextField:
		return "CLOB"
	case orm.IntegerField, orm.ForeignKey, orm.OneToOneField:
		return "NUMBER(10)"
	case orm.SmallIntegerField:
		return "NUMBER(5)"
	case orm.BigIntegerField:
		return "NUMBER(19)"
	case orm.BigAutoField:
		return "NUMBER(19)"
	case orm.BooleanField, orm.NullBooleanField:
		return "NUMBER(1)"
	case orm.FloatField:
		return "BINARY_FLOAT"
	case orm.DecimalField:
		digits := opts.MaxDigits
		if digits <= 0 {
			digits = 10
		}
		places := opts.DecimalPlaces
		return fmt.Sprintf("NUMBER(%d,%d)", digits, places)
	case orm.DateTimeField:
		return "TIMESTAMP"
	case orm.DateField:
		return "DATE"
	case orm.TimeField:
		return "TIMESTAMP"
	case orm.DurationField:
		return "INTERVAL DAY TO SECOND"
	case orm.UUIDField:
		return "VARCHAR2(36)"
	case orm.JSONField:
		return "CLOB" // Oracle 12c+ supports JSON as CLOB with IS JSON constraint
	case orm.BinaryField:
		return "BLOB"
	case orm.IPAddressField:
		return "VARCHAR2(45)"
	default:
		return "CLOB"
	}
}

func (s *OracleSchemaEditor) DeleteTable(model *orm.ModelInfo) error {
	query := fmt.Sprintf("DROP TABLE %s CASCADE CONSTRAINTS", model.Meta.DbTable)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *OracleSchemaEditor) AddColumn(model *orm.ModelInfo, field *orm.Field) error {
	sqlType := s.typeMapping(field.Type, field.Options)
	query := fmt.Sprintf("ALTER TABLE %s ADD (%s %s)", model.Meta.DbTable, field.Column, sqlType)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *OracleSchemaEditor) RemoveColumn(model *orm.ModelInfo, fieldName string) error {
	query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", model.Meta.DbTable, fieldName)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *OracleSchemaEditor) AlterColumn(model *orm.ModelInfo, oldField, newField *orm.Field) error {
	sqlType := s.typeMapping(newField.Type, newField.Options)
	query := fmt.Sprintf("ALTER TABLE %s MODIFY (%s %s)", model.Meta.DbTable, newField.Column, sqlType)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *OracleSchemaEditor) CreateIndex(model *orm.ModelInfo, index orm.Index) error {
	unique := ""
	if index.Unique {
		unique = "UNIQUE "
	}
	name := index.Name
	if name == "" {
		name = fmt.Sprintf("idx_%s_%s", model.Meta.DbTable, strings.Join(index.Fields, "_"))
	}
	// Oracle index names have a 30-byte limit (128 in 12.2R2+); truncate safely
	if len(name) > 30 {
		name = name[:30]
	}
	query := fmt.Sprintf("CREATE %sINDEX %s ON %s (%s)", unique, name, model.Meta.DbTable, strings.Join(index.Fields, ", "))
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *OracleSchemaEditor) DeleteIndex(model *orm.ModelInfo, indexName string) error {
	query := fmt.Sprintf("DROP INDEX %s", indexName)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *OracleSchemaEditor) AddForeignKey(model *orm.ModelInfo, field *orm.Field) error {
	if field.Options.To == "" {
		return fmt.Errorf("AddForeignKey: missing target table for field %s", field.Name)
	}
	fkName := fmt.Sprintf("fk_%s_%s", model.Meta.DbTable, field.Column)
	if len(fkName) > 30 {
		fkName = fkName[:30]
	}
	query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(id)",
		model.Meta.DbTable, fkName, field.Column, field.Options.To)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}

func (s *OracleSchemaEditor) RemoveForeignKey(model *orm.ModelInfo, fieldName string) error {
	fkName := fmt.Sprintf("fk_%s_%s", model.Meta.DbTable, fieldName)
	if len(fkName) > 30 {
		fkName = fkName[:30]
	}
	query := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", model.Meta.DbTable, fkName)
	_, err := s.backend.Execute(context.Background(), query)
	return err
}
