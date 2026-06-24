// Package contenttypes implements Django's contenttypes framework.
// It provides a registry of all models (content types) that enables
// generic relations and polymorphic model lookups.
package contenttypes

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// ContentType represents a model type registered in the system.
// Equivalent to Django's django.contrib.contenttypes.models.ContentType.
type ContentType struct {
	ID       int
	AppLabel string
	Model    string
	GoType   reflect.Type
}

// String returns a human-readable label for the content type.
func (ct *ContentType) String() string {
	return fmt.Sprintf("%s | %s", ct.AppLabel, ct.Model)
}

// ModelClass returns the reflect.Type of the underlying Go struct.
func (ct *ContentType) ModelClass() reflect.Type {
	return ct.GoType
}

// Registry manages all registered content types.
type Registry struct {
	mu     sync.RWMutex
	types  map[int]*ContentType
	byKey  map[string]*ContentType // key = "app_label.model"
	nextID int
}

var globalRegistry = &Registry{
	types:  make(map[int]*ContentType),
	byKey:  make(map[string]*ContentType),
	nextID: 1,
}

// Register adds a model to the content type registry.
// appLabel is the application name, model is the model name (lowercase).
func Register(appLabel, model string, goType reflect.Type) *ContentType {
	return globalRegistry.Register(appLabel, model, goType)
}

func (r *Registry) Register(appLabel, model string, goType reflect.Type) *ContentType {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := strings.ToLower(appLabel) + "." + strings.ToLower(model)
	if ct, exists := r.byKey[key]; exists {
		return ct
	}

	ct := &ContentType{
		ID:       r.nextID,
		AppLabel: strings.ToLower(appLabel),
		Model:    strings.ToLower(model),
		GoType:   goType,
	}
	r.types[ct.ID] = ct
	r.byKey[key] = ct
	r.nextID++
	return ct
}

// Get retrieves a content type by ID.
func Get(id int) (*ContentType, error) {
	return globalRegistry.Get(id)
}

func (r *Registry) Get(id int) (*ContentType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ct, exists := r.types[id]
	if !exists {
		return nil, fmt.Errorf("contenttypes: no content type with ID %d", id)
	}
	return ct, nil
}

// GetForModel retrieves the content type for a given app and model name.
func GetForModel(appLabel, model string) (*ContentType, error) {
	return globalRegistry.GetForModel(appLabel, model)
}

func (r *Registry) GetForModel(appLabel, model string) (*ContentType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := strings.ToLower(appLabel) + "." + strings.ToLower(model)
	ct, exists := r.byKey[key]
	if !exists {
		return nil, fmt.Errorf("contenttypes: no content type for %s.%s", appLabel, model)
	}
	return ct, nil
}

// GetForGoType retrieves the content type for a given Go type.
func GetForGoType(t reflect.Type) (*ContentType, error) {
	return globalRegistry.GetForGoType(t)
}

func (r *Registry) GetForGoType(t reflect.Type) (*ContentType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, ct := range r.types {
		if ct.GoType == t {
			return ct, nil
		}
	}
	return nil, fmt.Errorf("contenttypes: no content type for Go type %v", t)
}

// All returns all registered content types.
func All() []*ContentType {
	return globalRegistry.All()
}

func (r *Registry) All() []*ContentType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*ContentType, 0, len(r.types))
	for _, ct := range r.types {
		result = append(result, ct)
	}
	return result
}

// Clear removes all registered content types (for testing).
func Clear() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.types = make(map[int]*ContentType)
	globalRegistry.byKey = make(map[string]*ContentType)
	globalRegistry.nextID = 1
}
