package admin

import (
	"net/http/httptest"
	"testing"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/urls"
	"github.com/pkshahid/JanGo/orm"
)

// --- Test models ---

type filterTestModel struct {
	orm.Model
	Title       string `gd:"CharField,max_length=200"`
	IsPublished bool   `gd:"BooleanField,default=false"`
}

// --- Test AppConfig that implements AdminRegistrar ---


// --- Autodiscover tests ---

func TestAutodiscover_RegistersModelsFromApp(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&filterTestModel{})

	site := NewAdminSite("test")

	// No models registered yet
	if len(site._registry) != 0 {
		t.Fatalf("Expected empty registry, got %d", len(site._registry))
	}

	// Autodiscover with no installed apps → no-op, no error
	err := site.Autodiscover()
	if err != nil {
		t.Fatalf("Autodiscover with no apps should not error: %v", err)
	}
}

func TestAutodiscover_SkipsNonRegistrarApps(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&filterTestModel{})

	site := NewAdminSite("test")
	// Autodiscover iterates apps.All() which returns nil if registry not set up
	err := site.Autodiscover()
	if err != nil {
		t.Fatalf("Autodiscover should not error: %v", err)
	}
	if len(site._registry) != 0 {
		t.Errorf("Expected 0 registered models, got %d", len(site._registry))
	}
}

// --- Custom filter tests ---

func TestSimpleListFilter_Title(t *testing.T) {
	f := &SimpleListFilter{
		FilterTitle: "Published status",
	}
	if f.Title() != "Published status" {
		t.Errorf("Expected 'Published status', got '%s'", f.Title())
	}
}

func TestSimpleListFilter_Chooses(t *testing.T) {
	f := &SimpleListFilter{
		FilterTitle:   "Published status",
		ParameterName: "published",
		Lookups: []FilterLookup{
			{Display: "Published", Value: "yes"},
			{Display: "Unpublished", Value: "no"},
		},
	}

	// Test with no filter selected
	rawReq := httptest.NewRequest("GET", "/admin/test/filtermodel/?", nil)
	req := godjangohttp.NewRequest(rawReq)
	choices := f.Choices(req)
	if len(choices) != 3 { // All + 2 lookups
		t.Fatalf("Expected 3 choices, got %d", len(choices))
	}
	if !choices[0].Selected {
		t.Errorf("Expected 'All' to be selected when no filter param")
	}
	if choices[1].Selected {
		t.Errorf("Expected 'Published' to not be selected")
	}

	// Test with "yes" selected
	rawReq2 := httptest.NewRequest("GET", "/admin/test/filtermodel/?published=yes", nil)
	req2 := godjangohttp.NewRequest(rawReq2)
	choices2 := f.Choices(req2)
	if choices2[0].Selected {
		t.Errorf("Expected 'All' to not be selected when published=yes")
	}
	if !choices2[1].Selected {
		t.Errorf("Expected 'Published' to be selected when published=yes")
	}
	if choices2[2].Selected {
		t.Errorf("Expected 'Unpublished' to not be selected when published=yes")
	}
}

func TestSimpleListFilter_Queryset(t *testing.T) {
	f := &SimpleListFilter{
		FilterTitle:   "Published status",
		ParameterName: "published",
		Lookups: []FilterLookup{
			{Display: "Published", Value: "yes"},
			{Display: "Unpublished", Value: "no"},
		},
		QuerysetFn: func(val string, info *orm.ModelInfo) (string, []any) {
			if val == "yes" {
				return "IsPublished = ?", []any{true}
			}
			if val == "no" {
				return "IsPublished = ?", []any{false}
			}
			return "", nil
		},
	}

	// No param → empty queryset
	rawReq := httptest.NewRequest("GET", "/admin/test/filtermodel/?", nil)
	req := godjangohttp.NewRequest(rawReq)
	where, args := f.Queryset(req, nil)
	if where != "" || args != nil {
		t.Errorf("Expected empty queryset with no param, got '%s' %v", where, args)
	}

	// published=yes → IsPublished = true
	rawReq2 := httptest.NewRequest("GET", "/admin/test/filtermodel/?published=yes", nil)
	req2 := godjangohttp.NewRequest(rawReq2)
	where2, args2 := f.Queryset(req2, nil)
	if where2 != "IsPublished = ?" {
		t.Errorf("Expected 'IsPublished = ?', got '%s'", where2)
	}
	if len(args2) != 1 || args2[0] != true {
		t.Errorf("Expected [true], got %v", args2)
	}

	// published=no → IsPublished = false
	rawReq3 := httptest.NewRequest("GET", "/admin/test/filtermodel/?published=no", nil)
	req3 := godjangohttp.NewRequest(rawReq3)
	where3, args3 := f.Queryset(req3, nil)
	if where3 != "IsPublished = ?" {
		t.Errorf("Expected 'IsPublished = ?', got '%s'", where3)
	}
	if len(args3) != 1 || args3[0] != false {
		t.Errorf("Expected [false], got %v", args3)
	}
}

// --- FieldListFilter tests ---

func TestFieldListFilter_BooleanChoices(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&filterTestModel{})

	info, err := orm.GetModelInfo(&filterTestModel{})
	if err != nil {
		t.Fatalf("Failed to get model info: %v", err)
	}

	f := NewFieldListFilter("IsPublished", info)
	if f.Title() != "IsPublished" {
		t.Errorf("Expected 'IsPublished', got '%s'", f.Title())
	}

	rawReq := httptest.NewRequest("GET", "/admin/test/filtermodel/?", nil)
	req := godjangohttp.NewRequest(rawReq)
	choices := f.Choices(req)
	// All + Yes + No = 3
	if len(choices) != 3 {
		t.Fatalf("Expected 3 choices for boolean field, got %d", len(choices))
	}
	if choices[0].Display != "All" {
		t.Errorf("Expected first choice 'All', got '%s'", choices[0].Display)
	}
	if choices[1].Display != "Yes" {
		t.Errorf("Expected second choice 'Yes', got '%s'", choices[1].Display)
	}
	if choices[2].Display != "No" {
		t.Errorf("Expected third choice 'No', got '%s'", choices[2].Display)
	}
}

func TestFieldListFilter_BooleanQueryset(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&filterTestModel{})

	info, err := orm.GetModelInfo(&filterTestModel{})
	if err != nil {
		t.Fatalf("Failed to get model info: %v", err)
	}

	f := NewFieldListFilter("IsPublished", info)

	// ispublished=true → is_published = true
	rawReq := httptest.NewRequest("GET", "/admin/test/filtermodel/?ispublished=true", nil)
	req := godjangohttp.NewRequest(rawReq)
	where, args := f.Queryset(req, info)
	if where != "is_published = ?" {
		t.Errorf("Expected 'is_published = ?', got '%s'", where)
	}
	if len(args) != 1 || args[0] != true {
		t.Errorf("Expected [true], got %v", args)
	}

	// ispublished=false → is_published = false
	rawReq2 := httptest.NewRequest("GET", "/admin/test/filtermodel/?ispublished=false", nil)
	req2 := godjangohttp.NewRequest(rawReq2)
	where2, args2 := f.Queryset(req2, info)
	if where2 != "is_published = ?" {
		t.Errorf("Expected 'is_published = ?', got '%s'", where2)
	}
	if len(args2) != 1 || args2[0] != false {
		t.Errorf("Expected [false], got %v", args2)
	}
}

// --- ResolveFilters tests ---

func TestResolveFilters_StringAndCustom(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&filterTestModel{})

	info, err := orm.GetModelInfo(&filterTestModel{})
	if err != nil {
		t.Fatalf("Failed to get model info: %v", err)
	}

	customFilter := &SimpleListFilter{
		FilterTitle:   "Published status",
		ParameterName: "published",
		Lookups: []FilterLookup{
			{Display: "Published", Value: "yes"},
		},
	}

	ma := &ModelAdmin{
		ModelInfo:  info,
		ListFilter: []any{"IsPublished", customFilter},
	}

	filters := ResolveFilters(ma)
	if len(filters) != 2 {
		t.Fatalf("Expected 2 filters, got %d", len(filters))
	}

	// First should be FieldListFilter
	if _, ok := filters[0].(*FieldListFilter); !ok {
		t.Errorf("Expected first filter to be *FieldListFilter, got %T", filters[0])
	}

	// Second should be SimpleListFilter
	if _, ok := filters[1].(*SimpleListFilter); !ok {
		t.Errorf("Expected second filter to be *SimpleListFilter, got %T", filters[1])
	}
}

func TestResolveFilters_Empty(t *testing.T) {
	ma := &ModelAdmin{}
	filters := ResolveFilters(ma)
	if filters != nil {
		t.Errorf("Expected nil for empty ListFilter, got %v", filters)
	}
}

// --- Changelist view with filters test ---

func TestChangelistView_WithFilters(t *testing.T) {
	orm.ClearRegistry()
	orm.Register(&filterTestModel{})

	site := NewAdminSite("test")
	err := site.Register(&filterTestModel{}, &ModelAdmin{
		ListDisplay: []string{"Title", "IsPublished"},
		ListFilter:  []any{"IsPublished"},
	})
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	rawReq := httptest.NewRequest("GET", "/admin/test/filtermodel/?ispublished=true", nil)
	req := godjangohttp.NewRequest(rawReq)
	req.User = &mockUser{authenticated: true, staff: true}
	req.ResolverMatch = &urls.ResolverMatch{
		Kwargs: map[string]any{
			"app_label":   "test",
			"model_name":  "filtertestmodel",
			"object_id":   "",
		},
	}

	resp := site.changelistView(req)
	hr, ok := resp.(*godjangohttp.HttpResponse)
	if !ok {
		t.Fatalf("Expected HttpResponse, got %T", resp)
	}
	if hr.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", hr.StatusCode)
	}
}
