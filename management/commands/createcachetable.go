package commands

import (
	"context"
	"fmt"

	"github.com/godjango/godjango/core/settings"
	"github.com/godjango/godjango/management"
	"github.com/godjango/godjango/orm/backends"
	"github.com/spf13/cobra"
)

func init() {
	management.Register(&CreateCacheTableCommand{})
}

type CreateCacheTableCommand struct {
	database string
	dryRun   bool
}

func (c *CreateCacheTableCommand) Name() string {
	return "createcachetable"
}

func (c *CreateCacheTableCommand) Help() string {
	return "Creates the cache table in the database."
}

func (c *CreateCacheTableCommand) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.database, "database", "default", "Nominates a database to put the cache table in.")
	cmd.Flags().BoolVar(&c.dryRun, "dry-run", false, "Do not run the SQL.")
}

func (c *CreateCacheTableCommand) Execute(ctx context.Context, args []string) error {
	tableName := "cache_table"
	if len(args) > 0 {
		tableName = args[0]
	}

	s := settings.Get()
	// Loop over caches, find ones that are DatabaseCache
	for _, config := range s.CACHES {
		if config.Backend == "DatabaseCache" {
			if loc, ok := config.Options["LOCATION"].(string); ok && loc != "" {
				tableName = loc
			}
		}
	}

	// Simple dialect detection
	dbConfig := s.DATABASES[c.database]

	var sql string
	if dbConfig.Engine == "postgres" || dbConfig.Engine == "mysql" {
		sql = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			cache_key VARCHAR(255) PRIMARY KEY,
			value TEXT NOT NULL,
			expires TIMESTAMP NOT NULL
		);`, tableName)
	} else {
		// SQLite
		sql = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			cache_key VARCHAR(255) PRIMARY KEY,
			value TEXT NOT NULL,
			expires DATETIME NOT NULL
		);`, tableName)
	}

	if c.dryRun {
		fmt.Printf("Dry-run executing SQL for %s:\n%s\n", c.database, sql)
		return nil
	}

	// Real ORM manager integration
	db, err := backends.GetConnection(c.database)
	if err != nil {
		return fmt.Errorf("could not connect to database %s: %w", c.database, err)
	}

	_, err = db.ExecContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to create cache table: %w", err)
	}

	fmt.Println("Cache table created successfully.")
	return nil
}
