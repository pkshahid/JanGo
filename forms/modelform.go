package forms

import (
	"fmt"
	"reflect"

	"github.com/godjango/godjango/orm"
)

// ModelForm represents a form automatically generated from an ORM model.
type ModelForm struct {
	Form
	ModelType reflect.Type
	Instance  any
	Includes  []string
	Excludes  []string
}

// NewModelForm generates a new ModelForm based on the specified model struct instance.
func NewModelForm(modelInstance any, includes, excludes []string) (*ModelForm, error) {
	info, err := orm.GetModelInfo(modelInstance)
	if err != nil {
		return nil, fmt.Errorf("forms: failed to get model info: %w", err)
	}

	fields := make(map[string]Field)
	var order []string

	for _, f := range info.Fields {
		if f.PrimaryKey && f.Type == orm.BigAutoField {
			continue // Skip auto PK
		}

		// Check includes/excludes
		if len(includes) > 0 && !contains(includes, f.Name) {
			continue
		}
		if len(excludes) > 0 && contains(excludes, f.Name) {
			continue
		}

		formField := generateFormField(f)
		if formField != nil {
			fields[f.Name] = formField
			order = append(order, f.Name)
		}
	}

	form := NewForm(fields, order)

	mf := &ModelForm{
		Form:      *form,
		ModelType: info.Type,
		Instance:  modelInstance,
		Includes:  includes,
		Excludes:  excludes,
	}

	// Populate initial data if instance has values
	mf.populateInitial(info)

	return mf, nil
}

func generateFormField(f *orm.Field) Field {
	req := !f.Options.Blank && !f.Options.Null

	switch f.Type {
	case orm.CharField, orm.TextField:
		cf := &CharField{
			BaseField: BaseField{IsRequired: req, LabelStr: f.Name},
			MaxLength: f.Options.MaxLength,
		}
		if f.Type == orm.TextField {
			cf.SetWidget(NewTextarea(nil))
		}
		return cf
	case orm.IntegerField, orm.SmallIntegerField, orm.BigIntegerField:
		return &IntegerField{BaseField: BaseField{IsRequired: req, LabelStr: f.Name, Widget: NewNumberInput(nil)}}
	case orm.FloatField, orm.DecimalField:
		return &FloatField{BaseField: BaseField{IsRequired: req, LabelStr: f.Name, Widget: NewNumberInput(map[string]string{"step": "any"})}}
	case orm.BooleanField, orm.NullBooleanField:
		// NullBooleanField might use a select, but we default to Checkbox for prototype
		return &BooleanField{BaseField: BaseField{IsRequired: req, LabelStr: f.Name, Widget: NewCheckboxInput(nil)}}
	case orm.EmailField:
		return &EmailField{CharField{BaseField: BaseField{IsRequired: req, LabelStr: f.Name, Widget: NewEmailInput(nil)}}}
	case orm.URLField:
		return &URLField{CharField{BaseField: BaseField{IsRequired: req, LabelStr: f.Name, Widget: NewURLInput(nil)}}}
	case orm.SlugField:
		return &SlugField{CharField{BaseField: BaseField{IsRequired: req, LabelStr: f.Name}}}
	case orm.DateField, orm.TimeField, orm.DateTimeField:
		widget := NewDateInput(nil) // Default
		return &DateField{BaseField: BaseField{IsRequired: req, LabelStr: f.Name, Widget: widget}}
	case orm.FileField, orm.ImageField:
		return &FileField{BaseField: BaseField{IsRequired: req, LabelStr: f.Name, Widget: NewFileInput(nil)}}
	case orm.ForeignKey, orm.OneToOneField, orm.ManyToManyField:
		// In a real framework, this would query the target model to generate choices.
		// For prototype, we generate an empty ChoiceField.
		return &ChoiceField{BaseField: BaseField{IsRequired: req, LabelStr: f.Name, Widget: NewSelect(nil, nil)}}
	}

	return nil // Skip unknown or unmappable
}

func (mf *ModelForm) populateInitial(info *orm.ModelInfo) {
	val := reflect.ValueOf(mf.Instance)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
	}

	initial := make(map[string]any)
	for name := range mf.Fields {
		f := val.FieldByName(name)
		if f.IsValid() {
			// Zero values check could be added, but for now we just push everything
			initial[name] = f.Interface()
		}
	}

	mf.Data = initial
	// Model forms are not "Bound" with initial data, so IsBound remains false.
}

// Save commits the CleanedData to the underlying Model instance.
// If commit=true, it would theoretically call the ORM to save to the database.
func (mf *ModelForm) Save(commit bool) (any, error) {
	if !mf.IsValid() {
		return nil, fmt.Errorf("cannot save invalid form")
	}

	val := reflect.ValueOf(mf.Instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for name, cleanedVal := range mf.CleanedData {
		f := val.FieldByName(name)
		if f.IsValid() && f.CanSet() && cleanedVal != nil {
			// Basic type conversion assignment. A real ORM needs robust type casting.
			cv := reflect.ValueOf(cleanedVal)
			if cv.Type().AssignableTo(f.Type()) {
				f.Set(cv)
			}
		}
	}

	// In a real framework:
	// if commit { orm.NewQuerySet(mf.ModelType).Save(mf.Instance) }

	return mf.Instance, nil
}
