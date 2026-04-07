package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/godjango/godjango/management"
	"github.com/godjango/godjango/orm/backends"
	"github.com/godjango/godjango/orm/migrations"
	"github.com/spf13/cobra"
)

func init() {
	management.Register(&MigrateCommand{})
	management.Register(&MakeMigrationsCommand{})
}

type MigrateCommand struct{}

func (c *MigrateCommand) Name() string { return "migrate" }
func (c *MigrateCommand) Help() string { return "Updates database schema." }
func (c *MigrateCommand) AddFlags(cmd *cobra.Command) {}

func (c *MigrateCommand) Execute(ctx context.Context, args []string) error {
	fmt.Println("Operations to perform:")
	fmt.Println("  Apply all migrations: admin, auth, contenttypes, sessions") // Mock output similar to Django

	if err := backends.Init(); err != nil {
		return err
	}
	defer backends.Close()

	executor, err := migrations.NewMigrationExecutor("default")
	if err != nil {
		return err
	}

	plan, err := executor.Plan(ctx, "")
	if err != nil {
		return err
	}

	if len(plan) == 0 {
		fmt.Println("Running migrations:\n  No migrations to apply.")
		return nil
	}

	fmt.Println("Running migrations:")
	for _, p := range plan {
		fmt.Printf("  Applying %s.%s... ", p.Migration.App, p.Migration.Name)
		err := executor.Execute(ctx, []migrations.MigrationPlan{p})
		if err != nil {
			fmt.Println("FAILED")
			return err
		}
		fmt.Println("OK")
	}

	return nil
}

type MakeMigrationsCommand struct{}

func (c *MakeMigrationsCommand) Name() string { return "makemigrations" }
func (c *MakeMigrationsCommand) Help() string { return "Creates new migrations based on the changes detected to your models." }
func (c *MakeMigrationsCommand) AddFlags(cmd *cobra.Command) {}

func (c *MakeMigrationsCommand) Execute(ctx context.Context, args []string) error {
	// Rebuild current project state from currently loaded schema.
	// Normally we would build from existing migrations applied forwards,
	// then compare with current model state from registry.

	graph, _ := migrations.BuildGraph()
	oldState := migrations.NewProjectState()

	plan, _ := graph.ForwardsPlan()
	for _, m := range plan {
		for _, op := range m.Operations {
			op.StateForwards(m.App, oldState)
		}
	}

	newState := migrations.ProjectStateFromApps()

	autodetector := migrations.NewAutodetector(oldState, newState)
	changes := autodetector.Changes()

	if len(changes) == 0 {
		fmt.Println("No changes detected")
		return nil
	}

	migrationsMaps := migrations.ArrangeDependencies(changes, graph)

	baseDir, _ := os.Getwd() // Current project dir

	for app, m := range migrationsMaps {
		fmt.Printf("Migrations for '%s':\n", app)
		fmt.Printf("  %s/migrations/%s.go\n", app, m.Name)
		for _, op := range m.Operations {
			fmt.Printf("    - %T\n", op)
		}

		writer := migrations.NewMigrationWriter(m, baseDir)
		_, err := writer.WriteToFile()
		if err != nil {
			return fmt.Errorf("failed to write migration: %v", err)
		}
	}

	return nil
}
