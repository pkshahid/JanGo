package tags

import (
	godjango "github.com/pkshahid/JanGo/template"
	"testing"
)

func TestLogicTags(t *testing.T) {
	lib := godjango.NewLibrary()
	RegisterLogicTags(lib)

	engine := godjango.NewEngine(nil, false)
	engine.AddBuiltin(lib)

	// Test If / Elif / Else
	inputIf := `{% if flag %}A{% elif other %}B{% else %}C{% endif %}`
	tmplIf, _ := engine.FromString(inputIf)

	resIf1, _ := tmplIf.Render(godjango.NewContext(map[string]any{"flag": true}))
	if resIf1 != "A" {
		t.Errorf("Expected A, got %s", resIf1)
	}

	resIf2, _ := tmplIf.Render(godjango.NewContext(map[string]any{"other": true}))
	if resIf2 != "B" {
		t.Errorf("Expected B, got %s", resIf2)
	}

	resIf3, _ := tmplIf.Render(godjango.NewContext(map[string]any{}))
	if resIf3 != "C" {
		t.Errorf("Expected C, got %s", resIf3)
	}

	// Test For / Empty / forloop vars
	inputFor := `{% for x in list %}{{ forloop.counter }}:{{ x }}{% empty %}EMPTY{% endfor %}`
	tmplFor, _ := engine.FromString(inputFor)

	resFor1, _ := tmplFor.Render(godjango.NewContext(map[string]any{"list": []any{"a", "b"}}))
	if resFor1 != "1:a2:b" {
		t.Errorf("Expected 1:a2:b, got %s", resFor1)
	}

	resFor2, _ := tmplFor.Render(godjango.NewContext(map[string]any{"list": []any{}}))
	if resFor2 != "EMPTY" {
		t.Errorf("Expected EMPTY, got %s", resFor2)
	}
}
