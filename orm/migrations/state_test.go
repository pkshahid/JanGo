package migrations

import (
	"reflect"
	"testing"

	"github.com/pkshahid/JanGo/orm"
)

type TestAppModel struct {
	orm.Model
	Name string `gd:"CharField"`
}

func TestProjectState(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&TestAppModel{})

	state := ProjectStateFromApps()

	modelKey := "migrations.testappmodel"
	if _, ok := state.Models[modelKey]; !ok {
		t.Errorf("Expected model %s in state", modelKey)
	}

	mState := state.Models[modelKey]
	if mState.AppName != "migrations" {
		t.Errorf("Expected AppName migrations, got %s", mState.AppName)
	}

	clone := state.Clone()
	if !reflect.DeepEqual(clone.Models[modelKey], mState) {
		t.Errorf("Clone mismatch")
	}

	mState.RemoveField("Name")
	if mState.GetField("Name") != nil {
		t.Errorf("Field not removed")
	}

	if clone.Models[modelKey].GetField("Name") == nil {
		t.Errorf("Clone was modified by original's mutation")
	}
}
