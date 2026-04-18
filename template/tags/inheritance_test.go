package tags

import (
	"os"
	"path/filepath"
	"testing"
	godjango "github.com/pkshahid/JanGo/template"
)

func TestInheritanceTags(t *testing.T) {
	lib := godjango.NewLibrary()
	RegisterInheritanceTags(lib)

	tmpDir, _ := os.MkdirTemp("", "templates")
	defer os.RemoveAll(tmpDir)

	baseHTML := `BASE: {% block content %}empty{% endblock %}`
	os.WriteFile(filepath.Join(tmpDir, "base.html"), []byte(baseHTML), 0644)

	partialHTML := `PARTIAL: {{ var }}`
	os.WriteFile(filepath.Join(tmpDir, "partial.html"), []byte(partialHTML), 0644)

	engine := godjango.NewEngine([]string{tmpDir}, false)
	engine.AddBuiltin(lib)

	// Extends and Block
	inputExt := `{% extends "base.html" %}{% block content %}filled|{{ block.super }}{% endblock %}`
	tmplExt, _ := engine.FromString(inputExt)

	resExt, err := tmplExt.Render(godjango.NewContext(nil))
	if err != nil {
		t.Fatalf("Extends error: %v", err)
	}
	if resExt != "BASE: filled|empty" {
		t.Errorf("Extends mismatch, got: %s", resExt)
	}

	// Include with kwargs
	inputInc := `{% include "partial.html" with var="test" %}`
	tmplInc, _ := engine.FromString(inputInc)

	resInc, err := tmplInc.Render(godjango.NewContext(nil))
	if err != nil {
		t.Fatalf("Include error: %v", err)
	}
	if resInc != "PARTIAL: test" {
		t.Errorf("Include mismatch, got: %s", resInc)
	}
}
