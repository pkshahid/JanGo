package signals_test

import (
	"github.com/pkshahid/JanGo/core/signals"
	"testing"
)

type mockModel struct {
	ID int
}

func TestSignalConnectAndSend(t *testing.T) {
	sig := signals.NewSignal("test_signal")

	called := false
	receiver := func(sender any, kwargs map[string]any) {
		called = true
		if val, ok := kwargs["key"].(string); !ok || val != "value" {
			t.Errorf("Expected kwargs['key'] == 'value'")
		}
	}

	err := sig.Connect(receiver, nil, false)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	responses := sig.Send(nil, map[string]any{"key": "value"})
	if !called {
		t.Errorf("Receiver was not called")
	}
	if len(responses) != 1 {
		t.Errorf("Expected 1 response, got %d", len(responses))
	}
}

func TestSignalSenderFiltering(t *testing.T) {
	sig := signals.NewSignal("model_signal")

	calledSpecific := false
	receiverSpecific := func(sender any, kwargs map[string]any) {
		calledSpecific = true
	}

	calledAll := false
	receiverAll := func(sender any, kwargs map[string]any) {
		calledAll = true
	}

	sig.Connect(receiverSpecific, mockModel{}, false)
	sig.Connect(receiverAll, nil, false)

	// Send with wrong sender type
	sig.Send(struct{ Name string }{}, nil)
	if calledSpecific {
		t.Errorf("Receiver specific should not have been called for struct{}")
	}
	if !calledAll {
		t.Errorf("Receiver all should have been called")
	}

	// Reset
	calledSpecific = false
	calledAll = false

	// Send with correct sender type
	sig.Send(mockModel{}, nil)
	if !calledSpecific {
		t.Errorf("Receiver specific should have been called for mockModel{}")
	}
	if !calledAll {
		t.Errorf("Receiver all should have been called")
	}
}

func TestSignalDisconnect(t *testing.T) {
	sig := signals.NewSignal("disconnect_test")
	called := false
	receiver := func(sender any, kwargs map[string]any) {
		called = true
	}

	sig.Connect(receiver, nil, false)
	err := sig.Disconnect(receiver, nil)
	if err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	sig.Send(nil, nil)
	if called {
		t.Errorf("Receiver should not be called after disconnect")
	}
}

func TestSignalSendRobust(t *testing.T) {
	sig := signals.NewSignal("robust_test")
	receiver := func(sender any, kwargs map[string]any) {
		panic("test panic")
	}

	receiverSafe := func(sender any, kwargs map[string]any) {
		// Should execute even if another panicked
	}

	sig.Connect(receiver, nil, false)
	sig.Connect(receiverSafe, nil, false)

	responses := sig.SendRobust(nil, nil)
	if len(responses) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(responses))
	}

	// One should have error
	hasError := false
	for _, r := range responses {
		if r.Err != nil && r.Err.Error() == "panic recovered: test panic" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("Expected to recover panic, but didn't find error in responses")
	}
}
