package middleware

import (
	"strings"
	"net/smtp"

	httpsignals "github.com/pkshahid/JanGo/http/signals"

	"fmt"
	"log/slog"
	"runtime"

	"github.com/pkshahid/JanGo/core/settings"
	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/exceptions"
	"github.com/pkshahid/JanGo/monitoring"
)

// PanicRecoveryMiddleware catches panics, logs the stack trace, and returns a 500 error page.
func PanicRecoveryMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) (resp godjangohttp.Response) {
		defer func() {
			if r := recover(); r != nil {
				// Get stack trace
				buf := make([]byte, 1<<20) // 1MB
				n := runtime.Stack(buf, true)
				traceback := string(buf[:n])

				slog.Error("Panic recovered",
					slog.Any("recover", r),
					slog.String("traceback", traceback),
				)

				var err error
				if e, ok := r.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("%v", r)
				}

				// Capture error via monitoring
				monitoring.CaptureError(err, req)

				httpsignals.GotRequestException.Send(req, map[string]any{"error": err})

				s := settings.Get()

				// If production, send email to ADMINS
				if !s.DEBUG && len(s.ADMINS) > 0 {
					go func(admins []string, err error, trace string) {
						// Setup email
						subject := fmt.Sprintf("[GoDjango] Error (%s): %v", req.Path, err)
						body := fmt.Sprintf("Path: %s\nMethod: %s\n\nTraceback:\n%s", req.Path, req.Method, trace)
						s := settings.Get()

						// Very basic SMTP implementation using net/smtp
						// In a real framework, this would go through a `core/mail` backend
						// which respects EMAIL_BACKEND. But for this test, we use SMTP directly or dummy log.

						if s.EMAIL_HOST != "" {
							auth := smtp.PlainAuth("", "", "", s.EMAIL_HOST) // Simplification for test
							addr := fmt.Sprintf("%s:%d", s.EMAIL_HOST, s.EMAIL_PORT)

							msg := []byte("To: " + strings.Join(admins, ",") + "\r\nSubject: " + subject + "\r\n\r\n" + body + "\r\n")
								// Fire and forget
							err := smtp.SendMail(addr, auth, "errors@godjango.local", admins, msg)
							if err != nil {
								slog.Error("Failed to send 500 error email", slog.Any("error", err))
							}
						} else {
							// For testing and visibility whenEMAIL_HOST is not set
							slog.Info("Simulating 500 error email to ADMINS", slog.Any("admins", admins), slog.String("subject", subject))
						}
					}(s.ADMINS, err, traceback)
				}

				// Resolve the correct error page
				if s.DEBUG {
					resp = exceptions.RenderDebug500(req, r, traceback)
				} else {
					resolver := &exceptions.Resolver{}
					resp = resolver.Resolve500(req)
				}
			}
		}()

		// Try executing the handler
		resp = next(req)

		// Also check if the handler returned an error object natively?
		// Handlers return Response, so exceptions are usually caught via panics or
		// explicitly returned as HttpResponse objects. Let's assume errors panic in GoDjango
		// or are handled. If a specific error is thrown (like Http404), maybe we should recover it too.
		// Wait, panic usually carries the exception type. Let's see if we can catch specific ones.

		return resp
	}
}

// WrapExceptionsMiddleware catches errors returned as responses and transforms them.
// But wait, GoDjango handler signature is `func(req *Request) Response`.
// So standard exceptions like Http404 are typically paniced.
// Let's modify PanicRecoveryMiddleware to handle known exceptions.
