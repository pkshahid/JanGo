package http_test

import (
	"fmt"
	"net/http/httptest"

	godjangohttp "github.com/pkshahid/JanGo/http"
)

// Example of building and returning an HTTP response.
func ExampleNewHttpResponse() {
	resp := godjangohttp.NewHttpResponse("Hello, GoDjango!", 200)
	resp.Headers.Set("Content-Type", "text/plain")

	rec := httptest.NewRecorder()
	resp.Write(rec)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
	// Output:
	// 200
	// Hello, GoDjango!
}
