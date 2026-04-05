package management

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
)

// Command defines the interface that all management commands must implement.
type Command interface {
	Name() string
	Help() string
	AddFlags(cmd *cobra.Command)
	Execute(ctx context.Context, args []string) error
}

var (
	commands = make(map[string]Command)
)

// Register adds a command to the registry. It is typically called in an init() function.
func Register(cmd Command) {
	if _, exists := commands[cmd.Name()]; exists {
		panic(fmt.Sprintf("command %q is already registered", cmd.Name()))
	}
	commands[cmd.Name()] = cmd
}

// Execute is the root dispatcher for management commands.
func Execute(args []string) error {
	rootCmd := &cobra.Command{
		Use:   "manage",
		Short: "GoDjango management console",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add all registered commands
	for _, cmd := range commands {
		cmdObj := cmd // capture loop variable
		cobraCmd := &cobra.Command{
			Use:   cmdObj.Name(),
			Short: cmdObj.Help(),
			RunE: func(c *cobra.Command, a []string) error {
				return cmdObj.Execute(c.Context(), a)
			},
		}
		cmdObj.AddFlags(cobraCmd)
		rootCmd.AddCommand(cobraCmd)
	}

	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
