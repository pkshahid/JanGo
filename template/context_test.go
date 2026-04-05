package template

import (
	"testing"
)

func TestContext(t *testing.T) {
	ctx := NewContext(map[string]any{"a": 1, "b": 2})

	if val, ok := ctx.Get("a"); !ok || val != 1 {
		t.Errorf("Expected 1, got %v", val)
	}

	ctx.Push(map[string]any{"b": 3, "c": 4})

	if val, ok := ctx.Get("b"); !ok || val != 3 {
		t.Errorf("Expected 3, got %v", val)
	}

	ctx.Set("d", 5)
	if val, ok := ctx.Get("d"); !ok || val != 5 {
		t.Errorf("Expected 5, got %v", val)
	}

	flat := ctx.Flatten()
	if flat["a"] != 1 || flat["b"] != 3 || flat["c"] != 4 || flat["d"] != 5 {
		t.Errorf("Flatten failed: %v", flat)
	}

	ctx.Pop()

	if val, ok := ctx.Get("b"); !ok || val != 2 {
		t.Errorf("Expected 2, got %v", val)
	}
	if _, ok := ctx.Get("c"); ok {
		t.Errorf("Expected c to be removed")
	}
}
