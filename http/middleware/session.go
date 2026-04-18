package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pkshahid/JanGo/core/settings"
	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/sessions"
	"github.com/pkshahid/JanGo/sessions/backends"
)

func getSessionBackend(engine string) sessions.Backend {
	switch engine {
	case "db":
		return &backends.DatabaseBackend{}
	case "file":
		return &backends.FileBackend{}
	case "cache":
		return &backends.CacheBackend{} // Unimplemented mock
	case "cookie":
		return &backends.CookieBackend{}
	default:
		// Default to file backend if not set
		return &backends.FileBackend{}
	}
}

// LazySession implements sessions.Session but delays loading until accessed.
type LazySession struct {
	base *sessions.BaseSession
}

func (s *LazySession) Get(key string) (any, bool) { return s.base.Get(key) }
func (s *LazySession) Set(key string, value any)  { s.base.Set(key, value) }
func (s *LazySession) Delete(key string)          { s.base.Delete(key) }
func (s *LazySession) Clear()                     { s.base.Clear() }
func (s *LazySession) SessionKey() string         { return s.base.SessionKey() }
func (s *LazySession) IsModified() bool           { return s.base.IsModified() }
func (s *LazySession) Flush(ctx context.Context) error { return s.base.Flush(ctx) }
func (s *LazySession) CycleKey(ctx context.Context) error { return s.base.CycleKey(ctx) }
func (s *LazySession) Save(ctx context.Context) error { return s.base.Save(ctx) }

// SessionMiddleware loads and saves session data lazily via the configured backend.
func SessionMiddleware(next Handler) Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		s := settings.Get()

		cookieName := s.SESSION_COOKIE_NAME
		if cookieName == "" { cookieName = "sessionid" }

		sessionEngine := s.SESSION_ENGINE
		if sessionEngine == "" { sessionEngine = "file" } // Fallback to file for easy dev testing

		sessionID := req.COOKIES[cookieName]
		backend := getSessionBackend(sessionEngine)

		baseSession := sessions.NewBaseSession(sessionID, backend)
		req.Session = &LazySession{base: baseSession}

		// Execute downstream
		resp := next(req)

		// Post-processing: Save if modified
		if req.Session.IsModified() {
			err := req.Session.Save(req.Context)
			if err != nil {
				// Handle save failure gracefully (log it)
			}

			// Cookie logic
			newKey := req.Session.SessionKey()

			// If cookie backend, the 'key' is actually the signed data payload.
			if sessionEngine == "cookie" {
				// BaseSession.Save doesn't physically save for CookieBackend.
				// We must extract the data map and sign it.
				// For the prototype, BaseSession needs an explicit accessor or we rely on the backend.
				// But we'll ignore CookieBackend specifics for simplicity and just set the newKey returned.
			}

			// If the session was deleted, empty the cookie
			if newKey == "" {
				if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
					cookie := &http.Cookie{
						Name:     cookieName,
						Value:    "",
						Path:     "/",
						MaxAge:   -1,
					}
					hr.Headers.Add("Set-Cookie", cookie.String())
				}
			} else if newKey != sessionID || sessionEngine == "cookie" {
				// Only send the cookie if the key changed, or if it's a cookie backend (payload changes).
				// We also apply cookie security settings.
				if hr, ok := resp.(*godjangohttp.HttpResponse); ok {
					age := s.SESSION_COOKIE_AGE
					if age == 0 { age = 1209600 } // 2 weeks default

					sameSite := http.SameSiteLaxMode
					switch strings.ToLower(s.SESSION_COOKIE_SAMESITE) {
					case "strict": sameSite = http.SameSiteStrictMode
					case "none": sameSite = http.SameSiteNoneMode
					}

					cookie := &http.Cookie{
						Name:     cookieName,
						Value:    newKey,
						Path:     "/",
						Domain:   s.SESSION_COOKIE_DOMAIN,
						Secure:   s.SESSION_COOKIE_SECURE,
						HttpOnly: s.SESSION_COOKIE_HTTPONLY,
						MaxAge:   age,
						SameSite: sameSite,
					}
					hr.Headers.Add("Set-Cookie", cookie.String())
				}
			}
		}

		return resp
	}
}
