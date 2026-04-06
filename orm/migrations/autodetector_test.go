package migrations

import (
	"testing"
	"github.com/godjango/godjango/orm"
)

func TestAutodetector(t *testing.T) {
	oldState := NewProjectState()

	oldModel := &ModelState{
		AppName: "app1",
		Name:    "OldModel",
		Fields: []*orm.Field{
			{Name: "ID", Type: orm.BigAutoField, PrimaryKey: true},
			{Name: "Title", Type: orm.CharField, Options: orm.FieldOptions{MaxLength: 100}},
		},
		Meta: &orm.Meta{DbTable: "app1_oldmodel"},
	}
	oldState.AddModel(oldModel)

	newState := NewProjectState()

	newModel := &ModelState{
		AppName: "app1",
		Name:    "OldModel",
		Fields: []*orm.Field{
			{Name: "ID", Type: orm.BigAutoField, PrimaryKey: true},
			// Title max length changed
			{Name: "Title", Type: orm.CharField, Options: orm.FieldOptions{MaxLength: 200}},
			// New field
			{Name: "Body", Type: orm.TextField},
		},
		Meta: &orm.Meta{DbTable: "app1_oldmodel"},
	}

	addedModel := &ModelState{
		AppName: "app2",
		Name:    "NewModel",
		Fields: []*orm.Field{
			{Name: "ID", Type: orm.BigAutoField, PrimaryKey: true},
			{Name: "Author", Type: orm.ForeignKey, Options: orm.FieldOptions{To: "app1.OldModel"}},
		},
		Meta: &orm.Meta{DbTable: "app2_newmodel"},
	}

	newState.AddModel(newModel)
	newState.AddModel(addedModel)

	autodetector := NewAutodetector(oldState, newState)
	changes := autodetector.Changes()

	if len(changes) != 2 {
		t.Fatalf("Expected 2 apps with changes, got %d", len(changes))
	}

	// Verify app1 changes
	app1Ops := changes["app1"]
	if len(app1Ops) != 2 {
		t.Fatalf("Expected 2 operations for app1, got %d", len(app1Ops))
	}

	hasAlter := false
	hasAdd := false
	for _, op := range app1Ops {
		if alter, ok := op.(AlterField); ok {
			if alter.Name == "Title" && alter.Field.Options.MaxLength == 200 {
				hasAlter = true
			}
		} else if add, ok := op.(AddField); ok {
			if add.Name == "Body" {
				hasAdd = true
			}
		}
	}

	if !hasAlter { t.Errorf("Expected AlterField for Title") }
	if !hasAdd { t.Errorf("Expected AddField for Body") }

	// Verify app2 changes
	app2Ops := changes["app2"]
	if len(app2Ops) != 1 {
		t.Fatalf("Expected 1 operation for app2")
	}

	createOp, ok := app2Ops[0].(CreateModel)
	if !ok || createOp.Name != "NewModel" {
		t.Errorf("Expected CreateModel for NewModel")
	}
}
