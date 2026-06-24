package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// ShellConfig holds configuration for the management shell.
type ShellConfig struct {
	Banner     string
	PromptFunc func() string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	Commands   map[string]ShellCommand
}

// ShellCommand represents a command available in the shell.
type ShellCommand struct {
	Name    string
	Help    string
	Handler func(args []string) (string, error)
}

// Shell provides an interactive management shell.
// While Go doesn't have a native REPL like Python, this provides an
// interactive command-line interface for management operations.
type Shell struct {
	config   ShellConfig
	commands map[string]ShellCommand
	running  bool
}

// NewShell creates a new management shell.
func NewShell(config ShellConfig) *Shell {
	if config.Stdin == nil {
		config.Stdin = os.Stdin
	}
	if config.Stdout == nil {
		config.Stdout = os.Stdout
	}
	if config.Stderr == nil {
		config.Stderr = os.Stderr
	}
	if config.Banner == "" {
		config.Banner = "JanGo interactive shell. Type 'help' for available commands."
	}
	if config.PromptFunc == nil {
		config.PromptFunc = func() string { return ">>> " }
	}

	s := &Shell{
		config:   config,
		commands: make(map[string]ShellCommand),
	}

	// Register built-in commands
	s.registerBuiltins()

	// Register user commands
	for name, cmd := range config.Commands {
		s.commands[name] = cmd
	}

	return s
}

// RegisterCommand adds a command to the shell.
func (s *Shell) RegisterCommand(name, help string, handler func([]string) (string, error)) {
	s.commands[name] = ShellCommand{
		Name:    name,
		Help:    help,
		Handler: handler,
	}
}

// Run starts the interactive shell loop.
func (s *Shell) Run() error {
	fmt.Fprintln(s.config.Stdout, s.config.Banner)
	s.running = true

	scanner := bufio.NewScanner(s.config.Stdin)
	for s.running {
		fmt.Fprint(s.config.Stdout, s.config.PromptFunc())
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		output, err := s.Execute(line)
		if err != nil {
			fmt.Fprintf(s.config.Stderr, "Error: %v\n", err)
		} else if output != "" {
			fmt.Fprintln(s.config.Stdout, output)
		}
	}

	return scanner.Err()
}

// Execute runs a single command line.
func (s *Shell) Execute(line string) (string, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "", nil
	}

	cmdName := parts[0]
	args := parts[1:]

	cmd, exists := s.commands[cmdName]
	if !exists {
		return "", fmt.Errorf("unknown command: %s. Type 'help' for available commands", cmdName)
	}

	return cmd.Handler(args)
}

// Stop terminates the shell loop.
func (s *Shell) Stop() {
	s.running = false
}

func (s *Shell) registerBuiltins() {
	s.commands["help"] = ShellCommand{
		Name: "help",
		Help: "Show available commands",
		Handler: func(args []string) (string, error) {
			var sb strings.Builder
			sb.WriteString("Available commands:\n")
			for name, cmd := range s.commands {
				sb.WriteString(fmt.Sprintf("  %-20s %s\n", name, cmd.Help))
			}
			return sb.String(), nil
		},
	}

	s.commands["exit"] = ShellCommand{
		Name: "exit",
		Help: "Exit the shell",
		Handler: func(args []string) (string, error) {
			s.Stop()
			return "Goodbye!", nil
		},
	}

	s.commands["quit"] = ShellCommand{
		Name: "quit",
		Help: "Exit the shell",
		Handler: func(args []string) (string, error) {
			s.Stop()
			return "Goodbye!", nil
		},
	}

	s.commands["models"] = ShellCommand{
		Name: "models",
		Help: "List registered models",
		Handler: func(args []string) (string, error) {
			return "Use orm.ListModels() to see registered models.", nil
		},
	}

	s.commands["sql"] = ShellCommand{
		Name: "sql",
		Help: "Execute raw SQL (sql <query>)",
		Handler: func(args []string) (string, error) {
			if len(args) == 0 {
				return "", fmt.Errorf("usage: sql <query>")
			}
			query := strings.Join(args, " ")
			return fmt.Sprintf("Would execute: %s", query), nil
		},
	}

	s.commands["settings"] = ShellCommand{
		Name: "settings",
		Help: "Show current settings",
		Handler: func(args []string) (string, error) {
			return "Use settings.Get() to view current settings.", nil
		},
	}
}
