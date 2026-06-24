package commands

import (
	"bytes"
	"io"
	"testing"
)

// testCommand implements Command for testing.
type testCommand struct {
	*BaseCommand
}

func newTestCommand() *testCommand {
	return &testCommand{
		BaseCommand: NewBaseCommand("test", "A test command"),
	}
}

func (c *testCommand) Run(args []string, stdout io.Writer) error {
	c.Stdout = stdout
	c.Write("test output: %v", args)
	return nil
}

func TestCommandRegistry(t *testing.T) {
	ClearCommands()
	defer ClearCommands()

	cmd := newTestCommand()
	Register(cmd)

	// Test GetCommand
	got, ok := GetCommand("test")
	if !ok {
		t.Fatal("Expected to find 'test' command")
	}
	if got.Name() != "test" {
		t.Errorf("Expected name 'test', got %q", got.Name())
	}
	if got.Help() != "A test command" {
		t.Errorf("Expected help 'A test command', got %q", got.Help())
	}

	// Test ListCommands
	names := ListCommands()
	if len(names) != 1 || names[0] != "test" {
		t.Errorf("Expected [test], got %v", names)
	}

	// Test unknown command
	_, ok = GetCommand("nonexistent")
	if ok {
		t.Error("Expected not to find 'nonexistent' command")
	}
}

func TestCommandExecution(t *testing.T) {
	ClearCommands()
	defer ClearCommands()

	cmd := newTestCommand()
	Register(cmd)

	var buf bytes.Buffer
	got, _ := GetCommand("test")
	err := got.Run([]string{"arg1", "arg2"}, &buf)
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()
	if output != "test output: [arg1 arg2]" {
		t.Errorf("Expected 'test output: [arg1 arg2]', got %q", output)
	}
}

func TestManagementUtility(t *testing.T) {
	ClearCommands()
	defer ClearCommands()

	cmd := newTestCommand()
	Register(cmd)

	// Test help
	stdout := &bytes.Buffer{}
	mu := &ManagementUtility{
		Argv:   []string{"manage", "help"},
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	mu.Execute()
	if !bytes.Contains(stdout.Bytes(), []byte("Available commands")) {
		t.Error("Expected 'Available commands' in help output")
	}

	// Test executing command
	stdout.Reset()
	mu = &ManagementUtility{
		Argv:   []string{"manage", "test", "hello"},
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	}
	err := mu.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Test unknown command
	stderr := &bytes.Buffer{}
	mu = &ManagementUtility{
		Argv:   []string{"manage", "nonexistent"},
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
	}
	err = mu.Execute()
	if err == nil {
		t.Error("Expected error for unknown command")
	}
}

func TestBaseCommand(t *testing.T) {
	base := NewBaseCommand("mycommand", "Does something")

	if base.Name() != "mycommand" {
		t.Errorf("Expected name 'mycommand', got %q", base.Name())
	}
	if base.Help() != "Does something" {
		t.Errorf("Expected help 'Does something', got %q", base.Help())
	}

	// Test flag parsing
	var verbose bool
	base.FlagSet.BoolVar(&verbose, "verbose", false, "verbose mode")
	err := base.ParseFlags([]string{"-verbose"})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}
	if !verbose {
		t.Error("Expected verbose to be true")
	}
}
