package backends

import (
	"context"
	"database/sql"

	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/orm"
)

// DatabaseFeatures describes what a database backend supports.
type DatabaseFeatures struct {
	SupportsReturning bool
	SupportsJSON      bool
	SupportsSavepoints bool
}

// SchemaEditor defines the interface for executing DDL operations.
type SchemaEditor interface {
	CreateTable(model *orm.ModelInfo) error
	DeleteTable(model *orm.ModelInfo) error
	AddColumn(model *orm.ModelInfo, field *orm.Field) error
	RemoveColumn(model *orm.ModelInfo, fieldName string) error
	AlterColumn(model *orm.ModelInfo, oldField, newField *orm.Field) error
	CreateIndex(model *orm.ModelInfo, index orm.Index) error
	DeleteIndex(model *orm.ModelInfo, indexName string) error
	AddForeignKey(model *orm.ModelInfo, field *orm.Field) error
	RemoveForeignKey(model *orm.ModelInfo, fieldName string) error
}

// DBExecutor interface allows executing queries against a *sql.DB or *sql.Tx
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Backend interface for database adapters.
type Backend interface {
	Connect(config settings.DatabaseConfig) error
	Close() error

	// Query Execution
	Execute(ctx context.Context, sqlStr string, args ...any) (sql.Result, error)
	Query(ctx context.Context, sqlStr string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, sqlStr string, args ...any) *sql.Row

	// Transactions
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	// Metadata
	DatabaseName() string
	SchemaEditor() SchemaEditor
	Features() DatabaseFeatures

	// Raw DB access (internal)
	DB() *sql.DB
}

// DatabaseRouter determines which database to use for read/write.
type DatabaseRouter interface {
	DBForRead(model *orm.ModelInfo) string
	DBForWrite(model *orm.ModelInfo) string
	AllowMigrate(db string, model *orm.ModelInfo) bool
}
