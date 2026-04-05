package views

import (
	godjangohttp "github.com/godjango/godjango/http"
)

// ViewFunc represents a function that handles an HTTP request.
type ViewFunc func(req *godjangohttp.Request) godjangohttp.Response
