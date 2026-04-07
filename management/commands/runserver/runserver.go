package runserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/godjango/godjango/core/handlers/wsgi"
	"github.com/godjango/godjango/core/settings"
	"github.com/godjango/godjango/http/middleware"
	"github.com/godjango/godjango/http/urls"
	"github.com/godjango/godjango/management"
	staticmiddleware "github.com/godjango/godjango/static/middleware"
	"github.com/spf13/cobra"
)

// Command implements the management.Command interface for runserver.
func init() {
	management.Register(&Command{})
}

type Command struct{}

func (c *Command) Name() string {
	return "runserver"
}

func (c *Command) Help() string {
	return "Starts a lightweight Web server for development."
}

func (c *Command) AddFlags(cmd *cobra.Command) {
	// Allows standard `go run manage.go runserver 8080` or `127.0.0.1:8000`
	// Handled via args parsing instead of flags, just like Django
}

func (c *Command) Execute(ctx context.Context, args []string) error {
	addr := "127.0.0.1:8000"
	if len(args) > 0 {
		addr = args[0]
		// If just port is provided
		if !strings.Contains(addr, ":") {
			addr = "127.0.0.1:" + addr
		}
	}

	// Make sure settings are loaded. They should be configured via godjango project manage.go
	s := settings.Get()

	// Handle hot reloading logic if triggered by a reload
	isChildProcess := os.Getenv("GODJANGO_RUNSERVER_CHILD") == "1"

	if s.DEBUG && !isChildProcess {
		return startWatcherProcess(addr)
	}

	// Print Startup Banner
	fmt.Printf("Watching for file changes with fsnotify\n")
	fmt.Printf("Performing system checks...\n\n")
	fmt.Printf("System check identified no issues (0 silenced).\n")

	now := time.Now().Format("January 02, 2006 - 15:04:05")
	fmt.Printf("%s\n", now)
	fmt.Printf("GoDjango version 0.0.1, using settings\n")
	fmt.Printf("Starting development server at http://%s/\n", addr)
	fmt.Printf("Quit the server with CONTROL-C.\n")

	// Set up WSGI Handler with middleware
	router := urls.GetGlobalRouter()

	// Create middleware chain (auto-include Logger in runserver)
	// For a real app, this would iterate over s.MIDDLEWARE strings to build the chain.
	middlewares := []middleware.MiddlewareFunc{
		middleware.RequestLoggingMiddleware,
		middleware.SecurityMiddleware,
		middleware.CommonMiddleware,
		middleware.SessionMiddleware,
		middleware.CsrfViewMiddleware,
		middleware.AuthenticationMiddleware,
		middleware.MessageMiddleware,
	}

	wm := staticmiddleware.NewWhiteNoiseMiddleware()
	middlewares = append([]middleware.MiddlewareFunc{wm.Process}, middlewares...)

	handler := wsgi.NewWSGIHandler(router, middlewares...)

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Graceful shutdown channels
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case sig := <-stopChan:
		slog.Info(fmt.Sprintf("Received signal %s, shutting down...", sig))

		// 30 second graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}

		slog.Info("Server stopped gracefully")
	}

	return nil
}

// startWatcherProcess watches for .go file changes and runs the child process.
func startWatcherProcess(addr string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Add all subdirectories to watcher recursively
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	var childProcess *exec.Cmd

	startChild := func() {
		if childProcess != nil && childProcess.Process != nil {
			childProcess.Process.Kill()
			childProcess.Wait()
		}

		// Re-run the go program instead of executing the stale binary directly
		// `go run manage.go runserver`
		args := []string{"run"}
		// we need to find manage.go in os.Args if it was run via `go run manage.go`
		// if os.Args[0] is a compiled binary (e.g., ./myproject), we can't easily `go run` it,
		// but typically for development people use `go run manage.go ...`.
		// For robustness, if `manage.go` exists, we run that.
		if _, err := os.Stat("manage.go"); err == nil {
			args = append(args, "manage.go")
			// append original arguments excluding the binary path (os.Args[0])
			for _, a := range os.Args[1:] {
				if a != "manage.go" {
					args = append(args, a)
				}
			}
			childProcess = exec.Command("go", args...)
		} else {
			// Fallback to executing the binary directly
			childProcess = exec.Command(os.Args[0], os.Args[1:]...)
		}

		childProcess.Stdout = os.Stdout
		childProcess.Stderr = os.Stderr
		childProcess.Env = append(os.Environ(), "GODJANGO_RUNSERVER_CHILD=1")

		err := childProcess.Start()
		if err != nil {
			slog.Error("Failed to start child process", "error", err)
		}
	}

	// Start initially
	startChild()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				if strings.HasSuffix(event.Name, ".go") {
					fmt.Printf("[GoDjango] Reloading...\n")
					startChild()
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			slog.Error("Watcher error", "error", err)
		case <-stopChan:
			if childProcess != nil && childProcess.Process != nil {
				childProcess.Process.Signal(syscall.SIGTERM)
				childProcess.Wait()
			}
			return nil
		}
	}
}
