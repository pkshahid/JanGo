package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/godjango/godjango/orm/backends"
)

// MigrationRecord represents a row in the godjango_migrations table.
type MigrationRecord struct {
	ID      int
	App     string
	Name    string
	Applied time.Time
}

// MigrationExecutor manages running migrations against a database.
type MigrationExecutor struct {
	Backend      backends.Backend
	SchemaEditor backends.SchemaEditor
	DbAlias      string
}

// NewMigrationExecutor creates an executor for the given database alias.
func NewMigrationExecutor(dbAlias string) (*MigrationExecutor, error) {
	backend, err := backends.GetBackend(dbAlias)
	if err != nil {
		return nil, err
	}
	return &MigrationExecutor{
		Backend:      backend,
		SchemaEditor: backend.SchemaEditor(),
		DbAlias:      dbAlias,
	}, nil
}

// EnsureMigrationsTable creates the migrations table if it doesn't exist.
func (e *MigrationExecutor) EnsureMigrationsTable(ctx context.Context) error {
	// A simple raw SQL implementation for creating the table
	sqlStr := `CREATE TABLE IF NOT EXISTS godjango_migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		applied_at DATETIME NOT NULL
	);`

	if e.Backend.DatabaseName() == "postgres" {
		sqlStr = `CREATE TABLE IF NOT EXISTS godjango_migrations (
			id SERIAL PRIMARY KEY,
			app VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL
		);`
	}

	_, err := e.Backend.Execute(ctx, sqlStr)
	return err
}

// GetAppliedMigrations retrieves all applied migrations from the database.
func (e *MigrationExecutor) GetAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	if err := e.EnsureMigrationsTable(ctx); err != nil {
		return nil, err
	}

	rows, err := e.Backend.Query(ctx, "SELECT app, name FROM godjango_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var app, name string
		if err := rows.Scan(&app, &name); err != nil {
			return nil, err
		}
		applied[fmt.Sprintf("%s.%s", app, name)] = true
	}
	return applied, nil
}

// RecordMigration marks a migration as applied.
func (e *MigrationExecutor) RecordMigration(ctx context.Context, tx *sql.Tx, app, name string) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO godjango_migrations (app, name, applied_at) VALUES (?, ?, ?)", app, name, time.Now())
	// In Postgres we'd use $1, $2. For prototype we stick to ?, which works for sqlite/mysql.
	// Postgres adaptation needed for production.
	return err
}

// UnrecordMigration removes a migration from the table.
func (e *MigrationExecutor) UnrecordMigration(ctx context.Context, tx *sql.Tx, app, name string) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM godjango_migrations WHERE app = ? AND name = ?", app, name)
	return err
}

// MigrationPlan represents a migration and its intended direction.
type MigrationPlan struct {
	Migration *Migration
	Backwards bool
}

// Plan calculates the migrations to apply or unapply to reach a target.
// Target format: "app.name" or "" for all.
func (e *MigrationExecutor) Plan(ctx context.Context, target string) ([]MigrationPlan, error) {
	graph, err := BuildGraph()
	if err != nil {
		return nil, err
	}

	forwardsPlan, err := graph.ForwardsPlan()
	if err != nil {
		return nil, err
	}

	applied, err := e.GetAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	var plan []MigrationPlan

	// Basic prototype: apply all unapplied
	// A full framework plans forwards up to target or backwards down to target.
	for _, m := range forwardsPlan {
		key := fmt.Sprintf("%s.%s", m.App, m.Name)
		if !applied[key] {
			plan = append(plan, MigrationPlan{Migration: m, Backwards: false})
		}
	}

	return plan, nil
}

// Execute applies a plan within transactions.
func (e *MigrationExecutor) Execute(ctx context.Context, plan []MigrationPlan) error {
	// Rebuild state up to the starting point
	currentState := NewProjectState()

	// In a real implementation, we'd apply state operations for all previously applied migrations
	// to get the current state. For this prototype, we'll start empty.

	for _, p := range plan {
		m := p.Migration

		err := backends.Atomic(ctx, e.DbAlias, func(txCtx context.Context, tx *sql.Tx) error {
			if !p.Backwards {
				fromState := currentState.Clone()

				// Apply operations
				for _, op := range m.Operations {
					// 1. State forward
					op.StateForwards(m.App, currentState)

					// 2. Database forward
					if err := op.DatabaseForwards(m.App, e.SchemaEditor, fromState, currentState); err != nil {
						return fmt.Errorf("error applying %s.%s operation %T: %w", m.App, m.Name, op, err)
					}

					fromState = currentState.Clone()
				}

				// Record it
				return e.RecordMigration(txCtx, tx, m.App, m.Name)
			} else {
				// Reverse logic omitted for brevity in prototype.
				return nil
			}
		})

		if err != nil {
			return err
		}
	}

	return nil
}
