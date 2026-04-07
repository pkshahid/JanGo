package migrations

import (
	"fmt"

	"github.com/godjango/godjango/orm"
	"github.com/godjango/godjango/orm/backends"
)

// Dep represents a dependency on another migration.
type Dep struct {
	App  string
	Name string
}

// Migration represents a single database migration file.
type Migration struct {
	App          string
	Name         string
	Dependencies []Dep
	Operations   []Operation
}

// Operation defines a single schema manipulation action.
type Operation interface {
	StateForwards(appLabel string, state *ProjectState)
	DatabaseForwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error
	DatabaseBackwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error
}

// ==========================================
// Operations implementations
// ==========================================

type CreateModel struct {
	Name   string
	Fields []*orm.Field
	Meta   *orm.Meta
}

func (o CreateModel) StateForwards(appLabel string, state *ProjectState) {
	mState := &ModelState{
		AppName: appLabel,
		Name:    o.Name,
		Fields:  o.Fields,
		Meta:    o.Meta,
	}
	state.AddModel(mState)
}

func (o CreateModel) DatabaseForwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	mState := toState.Models[appLabel+"."+o.Name]

	// Convert ModelState to ModelInfo
	info := &orm.ModelInfo{
		Name: mState.Name,
		Fields: mState.Fields,
		Meta: mState.Meta,
	}
	return schemaEditor.CreateTable(info)
}

func (o CreateModel) DatabaseBackwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	mState := fromState.Models[appLabel+"."+o.Name]
	info := &orm.ModelInfo{
		Name: mState.Name,
		Fields: mState.Fields,
		Meta: mState.Meta,
	}
	return schemaEditor.DeleteTable(info)
}

type DeleteModel struct {
	Name string
}

func (o DeleteModel) StateForwards(appLabel string, state *ProjectState) {
	state.RemoveModel(appLabel, o.Name)
}

func (o DeleteModel) DatabaseForwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	mState := fromState.Models[appLabel+"."+o.Name]
	info := &orm.ModelInfo{
		Name: mState.Name,
		Fields: mState.Fields,
		Meta: mState.Meta,
	}
	return schemaEditor.DeleteTable(info)
}

func (o DeleteModel) DatabaseBackwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	mState := toState.Models[appLabel+"."+o.Name]
	info := &orm.ModelInfo{
		Name: mState.Name,
		Fields: mState.Fields,
		Meta: mState.Meta,
	}
	return schemaEditor.CreateTable(info)
}

type AddField struct {
	Model string
	Name  string
	Field *orm.Field
}

func (o AddField) StateForwards(appLabel string, state *ProjectState) {
	mState := state.Models[appLabel+"."+o.Model]
	if mState != nil {
		mState.Fields = append(mState.Fields, o.Field)
	}
}

func (o AddField) DatabaseForwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	mState := toState.Models[appLabel+"."+o.Model]
	info := &orm.ModelInfo{Name: mState.Name, Fields: mState.Fields, Meta: mState.Meta}
	return schemaEditor.AddColumn(info, o.Field)
}

func (o AddField) DatabaseBackwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	mState := fromState.Models[appLabel+"."+o.Model]
	info := &orm.ModelInfo{Name: mState.Name, Fields: mState.Fields, Meta: mState.Meta}
	return schemaEditor.RemoveColumn(info, o.Field.Column)
}

type RemoveField struct {
	Model string
	Name  string
	Field *orm.Field // Preserved for backwards creation
}

func (o RemoveField) StateForwards(appLabel string, state *ProjectState) {
	mState := state.Models[appLabel+"."+o.Model]
	if mState != nil {
		mState.RemoveField(o.Name)
	}
}

func (o RemoveField) DatabaseForwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	mState := fromState.Models[appLabel+"."+o.Model]
	info := &orm.ModelInfo{Name: mState.Name, Fields: mState.Fields, Meta: mState.Meta}
	return schemaEditor.RemoveColumn(info, o.Field.Column)
}

func (o RemoveField) DatabaseBackwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	mState := toState.Models[appLabel+"."+o.Model]
	info := &orm.ModelInfo{Name: mState.Name, Fields: mState.Fields, Meta: mState.Meta}
	return schemaEditor.AddColumn(info, o.Field)
}

type AlterField struct {
	Model string
	Name  string
	Field *orm.Field
}

func (o AlterField) StateForwards(appLabel string, state *ProjectState) {
	mState := state.Models[appLabel+"."+o.Model]
	if mState != nil {
		for i, f := range mState.Fields {
			if f.Name == o.Name {
				mState.Fields[i] = o.Field
				return
			}
		}
	}
}

func (o AlterField) DatabaseForwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	mState := toState.Models[appLabel+"."+o.Model]
	info := &orm.ModelInfo{Name: mState.Name, Fields: mState.Fields, Meta: mState.Meta}
	oldField := fromState.Models[appLabel+"."+o.Model].GetField(o.Name)
	return schemaEditor.AlterColumn(info, oldField, o.Field)
}

func (o AlterField) DatabaseBackwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	mState := fromState.Models[appLabel+"."+o.Model]
	info := &orm.ModelInfo{Name: mState.Name, Fields: mState.Fields, Meta: mState.Meta}
	oldField := toState.Models[appLabel+"."+o.Model].GetField(o.Name)
	return schemaEditor.AlterColumn(info, o.Field, oldField) // o.Field is the new field coming backwards
}

type RunSQL struct {
	SQL        string
	ReverseSQL string
}

func (o RunSQL) StateForwards(appLabel string, state *ProjectState) {}

func (o RunSQL) DatabaseForwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	// In a real framework, RunSQL operates directly on a DB connection.
	// For now, assume SchemaEditor handles raw exec or we use the underlying connection.
	// We'd expose Execute on SchemaEditor or pass the backend.
	// For proto, ignore direct execution details or stub it.
	return nil
}

func (o RunSQL) DatabaseBackwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	if o.ReverseSQL == "" {
		return fmt.Errorf("no reverse SQL defined for RunSQL")
	}
	return nil
}

type RunPython struct {
	Forwards  func(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error
	Backwards func(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error
}

func (o RunPython) StateForwards(appLabel string, state *ProjectState) {}

func (o RunPython) DatabaseForwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	if o.Forwards != nil {
		return o.Forwards(appLabel, schemaEditor, fromState, toState)
	}
	return nil
}

func (o RunPython) DatabaseBackwards(appLabel string, schemaEditor backends.SchemaEditor, fromState, toState *ProjectState) error {
	if o.Backwards != nil {
		return o.Backwards(appLabel, schemaEditor, fromState, toState)
	}
	return fmt.Errorf("no reverse function defined for RunPython")
}
