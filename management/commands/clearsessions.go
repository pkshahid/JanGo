package commands

import (
	"context"
	"fmt"

	"github.com/godjango/godjango/core/settings"
	"github.com/godjango/godjango/management"
	"github.com/godjango/godjango/sessions/backends"
	"github.com/spf13/cobra"
)

func init() {
	management.Register(&ClearSessionsCommand{})
}

type ClearSessionsCommand struct{}

func (c *ClearSessionsCommand) Name() string { return "clearsessions" }
func (c *ClearSessionsCommand) Help() string { return "Cleans out expired sessions." }
func (c *ClearSessionsCommand) AddFlags(cmd *cobra.Command) {}

func (c *ClearSessionsCommand) Execute(ctx context.Context, args []string) error {
	s := settings.Get()
	engine := s.SESSION_ENGINE
	if engine == "" {
		engine = "file"
	}

	var backend interface {
		ClearExpired(context.Context) error
	}

	switch engine {
	case "db":
		backend = &backends.DatabaseBackend{}
	case "file":
		backend = &backends.FileBackend{}
	case "cache":
		backend = &backends.CacheBackend{}
	case "cookie":
		// Cookie backend stores everything on the client. No server clearing needed.
		fmt.Println("clearsessions is not necessary with the cookie session backend.")
		return nil
	default:
		backend = &backends.FileBackend{}
	}

	fmt.Printf("Clearing expired %s sessions...\n", engine)

	err := backend.ClearExpired(ctx)
	if err != nil {
		return fmt.Errorf("failed to clear sessions: %v", err)
	}

	fmt.Println("Done.")
	return nil
}
