package tags

import (
	"testing"
	"strings"

	godjango "github.com/pkshahid/JanGo/template"
)

func TestUtilTags(t *testing.T) {
	lib := godjango.NewLibrary()
	RegisterUtilTags(lib)

	engine := godjango.NewEngine(nil, false)
	engine.AddBuiltin(lib)

	// With
	inputWith := `{% with a=1 b="2" %}{{ a }}{{ b }}{% endwith %}`
	tmplWith, _ := engine.FromString(inputWith)
	resWith, _ := tmplWith.Render(nil)
	if resWith != "12" { t.Errorf("With tag mismatch: %s", resWith) }

	// Comment
	inputComment := `A{% comment %}hidden{% endcomment %}B`
	tmplComment, _ := engine.FromString(inputComment)
	resComment, _ := tmplComment.Render(nil)
	if resComment != "AB" { t.Errorf("Comment tag mismatch: %s", resComment) }

	// Cycle
	inputCycle := `{% cycle 'X' 'Y' %}{% cycle 'X' 'Y' %}{% cycle 'X' 'Y' %}`
	tmplCycle, _ := engine.FromString(inputCycle)
	resCycle, _ := tmplCycle.Render(godjango.NewContext(nil))
	_ = resCycle
	// Let's test inside a loop for state tracking:
	// A new parser parses this once, returning 1 CycleNode. The state updates per Render iteration.
	inputCycle2 := `{% for i in list %}{% cycle 'X' 'Y' %}{% endfor %}`
	// We need to register logic tags for `for` to work
	RegisterLogicTags(lib)
	// Parse with the full library
	tmplCycle3, err := engine.FromString(inputCycle2)
	if err != nil { t.Errorf("Parse error: %v", err) }
	resCycle2, err2 := tmplCycle3.Render(godjango.NewContext(map[string]any{"list": []int{1, 2, 3}}))
	if err2 != nil { t.Errorf("Render error: %v", err2) }
	if resCycle2 != "XYX" { t.Errorf("Cycle tag mismatch: %q", resCycle2) }

	// Now
	inputNow := `{% now "Y" %}`
	tmplNow, _ := engine.FromString(inputNow)
	resNow, _ := tmplNow.Render(nil)
	if len(resNow) != 4 || !strings.HasPrefix(resNow, "20") { t.Errorf("Now tag mismatch: %s", resNow) }

	// Spaceless
	inputSpaceless := `{% spaceless %} <p>
	<a href="#"> Link </a>
</p> {% endspaceless %}`
	tmplSpaceless, _ := engine.FromString(inputSpaceless)
	resSpaceless, _ := tmplSpaceless.Render(nil)
	expectedSpaceless := ` <p><a href="#"> Link </a></p> `
	if resSpaceless != expectedSpaceless { t.Errorf("Spaceless mismatch:\n%q\nExpected:\n%q", resSpaceless, expectedSpaceless) }

	// Verbatim
	inputVerbatim := `{% verbatim %}{{ dont_render }}{% endverbatim %}`
	tmplVerbatim, _ := engine.FromString(inputVerbatim)
	resVerbatim, _ := tmplVerbatim.Render(nil)
	if resVerbatim != "{{ dont_render }}" { t.Errorf("Verbatim mismatch: %s", resVerbatim) }

	// Lorem
	inputLorem := `{% lorem %}`
	tmplLorem, _ := engine.FromString(inputLorem)
	resLorem, _ := tmplLorem.Render(nil)
	if !strings.HasPrefix(resLorem, "Lorem ipsum") { t.Errorf("Lorem mismatch: %s", resLorem) }
}
