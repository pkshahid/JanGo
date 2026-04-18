package signals_test

import (
	ormsignals "github.com/pkshahid/JanGo/orm/signals"
	"testing"
)

type MyModel struct {
	ID int
}

func TestORMSignals(t *testing.T) {
	called := false
	receiver := func(sender any, kwargs map[string]any) {
		called = true
		if _, ok := sender.(MyModel); !ok {
			t.Errorf("Expected sender to be MyModel, got %T", sender)
		}
		if val, ok := kwargs["created"].(bool); !ok || !val {
			t.Errorf("Expected kwargs['created'] == true")
		}
	}

	err := ormsignals.Connect(ormsignals.PostSave, receiver, MyModel{})
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	modelInstance := MyModel{ID: 1}

	// Simulate ORM trigger
	ormsignals.PostSave.Send(modelInstance, map[string]any{"instance": modelInstance, "created": true})

	if !called {
		t.Errorf("Receiver was not called for ORM PostSave signal")
	}

	ormsignals.PostSave.Disconnect(receiver, MyModel{})
}
