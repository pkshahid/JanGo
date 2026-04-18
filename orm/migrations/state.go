package migrations

import (
	"strings"

	"github.com/pkshahid/JanGo/orm"
)

// ProjectState represents the entire database schema state at a specific point in time.
type ProjectState struct {
	Models map[string]*ModelState
}

// NewProjectState initializes an empty project state.
func NewProjectState() *ProjectState {
	return &ProjectState{
		Models: make(map[string]*ModelState),
	}
}

// Clone creates a deep copy of the ProjectState.
func (p *ProjectState) Clone() *ProjectState {
	clone := NewProjectState()
	for k, m := range p.Models {
		clone.Models[k] = m.Clone()
	}
	return clone
}

// AddModel adds a new model state to the project.
func (p *ProjectState) AddModel(model *ModelState) {
	p.Models[model.Key()] = model
}

// RemoveModel removes a model state from the project.
func (p *ProjectState) RemoveModel(appName, name string) {
	key := strings.ToLower(appName + "." + name)
	delete(p.Models, key)
}

// ModelState represents the schema state of a single model.
type ModelState struct {
	AppName string
	Name    string
	Fields  []*orm.Field
	Meta    *orm.Meta
}

// Key returns the unique identifier for the model (app_name.model_name).
func (m *ModelState) Key() string {
	return strings.ToLower(m.AppName + "." + m.Name)
}

// Clone creates a deep copy of the ModelState.
func (m *ModelState) Clone() *ModelState {
	clone := &ModelState{
		AppName: m.AppName,
		Name:    m.Name,
		Fields:  make([]*orm.Field, len(m.Fields)),
		Meta:    m.Meta, // In a full implementation, deeply clone Meta
	}

	for i, f := range m.Fields {
		fClone := *f // Shallow copy struct
		clone.Fields[i] = &fClone
	}
	return clone
}

// GetField returns a field by name.
func (m *ModelState) GetField(name string) *orm.Field {
	for _, f := range m.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// RemoveField removes a field by name.
func (m *ModelState) RemoveField(name string) {
	for i, f := range m.Fields {
		if f.Name == name {
			m.Fields = append(m.Fields[:i], m.Fields[i+1:]...)
			return
		}
	}
}

// ProjectStateFromApps creates a ProjectState from the current ORM registry.
func ProjectStateFromApps() *ProjectState {
	state := NewProjectState()

	// Iterate through all registered models in the ORM.
	// Since orm package registry is internal and we only have `GetModelInfo` for specific structs,
	// we need a way to get ALL models. We will add a method to orm to fetch all ModelInfos.

	models := orm.AllModels()
	for _, info := range models {
		// Mock determining AppName. In Django, models belong to apps.
		// In Go, we'd use the package path or a registry mapping.
		// We'll use "main" or the package name as a proxy for app_name.
		parts := strings.Split(info.Type.PkgPath(), "/")
		appName := "main"
		if len(parts) > 0 {
			appName = parts[len(parts)-1]
			if appName == "" {
				appName = "main" // e.g. for simple main package models
			}
		}

		mState := &ModelState{
			AppName: appName,
			Name:    info.Name,
			Fields:  info.Fields,
			Meta:    info.Meta,
		}
		state.AddModel(mState)
	}

	return state
}
