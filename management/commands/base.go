// Package commands provides the Django-style custom management commands framework.
// Users define commands by implementing the Command interface and registering them.
package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Command is the interface all management commands must implement.
type Command interface {
	Name() string
	Help() string
	Run(args []string, stdout io.Writer) error
}

// BaseCommand provides a reusable base for custom management commands.
type BaseCommand struct {
	CommandName string
	HelpText    string
	FlagSet     *flag.FlagSet
	Stdout      io.Writer
	Stderr      io.Writer
}

// NewBaseCommand creates a new base command with defaults.
func NewBaseCommand(name, help string) *BaseCommand {
	return &BaseCommand{
		CommandName: name,
		HelpText:    help,
		FlagSet:     flag.NewFlagSet(name, flag.ContinueOnError),
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
	}
}

// Name returns the command name.
func (b *BaseCommand) Name() string { return b.CommandName }

// Help returns the help text.
func (b *BaseCommand) Help() string { return b.HelpText }

// ParseFlags parses command-line flags.
func (b *BaseCommand) ParseFlags(args []string) error {
	return b.FlagSet.Parse(args)
}

// Write outputs to stdout.
func (b *BaseCommand) Write(format string, args ...any) {
	fmt.Fprintf(b.Stdout, format, args...)
}

// WriteError outputs to stderr.
func (b *BaseCommand) WriteError(format string, args ...any) {
	fmt.Fprintf(b.Stderr, format, args...)
}

// CommandRegistry holds all registered management commands.
type CommandRegistry struct {
	commands map[string]Command
}

var globalRegistry = &CommandRegistry{
	commands: make(map[string]Command),
}

// Register adds a command to the global registry.
func Register(cmd Command) {
	globalRegistry.commands[cmd.Name()] = cmd
}

// GetCommand returns a registered command by name.
func GetCommand(name string) (Command, bool) {
	cmd, ok := globalRegistry.commands[name]
	return cmd, ok
}

// ListCommands returns all registered command names sorted alphabetically.
func ListCommands() []string {
	names := make([]string, 0, len(globalRegistry.commands))
	for name := range globalRegistry.commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ClearCommands removes all registered commands (for testing).
func ClearCommands() {
	globalRegistry.commands = make(map[string]Command)
}

// Execute runs a management command by name with the given arguments.
func Execute(name string, args []string) error {
	cmd, ok := GetCommand(name)
	if !ok {
		return fmt.Errorf("unknown command %q. Available commands:\n%s",
			name, strings.Join(ListCommands(), "\n"))
	}
	return cmd.Run(args, os.Stdout)
}

// ManagementUtility is the main entry point for management commands.
// It's analogous to Django's ManagementUtility.
type ManagementUtility struct {
	Argv   []string
	Stdout io.Writer
	Stderr io.Writer
}

// NewManagementUtility creates a new utility with the given arguments.
func NewManagementUtility(argv []string) *ManagementUtility {
	return &ManagementUtility{
		Argv:   argv,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// Execute processes command-line arguments and runs the appropriate command.
func (mu *ManagementUtility) Execute() error {
	if len(mu.Argv) < 2 {
		mu.showHelp()
		return nil
	}

	cmdName := mu.Argv[1]
	args := mu.Argv[2:]

	if cmdName == "help" || cmdName == "--help" || cmdName == "-h" {
		mu.showHelp()
		return nil
	}

	cmd, ok := GetCommand(cmdName)
	if !ok {
		fmt.Fprintf(mu.Stderr, "Unknown command: %s\n\n", cmdName)
		mu.showHelp()
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	return cmd.Run(args, mu.Stdout)
}

func (mu *ManagementUtility) showHelp() {
	fmt.Fprintln(mu.Stdout, "Available commands:")
	fmt.Fprintln(mu.Stdout)
	for _, name := range ListCommands() {
		cmd, _ := GetCommand(name)
		fmt.Fprintf(mu.Stdout, "  %-20s %s\n", name, cmd.Help())
	}
}
