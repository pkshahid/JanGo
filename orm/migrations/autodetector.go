package migrations

import (
	"reflect"
	"strings"

	"github.com/godjango/godjango/orm"
)

// MigrationAutodetector detects changes between two project states
// and generates a set of operations to transition from oldState to newState.
type MigrationAutodetector struct {
	FromState *ProjectState
	ToState   *ProjectState
}

// NewAutodetector initializes a new autodetector.
func NewAutodetector(fromState, toState *ProjectState) *MigrationAutodetector {
	return &MigrationAutodetector{
		FromState: fromState,
		ToState:   toState,
	}
}

// Changes calculates the operations required per app.
func (a *MigrationAutodetector) Changes() map[string][]Operation {
	changes := make(map[string][]Operation)

	// Track handled models to detect deletions
	handledModels := make(map[string]bool)

	// Detect Add/Alter
	for key, toModel := range a.ToState.Models {
		app := toModel.AppName
		fromModel, exists := a.FromState.Models[key]

		if !exists {
			// Model created
			op := CreateModel{
				Name:   toModel.Name,
				Fields: toModel.Fields,
				Meta:   toModel.Meta,
			}
			changes[app] = append(changes[app], op)
		} else {
			// Model exists, check for field changes
			handledFields := make(map[string]bool)

			// Detect added / altered fields
			for _, toField := range toModel.Fields {
				fromField := fromModel.GetField(toField.Name)
				if fromField == nil {
					// Field added
					op := AddField{
						Model: toModel.Name,
						Name:  toField.Name,
						Field: toField,
					}
					changes[app] = append(changes[app], op)
				} else {
					// Compare field signatures.
					// In a full implementation, this handles deep comparisons of FieldOptions, Type, etc.
					// For our prototype, we'll do a simple DeepEqual ignoring GoType specifics.
					if !fieldsEqual(toField, fromField) {
						op := AlterField{
							Model: toModel.Name,
							Name:  toField.Name,
							Field: toField,
						}
						changes[app] = append(changes[app], op)
					}
				}
				handledFields[toField.Name] = true
			}

			// Detect deleted fields
			for _, fromField := range fromModel.Fields {
				if !handledFields[fromField.Name] {
					op := RemoveField{
						Model: fromModel.Name,
						Name:  fromField.Name,
						Field: fromField,
					}
					changes[app] = append(changes[app], op)
				}
			}
		}
		handledModels[key] = true
	}

	// Detect deleted models
	for key, fromModel := range a.FromState.Models {
		if !handledModels[key] {
			app := fromModel.AppName
			op := DeleteModel{
				Name: fromModel.Name,
			}
			changes[app] = append(changes[app], op)
		}
	}

	return changes
}

func fieldsEqual(f1, f2 *orm.Field) bool {
	if f1.Type != f2.Type || f1.Column != f2.Column || f1.PrimaryKey != f2.PrimaryKey {
		return false
	}

	// Option checks
	return reflect.DeepEqual(f1.Options, f2.Options)
}

// ArrangeDependencies generates basic dependencies for new migrations
// based on ForeignKey references within the operations.
func ArrangeDependencies(changes map[string][]Operation, graph *MigrationGraph) map[string]*Migration {
	migrations := make(map[string]*Migration)

	// Determine latest migration for each app
	latest := make(map[string]string)
	for _, n := range graph.Nodes {
		// Just a mock heuristic. In reality, Django maintains a linked list per app.
		// For our prototype, we'll just say the last loaded is latest.
		// A full framework loads from db / graph leaves.
		if latest[n.Migration.App] < n.Migration.Name {
			latest[n.Migration.App] = n.Migration.Name
		}
	}

	for app, ops := range changes {
		if len(ops) == 0 {
			continue
		}

		deps := []Dep{}
		if last, ok := latest[app]; ok {
			deps = append(deps, Dep{App: app, Name: last})
		}

		// Detect cross-app dependencies via ForeignKeys
		for _, op := range ops {
			switch o := op.(type) {
			case CreateModel:
				for _, f := range o.Fields {
					if f.Type == orm.ForeignKey && f.Options.To != "" {
						targetParts := strings.Split(f.Options.To, ".")
						if len(targetParts) == 2 {
							targetApp := targetParts[0]
							if targetApp != app {
								if lastTarget, ok := latest[targetApp]; ok {
									deps = append(deps, Dep{App: targetApp, Name: lastTarget})
								}
							}
						}
					}
				}
			}
		}

		// Clean up duplicate deps
		depMap := make(map[string]Dep)
		for _, d := range deps {
			depMap[d.App+"."+d.Name] = d
		}

		finalDeps := []Dep{}
		for _, d := range depMap {
			finalDeps = append(finalDeps, d)
		}

		name := "0001_initial"
		if last, ok := latest[app]; ok {
			name = last + "_auto" // Naive naming
		}

		migrations[app] = &Migration{
			App:          app,
			Name:         name,
			Dependencies: finalDeps,
			Operations:   ops,
		}
	}

	return migrations
}
