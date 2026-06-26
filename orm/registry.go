package orm

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	globalRegistry = sync.Map{} // key: reflect.Type, value: *ModelInfo
)

// Register registers a model with the ORM.
// It parses the struct fields and caches the ModelInfo.
func Register(models ...any) error {
	for _, m := range models {
		info, err := parseModel(m)
		if err != nil {
			return fmt.Errorf("orm: failed to register model %T: %w", m, err)
		}

		t := reflect.TypeOf(m)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		if _, loaded := globalRegistry.LoadOrStore(t, info); loaded {
			return fmt.Errorf("orm: model %s is already registered", info.Name)
		}
	}
	return nil
}

// GetModelInfo returns the parsed ModelInfo for a given model struct.
func GetModelInfo(m any) (*ModelInfo, error) {
	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if info, ok := globalRegistry.Load(t); ok {
		return info.(*ModelInfo), nil
	}

	return nil, fmt.Errorf("orm: model %s is not registered", t.Name())
}

// GetModelInfoByType returns the parsed ModelInfo for a registered model
// identified by its reflect.Type. This is used internally by the queryset
// layer to resolve ForeignKey targets without an instance.
func GetModelInfoByType(t reflect.Type) (*ModelInfo, error) {
	if info, ok := globalRegistry.Load(t); ok {
		return info.(*ModelInfo), nil
	}
	return nil, fmt.Errorf("orm: model %s is not registered", t.Name())
}

// ClearRegistry clears the global model registry. Mostly used for testing.
func ClearRegistry() {
	globalRegistry.Range(func(key any, value any) bool {
		globalRegistry.Delete(key)
		return true
	})
}

// AllModels returns a slice of all registered ModelInfo structs.
func AllModels() []*ModelInfo {
	var models []*ModelInfo
	globalRegistry.Range(func(key, value any) bool {
		models = append(models, value.(*ModelInfo))
		return true
	})
	return models
}
