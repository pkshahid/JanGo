package forms

import (
	"fmt"
	"html"
	"strings"

	"github.com/pkshahid/JanGo/template"
)

// Widget interface defining the rendering of a form field.
type Widget interface {
	Render(name string, value any, attrs map[string]string) template.SafeString
	GetMedia() *Media
}

type BaseWidget struct {
	Attrs map[string]string
	Media *Media
}

func (w *BaseWidget) GetMedia() *Media {
	return w.Media
}

func buildAttrs(attrs map[string]string, defaults map[string]string) string {
	merged := make(map[string]string)
	for k, v := range defaults {
		merged[k] = v
	}
	for k, v := range attrs {
		merged[k] = v
	}

	var buf strings.Builder
	for k, v := range merged {
		buf.WriteString(fmt.Sprintf(` %s="%s"`, k, html.EscapeString(v)))
	}
	return buf.String()
}

// Input is a generic `<input type="...">` widget.
type Input struct {
	BaseWidget
	InputType string
}

func (w *Input) Render(name string, value any, attrs map[string]string) template.SafeString {
	mergedAttrs := buildAttrs(attrs, w.Attrs)
	valStr := ""
	if value != nil {
		valStr = fmt.Sprintf("%v", value)
	}
	return template.SafeString(fmt.Sprintf(`<input type="%s" name="%s" value="%s"%s>`, w.InputType, name, html.EscapeString(valStr), mergedAttrs))
}

// TextInput is `<input type="text">`.
func NewTextInput(attrs map[string]string) *Input {
	return &Input{BaseWidget: BaseWidget{Attrs: attrs}, InputType: "text"}
}

// NumberInput is `<input type="number">`.
func NewNumberInput(attrs map[string]string) *Input {
	return &Input{BaseWidget: BaseWidget{Attrs: attrs}, InputType: "number"}
}

// EmailInput is `<input type="email">`.
func NewEmailInput(attrs map[string]string) *Input {
	return &Input{BaseWidget: BaseWidget{Attrs: attrs}, InputType: "email"}
}

// URLInput is `<input type="url">`.
func NewURLInput(attrs map[string]string) *Input {
	return &Input{BaseWidget: BaseWidget{Attrs: attrs}, InputType: "url"}
}

// PasswordInput is `<input type="password">`.
func NewPasswordInput(attrs map[string]string) *Input {
	return &Input{BaseWidget: BaseWidget{Attrs: attrs}, InputType: "password"}
}

// HiddenInput is `<input type="hidden">`.
func NewHiddenInput(attrs map[string]string) *Input {
	return &Input{BaseWidget: BaseWidget{Attrs: attrs}, InputType: "hidden"}
}

// DateInput is `<input type="date">`.
func NewDateInput(attrs map[string]string) *Input {
	return &Input{BaseWidget: BaseWidget{Attrs: attrs}, InputType: "date"}
}

// TimeInput is `<input type="time">`.
func NewTimeInput(attrs map[string]string) *Input {
	return &Input{BaseWidget: BaseWidget{Attrs: attrs}, InputType: "time"}
}

// DateTimeInput is `<input type="datetime-local">`.
func NewDateTimeInput(attrs map[string]string) *Input {
	return &Input{BaseWidget: BaseWidget{Attrs: attrs}, InputType: "datetime-local"}
}

// FileInput is `<input type="file">`.
func NewFileInput(attrs map[string]string) *Input {
	return &Input{BaseWidget: BaseWidget{Attrs: attrs}, InputType: "file"}
}

// ClearableFileInput extends FileInput allowing clearing the selection (simplified).
func NewClearableFileInput(attrs map[string]string) *Input {
	// A full implementation renders a checkbox next to it to clear existing files.
	// We map it to FileInput for prototype.
	return NewFileInput(attrs)
}

// Textarea is `<textarea>...</textarea>`.
type Textarea struct {
	BaseWidget
}

func NewTextarea(attrs map[string]string) *Textarea {
	if attrs == nil {
		attrs = make(map[string]string)
	}
	if _, ok := attrs["cols"]; !ok {
		attrs["cols"] = "40"
	}
	if _, ok := attrs["rows"]; !ok {
		attrs["rows"] = "10"
	}
	return &Textarea{BaseWidget: BaseWidget{Attrs: attrs}}
}

func (w *Textarea) Render(name string, value any, attrs map[string]string) template.SafeString {
	mergedAttrs := buildAttrs(attrs, w.Attrs)
	valStr := ""
	if value != nil {
		valStr = fmt.Sprintf("%v", value)
	}
	return template.SafeString(fmt.Sprintf(`<textarea name="%s"%s>%s</textarea>`, name, mergedAttrs, html.EscapeString(valStr)))
}

// CheckboxInput is `<input type="checkbox">`.
type CheckboxInput struct {
	BaseWidget
}

func NewCheckboxInput(attrs map[string]string) *CheckboxInput {
	return &CheckboxInput{BaseWidget: BaseWidget{Attrs: attrs}}
}

func (w *CheckboxInput) Render(name string, value any, attrs map[string]string) template.SafeString {
	mergedAttrs := buildAttrs(attrs, w.Attrs)

	checked := ""
	// Simple boolean check
	if b, ok := value.(bool); ok && b {
		checked = " checked"
	} else if s, ok := value.(string); ok && (s == "true" || s == "on" || s == "1") {
		checked = " checked"
	} else if value != nil && fmt.Sprintf("%v", value) == "true" {
		checked = " checked"
	}

	return template.SafeString(fmt.Sprintf(`<input type="checkbox" name="%s"%s%s>`, name, mergedAttrs, checked))
}

// Select is `<select>...<option>...</select>`.
type Select struct {
	BaseWidget
	Choices []Choice
}

type Choice struct {
	Value string
	Label string
}

func NewSelect(choices []Choice, attrs map[string]string) *Select {
	return &Select{BaseWidget: BaseWidget{Attrs: attrs}, Choices: choices}
}

func (w *Select) Render(name string, value any, attrs map[string]string) template.SafeString {
	mergedAttrs := buildAttrs(attrs, w.Attrs)
	valStr := ""
	if value != nil {
		valStr = fmt.Sprintf("%v", value)
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf(`<select name="%s"%s>`+"\n", name, mergedAttrs))
	for _, c := range w.Choices {
		selected := ""
		if c.Value == valStr {
			selected = " selected"
		}
		buf.WriteString(fmt.Sprintf(`  <option value="%s"%s>%s</option>`+"\n", html.EscapeString(c.Value), selected, html.EscapeString(c.Label)))
	}
	buf.WriteString(`</select>`)
	return template.SafeString(buf.String())
}

// NullBooleanSelect is a Select with Yes/No/Unknown.
func NewNullBooleanSelect(attrs map[string]string) *Select {
	choices := []Choice{
		{Value: "unknown", Label: "Unknown"},
		{Value: "true", Label: "Yes"},
		{Value: "false", Label: "No"},
	}
	return NewSelect(choices, attrs)
}

// SelectMultiple is `<select multiple>`.
type SelectMultiple struct {
	Select
}

func NewSelectMultiple(choices []Choice, attrs map[string]string) *SelectMultiple {
	if attrs == nil {
		attrs = make(map[string]string)
	}
	attrs["multiple"] = "multiple"
	return &SelectMultiple{Select: *NewSelect(choices, attrs)}
}

// CheckboxSelectMultiple renders a list of checkboxes.
type CheckboxSelectMultiple struct {
	BaseWidget
	Choices []Choice
}

func NewCheckboxSelectMultiple(choices []Choice, attrs map[string]string) *CheckboxSelectMultiple {
	return &CheckboxSelectMultiple{BaseWidget: BaseWidget{Attrs: attrs}, Choices: choices}
}

func (w *CheckboxSelectMultiple) Render(name string, value any, attrs map[string]string) template.SafeString {
	// A real implementation would parse 'value' as a slice/array to check which are selected.
	// For the prototype, we treat value as a single item or assume all unselected.

	// Fast naive check if value is in a string repr list.
	valStr := fmt.Sprintf("%v", value)

	var buf strings.Builder
	buf.WriteString("<ul>\n")
	for i, c := range w.Choices {
		idAttr := fmt.Sprintf(`id="id_%s_%d"`, name, i)
		checked := ""
		if strings.Contains(valStr, c.Value) { // Simplified
			checked = " checked"
		}

		buf.WriteString(fmt.Sprintf(`  <li><label for="id_%s_%d"><input type="checkbox" name="%s" value="%s" %s%s> %s</label></li>`+"\n",
			name, i, name, html.EscapeString(c.Value), idAttr, checked, html.EscapeString(c.Label)))
	}
	buf.WriteString("</ul>")
	return template.SafeString(buf.String())
}

// RadioSelect renders a list of radio buttons.
type RadioSelect struct {
	CheckboxSelectMultiple // Logic is very similar
}

func NewRadioSelect(choices []Choice, attrs map[string]string) *RadioSelect {
	return &RadioSelect{CheckboxSelectMultiple: *NewCheckboxSelectMultiple(choices, attrs)}
}

func (w *RadioSelect) Render(name string, value any, attrs map[string]string) template.SafeString {
	valStr := ""
	if value != nil {
		valStr = fmt.Sprintf("%v", value)
	}

	var buf strings.Builder
	buf.WriteString("<ul>\n")
	for i, c := range w.Choices {
		idAttr := fmt.Sprintf(`id="id_%s_%d"`, name, i)
		checked := ""
		if c.Value == valStr {
			checked = " checked"
		}

		buf.WriteString(fmt.Sprintf(`  <li><label for="id_%s_%d"><input type="radio" name="%s" value="%s" %s%s> %s</label></li>`+"\n",
			name, i, name, html.EscapeString(c.Value), idAttr, checked, html.EscapeString(c.Label)))
	}
	buf.WriteString("</ul>")
	return template.SafeString(buf.String())
}

// SplitDateTimeWidget renders date and time separately.
type SplitDateTimeWidget struct {
	BaseWidget
	DateWidget Widget
	TimeWidget Widget
}

func NewSplitDateTimeWidget(attrs map[string]string) *SplitDateTimeWidget {
	return &SplitDateTimeWidget{
		BaseWidget: BaseWidget{Attrs: attrs},
		DateWidget: NewDateInput(attrs),
		TimeWidget: NewTimeInput(attrs),
	}
}

func (w *SplitDateTimeWidget) Render(name string, value any, attrs map[string]string) template.SafeString {
	// A real implementation splits value (time.Time) into string formats.
	// For proto, pass empty to both.

	dStr := w.DateWidget.Render(name+"_0", "", attrs)
	tStr := w.TimeWidget.Render(name+"_1", "", attrs)
	return template.SafeString(string(dStr) + " " + string(tStr))
}
