package forms

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"net/http"
	"mime/multipart"
	"io"
	"image"
	_ "image/jpeg"
	_ "image/png"
	_ "image/gif"
	"time"
)

// Field interface handles rendering and validation of a form field.
type Field interface {
	Clean(value any) (any, error)
	Widget() Widget
	Required() bool
	Label() string
	HelpText() string
	SetWidget(w Widget)
}

type BaseField struct {
	WidgetField Widget
	IsRequired  bool
	LabelStr    string
	HelpTextStr string
}

func (f *BaseField) Widget() Widget {
	if f.WidgetField == nil {
		f.WidgetField = NewTextInput(nil)
	}
	return f.WidgetField
}
func (f *BaseField) SetWidget(w Widget) { f.WidgetField = w }
func (f *BaseField) Required() bool     { return f.IsRequired }
func (f *BaseField) Label() string      { return f.LabelStr }
func (f *BaseField) HelpText() string   { return f.HelpTextStr }

// CharField
type CharField struct {
	BaseField
	MaxLength int
	MinLength int
	Strip     bool
}

func (f *CharField) Clean(value any) (any, error) {
	str := ""
	if value != nil {
		str = fmt.Sprintf("%v", value)
	}
	if f.Strip {
		str = strings.TrimSpace(str)
	}

	if str == "" {
		if f.IsRequired {
			return nil, fmt.Errorf("this field is required")
		}
		return "", nil
	}

	runes := []rune(str)
	if f.MinLength > 0 && len(runes) < f.MinLength {
		return nil, fmt.Errorf("ensure this value has at least %d characters (it has %d)", f.MinLength, len(runes))
	}
	if f.MaxLength > 0 && len(runes) > f.MaxLength {
		return nil, fmt.Errorf("ensure this value has at most %d characters (it has %d)", f.MaxLength, len(runes))
	}

	return str, nil
}

// IntegerField
type IntegerField struct {
	BaseField
}

func (f *IntegerField) Clean(value any) (any, error) {
	if value == nil || value == "" {
		if f.IsRequired {
			return nil, fmt.Errorf("this field is required")
		}
		return nil, nil
	}
	valStr := fmt.Sprintf("%v", value)
	num, err := strconv.Atoi(strings.TrimSpace(valStr))
	if err != nil {
		return nil, fmt.Errorf("enter a valid integer")
	}
	return num, nil
}

// FloatField
type FloatField struct {
	BaseField
}

func (f *FloatField) Clean(value any) (any, error) {
	if value == nil || value == "" {
		if f.IsRequired {
			return nil, fmt.Errorf("this field is required")
		}
		return nil, nil
	}
	valStr := fmt.Sprintf("%v", value)
	num, err := strconv.ParseFloat(strings.TrimSpace(valStr), 64)
	if err != nil {
		return nil, fmt.Errorf("enter a valid number")
	}
	return num, nil
}

// BooleanField
type BooleanField struct {
	BaseField
}

func (f *BooleanField) Clean(value any) (any, error) {
	if b, ok := value.(bool); ok {
		return b, nil
	}
	str := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", value)))
	val := str == "true" || str == "on" || str == "1"
	if f.IsRequired && !val {
		return false, fmt.Errorf("this field is required")
	}
	return val, nil
}

// EmailField
type EmailField struct {
	CharField
}

func (f *EmailField) Clean(value any) (any, error) {
	cleaned, err := f.CharField.Clean(value)
	if err != nil || cleaned == "" {
		return cleaned, err
	}
	str := cleaned.(string)
	// Very naive email regex
	if !strings.Contains(str, "@") {
		return nil, fmt.Errorf("enter a valid email address")
	}
	return str, nil
}

// URLField
type URLField struct {
	CharField
}

func (f *URLField) Clean(value any) (any, error) {
	cleaned, err := f.CharField.Clean(value)
	if err != nil || cleaned == "" {
		return cleaned, err
	}
	str := cleaned.(string)
	if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
		str = "http://" + str
	}
	return str, nil
}

// SlugField
type SlugField struct {
	CharField
}

func (f *SlugField) Clean(value any) (any, error) {
	cleaned, err := f.CharField.Clean(value)
	if err != nil || cleaned == "" {
		return cleaned, err
	}
	str := cleaned.(string)
	matched, _ := regexp.MatchString(`^[-a-zA-Z0-9_]+$`, str)
	if !matched {
		return nil, fmt.Errorf("enter a valid slug consisting of letters, numbers, underscores or hyphens")
	}
	return str, nil
}

// DateField (Simplified parsing)
type DateField struct {
	BaseField
}

func (f *DateField) Clean(value any) (any, error) {
	if value == nil || value == "" {
		if f.IsRequired {
			return nil, fmt.Errorf("this field is required")
		}
		return nil, nil
	}
	str := strings.TrimSpace(fmt.Sprintf("%v", value))
	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return nil, fmt.Errorf("enter a valid date")
	}
	return t, nil
}

// ChoiceField
type ChoiceField struct {
	BaseField
	Choices []Choice
}

func (f *ChoiceField) Clean(value any) (any, error) {
	if value == nil || value == "" {
		if f.IsRequired {
			return nil, fmt.Errorf("this field is required")
		}
		return "", nil
	}
	str := fmt.Sprintf("%v", value)
	for _, c := range f.Choices {
		if c.Value == str {
			return str, nil
		}
	}
	return nil, fmt.Errorf("select a valid choice. %s is not one of the available choices", str)
}

// MultipleChoiceField
type MultipleChoiceField struct {
	BaseField
	Choices []Choice
}

func (f *MultipleChoiceField) Clean(value any) (any, error) {
	// A real implementation expects `[]string` from `req.POST["field"]`.
	// For proto, we'll try to convert or stringify.
	var vals []string
	if slice, ok := value.([]string); ok {
		vals = slice
	} else if value != nil && value != "" {
		vals = []string{fmt.Sprintf("%v", value)}
	}

	if len(vals) == 0 {
		if f.IsRequired {
			return nil, fmt.Errorf("this field is required")
		}
		return []string{}, nil
	}

	var cleaned []string
	validMap := make(map[string]bool)
	for _, c := range f.Choices {
		validMap[c.Value] = true
	}

	for _, v := range vals {
		if !validMap[v] {
			return nil, fmt.Errorf("select a valid choice. %s is not one of the available choices", v)
		}
		cleaned = append(cleaned, v)
	}

	return cleaned, nil
}

// FileField and ImageField stubs
type FileField struct {
	BaseField
	MaxBytes     int64
	AllowedTypes []string
}

func (f *FileField) Clean(value any) (any, error) {
	if value == nil {
		if f.IsRequired {
			return nil, fmt.Errorf("this field is required")
		}
		return nil, nil
	}

	// Value should be *multipart.FileHeader
	header, ok := value.(*multipart.FileHeader)
	if !ok {
		return nil, fmt.Errorf("invalid file upload format")
	}

	if f.MaxBytes > 0 && header.Size > f.MaxBytes {
		return nil, fmt.Errorf("file size exceeds maximum allowed")
	}

	if len(f.AllowedTypes) > 0 {
		file, err := header.Open()
		if err != nil {
			return nil, fmt.Errorf("could not read file")
		}
		defer file.Close()

		buf := make([]byte, 512)
		if _, err := file.Read(buf); err != nil && err != io.EOF {
			return nil, fmt.Errorf("could not read file headers")
		}

		contentType := http.DetectContentType(buf)
		allowed := false
		for _, t := range f.AllowedTypes {
			if strings.HasPrefix(contentType, t) {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("file type %s is not allowed", contentType)
		}
	}

	return header, nil
}

type ImageField struct {
	FileField
}

func (f *ImageField) Clean(value any) (any, error) {
	cleaned, err := f.FileField.Clean(value)
	if err != nil || cleaned == nil {
		return cleaned, err
	}

	header := cleaned.(*multipart.FileHeader)
	file, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("could not read file")
	}
	defer file.Close()

	_, _, err = image.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("uploaded file is not a valid image")
	}

	return header, nil
}
