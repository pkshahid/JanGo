package forms

import (
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/pkshahid/JanGo/template"
)

// FormErrors maps field names to lists of error messages.
type FormErrors map[string][]string

// Form represents a collection of fields and their bound data.
type Form struct {
	Fields      map[string]Field
	Data        map[string]any
	Files       map[string][]*multipart.FileHeader
	CleanedData map[string]any
	Errors      FormErrors

	// Ordered fields for rendering
	Order []string

	IsBound bool

	// CleanFunc is an optional cross-field validation callback.
	// Set this to implement form-level validation (equivalent to Django's clean()).
	CleanFunc func() error
}

// NewForm initializes an empty form.
func NewForm(fields map[string]Field, order []string) *Form {
	if order == nil {
		for k := range fields {
			order = append(order, k)
		}
	}
	return &Form{
		Fields:  fields,
		Order:   order,
		Data:    make(map[string]any),
		Files:   make(map[string][]*multipart.FileHeader),
		IsBound: false,
	}
}

// Bind assigns data and files to the form for validation.
func (f *Form) Bind(data map[string]any, files map[string][]*multipart.FileHeader) {
	f.Data = data
	f.Files = files
	f.IsBound = true
	f.CleanedData = nil
	f.Errors = nil
}

// IsValid runs all field validation and form-level validation.
func (f *Form) IsValid() bool {
	if !f.IsBound {
		return false
	}
	if f.Errors != nil {
		return len(f.Errors) == 0
	}

	f.Errors = make(FormErrors)
	f.CleanedData = make(map[string]any)

	// Field-level validation
	for name, field := range f.Fields {
		var val any
		// Handle files separately
		if _, isFile := field.(*FileField); isFile {
			if f.Files != nil && len(f.Files[name]) > 0 {
				val = f.Files[name][0] // For prototype just take the first
			}
		} else if _, isImg := field.(*ImageField); isImg {
			if f.Files != nil && len(f.Files[name]) > 0 {
				val = f.Files[name][0] // For prototype just take the first
			}
		} else {
			if f.Data != nil {
				val = f.Data[name]
			}
		}

		cleaned, err := field.Clean(val)
		if err != nil {
			f.addError(name, err.Error())
		} else {
			f.CleanedData[name] = cleaned
		}
	}

	// Cross-field validation override
	if len(f.Errors) == 0 {
		if f.CleanFunc != nil {
			if err := f.CleanFunc(); err != nil {
				f.addError("__all__", err.Error())
			}
		} else if err := f.Clean(); err != nil {
			f.addError("__all__", err.Error())
		}
	}

	return len(f.Errors) == 0
}

// Clean can be overridden to implement cross-field validation.
func (f *Form) Clean() error {
	return nil
}

func (f *Form) addError(field string, errStr string) {
	f.Errors[field] = append(f.Errors[field], errStr)
}

// NonFieldErrors returns errors not associated with a specific field.
func (f *Form) NonFieldErrors() []string {
	if f.Errors == nil {
		return nil
	}
	return f.Errors["__all__"]
}

// Render represents the form as a table structure.
func (f *Form) RenderTable() template.SafeString {
	var buf strings.Builder

	// Non-field errors
	if errs := f.NonFieldErrors(); len(errs) > 0 {
		buf.WriteString(`<tr><td colspan="2"><ul class="errorlist nonfield">`)
		for _, e := range errs {
			buf.WriteString(fmt.Sprintf("<li>%s</li>", e))
		}
		buf.WriteString(`</ul></td></tr>` + "\n")
	}

	for _, name := range f.Order {
		field := f.Fields[name]
		var val any
		if f.IsBound {
			if f.CleanedData != nil {
				val = f.CleanedData[name]
			} else {
				val = f.Data[name]
			}
		}

		// Field errors
		if errs := f.Errors[name]; len(errs) > 0 {
			buf.WriteString(`<tr><td colspan="2"><ul class="errorlist">`)
			for _, e := range errs {
				buf.WriteString(fmt.Sprintf("<li>%s</li>", e))
			}
			buf.WriteString(`</ul></td></tr>` + "\n")
		}

		label := field.Label()
		if label == "" {
			label = strings.Title(name)
		}

		widgetStr := field.Widget().Render(name, val, nil)

		helpText := field.HelpText()
		if helpText != "" {
			helpText = fmt.Sprintf(`<br><span class="helptext">%s</span>`, helpText)
		}

		required := ""
		if field.Required() {
			required = " *"
		}

		buf.WriteString(fmt.Sprintf(`<tr><th><label for="id_%s">%s%s:</label></th><td>%s%s</td></tr>`+"\n",
			name, label, required, widgetStr, helpText))
	}

	return template.SafeString(buf.String())
}

// Render is a shortcut mapping to RenderTable.
func (f *Form) Render() template.SafeString {
	return f.RenderTable()
}

// Media returns all unique CSS and JS paths required by the form's widgets.
func (f *Form) Media() *Media {
	m := NewMedia()
	for _, field := range f.Fields {
		m.Merge(field.Widget().GetMedia())
	}
	return m
}
