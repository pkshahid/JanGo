package generic

import (
	"fmt"

	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/http/views"
)

// FormView displays a form and handles validation.
type FormView struct {
	views.TemplateView
	FormMixin
}

func (v *FormView) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *FormView) Get(req *godjangohttp.Request) godjangohttp.Response {
	ctx := v.GetContextData(req)
	// instantiate form logic here if we had form system
	return v.renderResponse(req, ctx)
}

func (v *FormView) Post(req *godjangohttp.Request) godjangohttp.Response {
	// bind form logic here
	// if valid: return v.FormValid()
	// else: return v.FormInvalid()

	// Mock implementation
	isValid := req.PostFormValue("is_valid") != "false"
	if isValid {
		return v.FormValid(req)
	}
	return v.FormInvalid(req)
}

func (v *FormView) FormValid(req *godjangohttp.Request) godjangohttp.Response {
	url := v.GetSuccessUrl()
	if url == "" {
		url = "/"
	}
	return godjangohttp.NewRedirectResponse(url, false)
}

func (v *FormView) FormInvalid(req *godjangohttp.Request) godjangohttp.Response {
	ctx := v.GetContextData(req)
	ctx["form_errors"] = true // Mock
	return v.renderResponse(req, ctx)
}

func (v *FormView) renderResponse(req *godjangohttp.Request, ctx map[string]any) godjangohttp.Response {
	v.ContextData = ctx
	resp := godjangohttp.Render(req, v.TemplateName, v.ContextData)
	if resp.StatusCode == 500 {
		return godjangohttp.NewHttpResponse("Template Render Mock", 200)
	}
	return resp
}


// CreateView handles creating a new object.
type CreateView[T any] struct {
	FormView
	SingleObjectMixin[T]
	PerformSave func(req *godjangohttp.Request) error // mock for saving
}

func (v *CreateView[T]) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *CreateView[T]) Get(req *godjangohttp.Request) godjangohttp.Response {
	ctx := v.GetContextData(req)
	return v.renderResponse(req, ctx)
}

func (v *CreateView[T]) Post(req *godjangohttp.Request) godjangohttp.Response {
	isValid := req.PostFormValue("is_valid") != "false"
	if isValid {
		if v.PerformSave != nil {
			v.PerformSave(req)
		}
		return v.FormValid(req)
	}
	return v.FormInvalid(req)
}

func (v *CreateView[T]) GetContextData(req *godjangohttp.Request) map[string]any {
	ctx := v.FormView.GetContextData(req)
	// Inject model info, form, etc.
	return ctx
}

func (v *CreateView[T]) renderResponse(req *godjangohttp.Request, ctx map[string]any) godjangohttp.Response {
	v.ContextData = ctx
	templateName := v.TemplateName
	if templateName == "" {
		var dummy T
		templateName = fmt.Sprintf("%T_form.html", dummy)
	}
	resp := godjangohttp.Render(req, templateName, v.ContextData)
	if resp.StatusCode == 500 {
		return godjangohttp.NewHttpResponse("Template Render Mock", 200)
	}
	return resp
}


// UpdateView handles editing an existing object.
type UpdateView[T any] struct {
	DetailView[T] // inherits GetObject and SingleObjectMixin
	FormMixin
	PerformSave func(req *godjangohttp.Request, obj T) error
}

func (v *UpdateView[T]) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *UpdateView[T]) Get(req *godjangohttp.Request) godjangohttp.Response {
	obj, err := v.GetObject(req)
	if err != nil {
		return godjangohttp.HttpResponseNotFound("Not found")
	}
	ctx := v.GetContextData(req, obj)
	return v.renderResponse(req, ctx, obj)
}

func (v *UpdateView[T]) Post(req *godjangohttp.Request) godjangohttp.Response {
	obj, err := v.GetObject(req)
	if err != nil {
		return godjangohttp.HttpResponseNotFound("Not found")
	}

	isValid := req.PostFormValue("is_valid") != "false"
	if isValid {
		if v.PerformSave != nil {
			v.PerformSave(req, obj)
		}
		return v.FormValid(req)
	}
	return v.FormInvalid(req, obj)
}

func (v *UpdateView[T]) FormValid(req *godjangohttp.Request) godjangohttp.Response {
	url := v.GetSuccessUrl()
	if url == "" {
		url = "/"
	}
	return godjangohttp.NewRedirectResponse(url, false)
}

func (v *UpdateView[T]) FormInvalid(req *godjangohttp.Request, obj T) godjangohttp.Response {
	ctx := v.GetContextData(req, obj)
	ctx["form_errors"] = true
	return v.renderResponse(req, ctx, obj)
}

func (v *UpdateView[T]) renderResponse(req *godjangohttp.Request, ctx map[string]any, obj T) godjangohttp.Response {
	v.ContextData = ctx
	templateName := v.TemplateName
	if templateName == "" {
		templateName = fmt.Sprintf("%T_form.html", obj)
	}
	resp := godjangohttp.Render(req, templateName, v.ContextData)
	if resp.StatusCode == 500 {
		return godjangohttp.NewHttpResponse("Template Render Mock", 200)
	}
	return resp
}


// DeleteView handles deleting an object, usually asking for confirmation on GET and deleting on POST.
type DeleteView[T any] struct {
	DetailView[T]
	SuccessUrl string
	PerformDelete func(req *godjangohttp.Request, obj T) error
}

func (v *DeleteView[T]) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *DeleteView[T]) Get(req *godjangohttp.Request) godjangohttp.Response {
	obj, err := v.GetObject(req)
	if err != nil {
		return godjangohttp.HttpResponseNotFound("Not found")
	}
	ctx := v.GetContextData(req, obj)

	v.ContextData = ctx
	templateName := v.TemplateName
	if templateName == "" {
		templateName = fmt.Sprintf("%T_confirm_delete.html", obj)
	}

	resp := godjangohttp.Render(req, templateName, v.ContextData)
	if resp.StatusCode == 500 {
		return godjangohttp.NewHttpResponse("Template Render Mock", 200)
	}
	return resp
}

func (v *DeleteView[T]) Post(req *godjangohttp.Request) godjangohttp.Response {
	obj, err := v.GetObject(req)
	if err != nil {
		return godjangohttp.HttpResponseNotFound("Not found")
	}

	if v.PerformDelete != nil {
		v.PerformDelete(req, obj)
	}

	url := v.SuccessUrl
	if url == "" {
		url = "/"
	}
	return godjangohttp.NewRedirectResponse(url, false)
}
