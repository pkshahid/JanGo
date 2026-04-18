package commands

import (
	"context"
	"fmt"

	"github.com/godjango/godjango/core"
	"github.com/godjango/godjango/management"
	"github.com/spf13/cobra"
)

// VersionCmd prints the GoDjango version.
type VersionCmd struct{}

func (c *VersionCmd) Name() string {
	return "version"
}

func (c *VersionCmd) Help() string {
	return "Prints the GoDjango framework version."
}

func (c *VersionCmd) AddFlags(cmd *cobra.Command) {}

func (c *VersionCmd) Execute(ctx context.Context, args []string) error {
	fmt.Println(core.VERSION)
	return nil
}

func init() {
	management.Register(&VersionCmd{})
}
