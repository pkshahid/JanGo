package backends

import (
	"context"
	"database/sql"
	"fmt"

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
	// Not fully implemented for prototype, returns empty interface stubs.
	// Postgres DDL differs slightly from SQLite (e.g. SERIAL, JSONB).
	return &MockSchemaEditor{}
}

// MockSchemaEditor is used to fulfill interfaces for incomplete backend drivers
type MockSchemaEditor struct{}

func (m *MockSchemaEditor) CreateTable(model *orm.ModelInfo) error                    { return nil }
func (m *MockSchemaEditor) DeleteTable(model *orm.ModelInfo) error                    { return nil }
func (m *MockSchemaEditor) AddColumn(model *orm.ModelInfo, field *orm.Field) error    { return nil }
func (m *MockSchemaEditor) RemoveColumn(model *orm.ModelInfo, fieldName string) error { return nil }
func (m *MockSchemaEditor) AlterColumn(model *orm.ModelInfo, oldField, newField *orm.Field) error {
	return nil
}
func (m *MockSchemaEditor) CreateIndex(model *orm.ModelInfo, index orm.Index) error       { return nil }
func (m *MockSchemaEditor) DeleteIndex(model *orm.ModelInfo, indexName string) error      { return nil }
func (m *MockSchemaEditor) AddForeignKey(model *orm.ModelInfo, field *orm.Field) error    { return nil }
func (m *MockSchemaEditor) RemoveForeignKey(model *orm.ModelInfo, fieldName string) error { return nil }
