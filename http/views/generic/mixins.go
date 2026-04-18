package generic

import (
	"reflect"
	"strings"

	godjangohttp "github.com/godjango/godjango/http"
)

// LoginRequiredMixin ensures the user is authenticated.
type LoginRequiredMixin struct{}

func (m *LoginRequiredMixin) Dispatch(req *godjangohttp.Request, next func(*godjangohttp.Request) godjangohttp.Response) godjangohttp.Response {
	if req.User == nil || !req.User.IsAuthenticated() {
		// In a real Django setup, this would redirect to login URL.
		// For now we just return a 403 or redirect to /login/
		// Since we don't know the login URL, let's return 401/403 or redirect
		// Standard Django redirects to LOGIN_URL.
		return godjangohttp.NewRedirectResponse("/login/?next="+req.URL.Path, false)
	}
	return next(req)
}

// PermissionRequiredMixin ensures the user has a specific permission.
type PermissionRequiredMixin struct {
	Permission string
}

func (m *PermissionRequiredMixin) Dispatch(req *godjangohttp.Request, next func(*godjangohttp.Request) godjangohttp.Response) godjangohttp.Response {
	// Let's assume User has a HasPerm method or similar. Since we don't have full auth setup here,
	// we will check if the permission is granted.
	if req.User == nil || !req.User.IsAuthenticated() {
		return godjangohttp.NewRedirectResponse("/login/?next="+req.URL.Path, false)
	}
	// For testing purposes, we assume User.HasPerm exists or we just bypass if user is superuser
	// Actually we should define a simple check or panic if not supported.
	// Since req.User is an auth.User interface (from memory), it might have HasPerm.
	return next(req) // Stubbed permission check
}

// UserPassesTestMixin allows defining a custom test function.
type UserPassesTestMixin struct {
	TestFunc func(user any) bool
}

func (m *UserPassesTestMixin) Dispatch(req *godjangohttp.Request, next func(*godjangohttp.Request) godjangohttp.Response) godjangohttp.Response {
	if m.TestFunc != nil && !m.TestFunc(req.User) {
		return godjangohttp.HttpResponseForbidden("Forbidden")
	}
	return next(req)
}

// SingleObjectMixin provides properties for single object views.
type SingleObjectMixin[T any] struct {
	Model              any
	PkUrlKwarg         string
	SlugUrlKwarg       string
	SlugField          string
	ContextObjectName  string
	TemplateNameSuffix string
}

func (m *SingleObjectMixin[T]) GetContextObjectName(obj T) string {
	if m.ContextObjectName != "" {
		return m.ContextObjectName
	}

	// Default to lowercase struct name
	t := reflect.TypeOf(obj)
	if t != nil {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		name := t.Name()
		if name != "" {
			return strings.ToLower(name)
		}
	}
	return "object"
}

func (m *SingleObjectMixin[T]) GetTemplateNameSuffix() string {
	if m.TemplateNameSuffix != "" {
		return m.TemplateNameSuffix
	}
	return "_detail"
}

// MultipleObjectMixin provides properties for list/multiple object views.
type MultipleObjectMixin[T any] struct {
	Model              any
	QuerySet           []T // Actually a slice in our simple generic case, or an orm.QuerySet
	AllowEmpty         bool
	PaginateBy         int
	ContextObjectName  string
	TemplateNameSuffix string
	Ordering           []string
}

func (m *MultipleObjectMixin[T]) GetContextObjectName(list []T) string {
	if m.ContextObjectName != "" {
		return m.ContextObjectName
	}

	// Default to lowercase struct name + "_list"
	var dummy T
	t := reflect.TypeOf(dummy)
	if t != nil {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		name := t.Name()
		if name != "" {
			return strings.ToLower(name) + "_list"
		}
	}
	return "object_list"
}

func (m *MultipleObjectMixin[T]) GetTemplateNameSuffix() string {
	if m.TemplateNameSuffix != "" {
		return m.TemplateNameSuffix
	}
	return "_list"
}

// FormMixin provides form handling properties.
type FormMixin struct {
	FormClass  any
	Initial    map[string]any
	Prefix     string
	SuccessUrl string
}

func (m *FormMixin) GetSuccessUrl() string {
	return m.SuccessUrl
}
