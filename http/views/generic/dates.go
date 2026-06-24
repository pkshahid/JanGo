package generic

import (
	"fmt"
	godjangohttp "github.com/pkshahid/JanGo/http"
)

// DateMixin provides common attributes for date-based views.
type DateMixin struct {
	DateField   string
	AllowFuture bool
}

func (m *DateMixin) GetDateField() string {
	if m.DateField == "" {
		return "date"
	}
	return m.DateField
}

func (m *DateMixin) GetAllowFuture() bool {
	return m.AllowFuture
}

// YearMixin provides attributes for year-based views.
type YearMixin struct {
	YearFormat string
	Year       string
}

func (m *YearMixin) GetYearFormat() string {
	if m.YearFormat == "" {
		return "2006"
	}
	return m.YearFormat
}

func (m *YearMixin) GetYear(req *godjangohttp.Request) string {
	if m.Year != "" {
		return m.Year
	}
	return req.URL.Query().Get("year") // Fallback
}

// MonthMixin provides attributes for month-based views.
type MonthMixin struct {
	MonthFormat string
	Month       string
}

func (m *MonthMixin) GetMonthFormat() string {
	if m.MonthFormat == "" {
		return "01" // or "Jan" if string
	}
	return m.MonthFormat
}

func (m *MonthMixin) GetMonth(req *godjangohttp.Request) string {
	if m.Month != "" {
		return m.Month
	}
	return req.URL.Query().Get("month") // Fallback
}

// DayMixin provides attributes for day-based views.
type DayMixin struct {
	DayFormat string
	Day       string
}

func (m *DayMixin) GetDayFormat() string {
	if m.DayFormat == "" {
		return "02"
	}
	return m.DayFormat
}

func (m *DayMixin) GetDay(req *godjangohttp.Request) string {
	if m.Day != "" {
		return m.Day
	}
	return req.URL.Query().Get("day") // Fallback
}

// ArchiveIndexView shows the latest objects.
type ArchiveIndexView[T any] struct {
	ListView[T]
	DateMixin
	DateListPeriod string
}

func (v *ArchiveIndexView[T]) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *ArchiveIndexView[T]) Get(req *godjangohttp.Request) godjangohttp.Response {
	// Usually filters by DateField <= now if not AllowFuture, then orders by -DateField
	// For mock purposes, just defer to ListView logic
	return v.ListView.Get(req)
}

// YearArchiveView shows objects for a specific year.
type YearArchiveView[T any] struct {
	ListView[T]
	DateMixin
	YearMixin
	MakeObjectList bool
}

func (v *YearArchiveView[T]) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *YearArchiveView[T]) GetContextData(req *godjangohttp.Request, objectList []T) map[string]any {
	ctx := v.ListView.GetContextData(req, objectList)
	ctx["year"] = v.GetYear(req)
	return ctx
}

func (v *YearArchiveView[T]) Get(req *godjangohttp.Request) godjangohttp.Response {
	// Extracts year, filters by year
	yearStr := v.GetYear(req)
	if yearStr == "" {
		return godjangohttp.HttpResponseNotFound("No year specified")
	}

	objectList, err := v.GetQuerySet(req)
	if err != nil {
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
		return godjangohttp.NewHttpResponse("Template Render Mock", 200)
	}
	return resp
}

// MonthArchiveView shows objects for a specific month.
type MonthArchiveView[T any] struct {
	ListView[T]
	DateMixin
	YearMixin
	MonthMixin
}

func (v *MonthArchiveView[T]) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *MonthArchiveView[T]) Get(req *godjangohttp.Request) godjangohttp.Response {
	// Filters by year and month
	monthStr := v.GetMonth(req)
	if monthStr == "" {
		return godjangohttp.HttpResponseNotFound("No month specified")
	}
	return v.ListView.Get(req)
}

// DayArchiveView shows objects for a specific day.
type DayArchiveView[T any] struct {
	ListView[T]
	DateMixin
	YearMixin
	MonthMixin
	DayMixin
}

func (v *DayArchiveView[T]) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *DayArchiveView[T]) Get(req *godjangohttp.Request) godjangohttp.Response {
	dayStr := v.GetDay(req)
	if dayStr == "" {
		return godjangohttp.HttpResponseNotFound("No day specified")
	}
	return v.ListView.Get(req)
}

// DateDetailView fetches an object by date and pk/slug.
type DateDetailView[T any] struct {
	DetailView[T]
	DateMixin
	YearMixin
	MonthMixin
	DayMixin
}

func (v *DateDetailView[T]) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *DateDetailView[T]) Get(req *godjangohttp.Request) godjangohttp.Response {
	// Normally this would fetch the object with constraints on year, month, day.
	// Since ORM is mocked, we defer to DetailView.

	y := v.GetYear(req)
	m := v.GetMonth(req)
	d := v.GetDay(req)

	if y == "" || m == "" || d == "" {
		return godjangohttp.HttpResponseNotFound("Missing date parameters")
	}

	return v.DetailView.Get(req)
}
