package middleware

import (
	godjangohttp "github.com/godjango/godjango/http"
)

// MessageMiddleware manages flash messages in the session.
// It depends on SessionMiddleware.
func MessageMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		if req.Session != nil {
			// Extract messages to make them available in views/templates
			msgs, _ := req.Session.Get("_messages")
			if msgs != nil {
				req.META["MESSAGES"] = msgs.(string)
				req.Session.Delete("_messages") // Consume messages
			}
		}

		resp := next(req)
		return resp
	}
}

// AddMessage is a utility to add a message to the session
func AddMessage(req *godjangohttp.Request, message string) {
	if req.Session != nil {
		existing, _ := req.Session.Get("_messages")
		if existing != nil {
			req.Session.Set("_messages", existing.(string)+"\n"+message)
		} else {
			req.Session.Set("_messages", message)
		}
	}
}
