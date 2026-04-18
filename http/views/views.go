package views

import (
	godjangohttp "github.com/pkshahid/JanGo/http"
)

// ViewFunc represents a function that handles an HTTP request.
type ViewFunc func(req *godjangohttp.Request) godjangohttp.Response
