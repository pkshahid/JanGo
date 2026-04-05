package views

import (
	"net/http"

	"github.com/godjango/godjango/core/settings"
	godjangohttp "github.com/godjango/godjango/http"
)

// PageNotFound renders a 404 response.
func PageNotFound(req *godjangohttp.Request, err error) godjangohttp.Response {
	s := settings.Get()
	if s.DEBUG {
		msg := "Not Found"
		if err != nil {
			msg += "\n" + err.Error()
		}
		return godjangohttp.NewHttpResponse(msg, http.StatusNotFound)
	}
	return godjangohttp.HttpResponseNotFound("Not Found")
}

// ServerError renders a 500 response.
func ServerError(req *godjangohttp.Request) godjangohttp.Response {
	s := settings.Get()
	if s.DEBUG {
		return godjangohttp.NewHttpResponse("Internal Server Error\nCheck logs for details.", http.StatusInternalServerError)
	}
	return godjangohttp.NewHttpResponse("Server Error", http.StatusInternalServerError)
}

// PermissionDenied renders a 403 response.
func PermissionDenied(req *godjangohttp.Request, err error) godjangohttp.Response {
	s := settings.Get()
	if s.DEBUG {
		msg := "Forbidden"
		if err != nil {
			msg += "\n" + err.Error()
		}
		return godjangohttp.NewHttpResponse(msg, http.StatusForbidden)
	}
	return godjangohttp.HttpResponseForbidden("Forbidden")
}

// BadRequest renders a 400 response.
func BadRequest(req *godjangohttp.Request, err error) godjangohttp.Response {
	s := settings.Get()
	if s.DEBUG {
		msg := "Bad Request"
		if err != nil {
			msg += "\n" + err.Error()
		}
		return godjangohttp.NewHttpResponse(msg, http.StatusBadRequest)
	}
	return godjangohttp.HttpResponseBadRequest("Bad Request")
}
