package forms

import (
	"strings"
	"testing"
)

func TestWidgets(t *testing.T) {
	// TextInput
	w1 := NewTextInput(map[string]string{"class": "form-control"})
	res1 := string(w1.Render("username", "admin", nil))
	if !strings.Contains(res1, `type="text"`) || !strings.Contains(res1, `name="username"`) || !strings.Contains(res1, `value="admin"`) || !strings.Contains(res1, `class="form-control"`) {
		t.Errorf("TextInput render mismatch: %s", res1)
	}

	// Textarea
	w2 := NewTextarea(nil)
	res2 := string(w2.Render("bio", "I'm a user.", nil))
	if !strings.Contains(res2, `<textarea name="bio"`) || !strings.Contains(res2, `I&#39;m a user.`) {
		t.Errorf("Textarea render mismatch: %s", res2)
	}

	// Select
	choices := []Choice{{Value: "1", Label: "One"}, {Value: "2", Label: "Two"}}
	w3 := NewSelect(choices, nil)
	res3 := string(w3.Render("choice", "2", nil))
	if !strings.Contains(res3, `<select name="choice"`) || !strings.Contains(res3, `value="2" selected>Two`) {
		t.Errorf("Select render mismatch: %s", res3)
	}

	// CheckboxInput
	w4 := NewCheckboxInput(nil)
	res4 := string(w4.Render("agree", true, nil))
	if !strings.Contains(res4, `type="checkbox" name="agree" checked`) {
		t.Errorf("CheckboxInput true mismatch: %s", res4)
	}

	res4b := string(w4.Render("agree", false, nil))
	if strings.Contains(res4b, `checked`) {
		t.Errorf("CheckboxInput false mismatch: %s", res4b)
	}

	// RadioSelect
	w5 := NewRadioSelect(choices, nil)
	res5 := string(w5.Render("radio_choice", "1", nil))
	if !strings.Contains(res5, `type="radio" name="radio_choice" value="1" id="id_radio_choice_0" checked> One`) {
		t.Errorf("RadioSelect render mismatch: %s", res5)
	}
}

func TestMediaMerge(t *testing.T) {
	m1 := NewMedia()
	m1.CSS["all"] = []string{"style1.css"}
	m1.JS = []string{"script1.js"}

	m2 := NewMedia()
	m2.CSS["all"] = []string{"style1.css", "style2.css"} // Deduplication expected
	m2.JS = []string{"script2.js", "script1.js"}

	m1.Merge(m2)

	if len(m1.CSS["all"]) != 2 || m1.CSS["all"][0] != "style1.css" || m1.CSS["all"][1] != "style2.css" {
		t.Errorf("CSS merge failed: %v", m1.CSS["all"])
	}

	if len(m1.JS) != 2 || m1.JS[0] != "script1.js" || m1.JS[1] != "script2.js" {
		t.Errorf("JS merge failed: %v", m1.JS)
	}

	cssHTML := string(m1.RenderCSS())
	if !strings.Contains(cssHTML, `href="style2.css"`) {
		t.Errorf("CSS render failed: %s", cssHTML)
	}

	jsHTML := string(m1.RenderJS())
	if !strings.Contains(jsHTML, `src="script2.js"`) {
		t.Errorf("JS render failed: %s", jsHTML)
	}
}
