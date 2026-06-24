package apps

import (
	"fmt"
	"sync"
)

// ModelInfo is a placeholder for model metadata.
type ModelInfo any

// AppConfig defines the interface for a GoDjango application.
type AppConfig interface {
	// Name returns the full module path or application name.
	Name() string
	// Label returns a short, unique identifier for the app. Defaults to Name() if empty.
	Label() string
	// Path returns the filesystem path to the app.
	Path() string
	// Models returns a list of models defined in this app.
	Models() []ModelInfo
	// Ready is called after all apps have been loaded and the registry is fully populated.
	Ready()
}

// Registry manages all registered GoDjango applications.
type Registry struct {
	mu      sync.RWMutex
	apps    map[string]AppConfig // Keyed by Name
	labels  map[string]string    // Maps Label to Name
	ordered []AppConfig          // Ordered by INSTALLED_APPS
	ready   bool                 // Indicates if Setup() has been completed
}

var (
	globalRegistry = &Registry{
		apps:   make(map[string]AppConfig),
		labels: make(map[string]string),
	}
)

// Register adds an AppConfig to the global registry.
// This is typically called from an app's init() function.
func Register(app AppConfig) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	name := app.Name()
	if name == "" {
		panic("AppConfig must have a non-empty Name()")
	}

	if _, exists := globalRegistry.apps[name]; exists {
		panic(fmt.Sprintf("AppConfig with name %q is already registered", name))
	}

	globalRegistry.apps[name] = app
}

// Setup initializes the registry based on the provided INSTALLED_APPS list.
// It discovers, validates, sets up load order, and calls Ready() on each app.
func Setup(installedApps []string) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if globalRegistry.ready {
		return fmt.Errorf("app registry is already populated and ready")
	}

	globalRegistry.ordered = make([]AppConfig, 0, len(installedApps))
	globalRegistry.labels = make(map[string]string)

	for _, appName := range installedApps {
		app, exists := globalRegistry.apps[appName]
		if !exists {
			return fmt.Errorf("app %q is in INSTALLED_APPS but has not been registered", appName)
		}

		label := app.Label()
		if label == "" {
			label = app.Name()
		}

		if existingName, exists := globalRegistry.labels[label]; exists {
			return fmt.Errorf("app label %q is already in use by app %q", label, existingName)
		}

		globalRegistry.labels[label] = app.Name()
		globalRegistry.ordered = append(globalRegistry.ordered, app)
	}

	// Call Ready() on all installed apps in order
	for _, app := range globalRegistry.ordered {
		app.Ready()
	}

	globalRegistry.ready = true
	return nil
}

// Get retrieves an AppConfig by its label.
func Get(label string) (AppConfig, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	if !globalRegistry.ready {
		return nil, fmt.Errorf("app registry is not ready yet")
	}

	name, exists := globalRegistry.labels[label]
	if !exists {
		return nil, fmt.Errorf("app with label %q not found", label)
	}

	app, exists := globalRegistry.apps[name]
	if !exists {
		return nil, fmt.Errorf("app with name %q not found", name)
	}

	return app, nil
}

// All returns a slice of all installed AppConfigs in the order they were specified in INSTALLED_APPS.
func All() []AppConfig {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	if !globalRegistry.ready {
		return nil
	}

	// Return a copy to prevent external modification
	result := make([]AppConfig, len(globalRegistry.ordered))
	copy(result, globalRegistry.ordered)
	return result
}

// IsInstalled returns true if an app with the given name is installed.
func IsInstalled(name string) bool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	if !globalRegistry.ready {
		return false
	}

	// Check if the app is in the installed list
	for _, app := range globalRegistry.ordered {
		if app.Name() == name {
			return true
		}
	}
	return false
}

// Reset clears the registry. Primarily used for testing.
func Reset() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	globalRegistry.apps = make(map[string]AppConfig)
	globalRegistry.labels = make(map[string]string)
	globalRegistry.ordered = nil
	globalRegistry.ready = false
}
