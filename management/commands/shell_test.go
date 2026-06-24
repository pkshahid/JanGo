package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestShellCreation(t *testing.T) {
	shell := NewShell(ShellConfig{})

	if shell == nil {
		t.Fatal("Expected non-nil shell")
	}

	// Check built-in commands registered
	builtins := []string{"help", "exit", "quit", "models", "sql", "settings"}
	for _, name := range builtins {
		if _, ok := shell.commands[name]; !ok {
			t.Errorf("Expected built-in command %q to be registered", name)
		}
	}
}

func TestShellExecute(t *testing.T) {
	shell := NewShell(ShellConfig{})

	// Test help command
	output, err := shell.Execute("help")
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}
	if !strings.Contains(output, "Available commands") {
		t.Error("Expected 'Available commands' in help output")
	}

	// Test unknown command
	_, err = shell.Execute("unknown_cmd")
	if err == nil {
		t.Error("Expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("Expected 'unknown command' in error, got: %v", err)
	}

	// Test exit command
	output, err = shell.Execute("exit")
	if err != nil {
		t.Fatalf("exit command failed: %v", err)
	}
	if output != "Goodbye!" {
		t.Errorf("Expected 'Goodbye!', got %q", output)
	}
	if shell.running {
		t.Error("Shell should be stopped after exit")
	}

	// Test sql command without args
	shell.running = true
	_, err = shell.Execute("sql")
	if err == nil {
		t.Error("Expected error for sql without args")
	}

	// Test sql command with args
	output, err = shell.Execute("sql SELECT * FROM users")
	if err != nil {
		t.Fatalf("sql command failed: %v", err)
	}
	if !strings.Contains(output, "Would execute") {
		t.Error("Expected 'Would execute' in sql output")
	}
}

func TestShellRegisterCommand(t *testing.T) {
	shell := NewShell(ShellConfig{})

	shell.RegisterCommand("greet", "Greet someone", func(args []string) (string, error) {
		if len(args) == 0 {
			return "Hello, World!", nil
		}
		return "Hello, " + strings.Join(args, " ") + "!", nil
	})

	output, err := shell.Execute("greet")
	if err != nil {
		t.Fatalf("greet command failed: %v", err)
	}
	if output != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got %q", output)
	}

	output, err = shell.Execute("greet Alice")
	if err != nil {
		t.Fatalf("greet Alice command failed: %v", err)
	}
	if output != "Hello, Alice!" {
		t.Errorf("Expected 'Hello, Alice!', got %q", output)
	}
}

func TestShellRun(t *testing.T) {
	input := "help\nexit\n"
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	shell := NewShell(ShellConfig{
		Stdin:  strings.NewReader(input),
		Stdout: stdout,
		Stderr: stderr,
	})

	err := shell.Run()
	if err != nil {
		t.Fatalf("Shell.Run() failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "JanGo interactive shell") {
		t.Error("Expected banner in output")
	}
	if !strings.Contains(output, "Available commands") {
		t.Error("Expected help output")
	}
	if !strings.Contains(output, "Goodbye!") {
		t.Error("Expected goodbye in output")
	}
}

func TestShellCustomConfig(t *testing.T) {
	commands := map[string]ShellCommand{
		"ping": {
			Name: "ping",
			Help: "Return pong",
			Handler: func(args []string) (string, error) {
				return "pong", nil
			},
		},
	}

	shell := NewShell(ShellConfig{
		Banner:   "Custom Shell",
		Commands: commands,
	})

	output, err := shell.Execute("ping")
	if err != nil {
		t.Fatalf("ping command failed: %v", err)
	}
	if output != "pong" {
		t.Errorf("Expected 'pong', got %q", output)
	}
}
