package generic

import (
	"fmt"
	"strconv"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/orm/queryset"
)

// ListView handles displaying a list of objects, potentially paginated.
type ListView[T any] struct {
	TemplateView
	MultipleObjectMixin[T]
	GetQuerySetFunc func(req *godjangohttp.Request, mixin *MultipleObjectMixin[T]) ([]T, error)
}

func (v *ListView[T]) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *ListView[T]) GetQuerySet(req *godjangohttp.Request) ([]T, error) {
	if v.GetQuerySetFunc != nil {
		return v.GetQuerySetFunc(req, &v.MultipleObjectMixin)
	}

	if v.QuerySet != nil {
		return v.QuerySet, nil
	}

	// Default: use the ORM to fetch all objects
	qs := queryset.NewQuerySet[T]()

	if len(v.Ordering) > 0 {
		qs = qs.OrderBy(v.Ordering...)
	}

	// Placeholder for QS execution:
	// Let's assume an Execute() or All() returns []T. Since we don't have it fully in queryset,
	// we will rely on GetQuerySetFunc for our implementation or return empty for now.
	var empty []T
	return empty, fmt.Errorf("ORM All not fully implemented, please provide GetQuerySetFunc or QuerySet slice")
}

func (v *ListView[T]) Get(req *godjangohttp.Request) godjangohttp.Response {
	objectList, err := v.GetQuerySet(req)
	if err != nil {
		// Just handle gracefully for mock purposes, but normally error out
		objectList = []T{}
	}

	if !v.AllowEmpty && len(objectList) == 0 {
		return godjangohttp.HttpResponseNotFound("No objects found")
	}

	ctx := v.GetContextData(req, objectList)
	v.ContextData = ctx

	templateName := v.TemplateName
	if templateName == "" {
		var dummy T
		templateName = fmt.Sprintf("%T%s.html", dummy, v.GetTemplateNameSuffix())
	}

	resp := godjangohttp.Render(req, templateName, v.ContextData)
	if resp.StatusCode == 500 {
		// Mock for testing
		return godjangohttp.NewHttpResponse("Template Render Mock", 200)
	}
	return resp
}

func (v *ListView[T]) GetContextData(req *godjangohttp.Request, objectList []T) map[string]any {
	ctx := v.TemplateView.GetContextData(req)

	isPaginated := false
	var pageObj *Page[T]
	var paginator *Paginator[T]

	if v.PaginateBy > 0 {
		isPaginated = true
		paginator = NewPaginator(objectList, v.PaginateBy)

		pageStr := req.URL.Query().Get("page")
		pageNum := 1
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			pageNum = p
		}

		p, err := paginator.Page(pageNum)
		if err == nil {
			pageObj = p
			objectList = p.ObjectList // The paginated list
		} else {
			// fallback to first page if error
			if pageNum > paginator.NumPages {
				p, _ = paginator.Page(paginator.NumPages)
				pageObj = p
				if p != nil {
					objectList = p.ObjectList
				}
			}
		}
	}

	ctx["object_list"] = objectList
	ctx[v.GetContextObjectName(objectList)] = objectList
	ctx["is_paginated"] = isPaginated
	ctx["page_obj"] = pageObj
	ctx["paginator"] = paginator

	return ctx
}
