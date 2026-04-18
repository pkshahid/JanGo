package i18n

import (
	"strings"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/middleware"
	"golang.org/x/text/language"
)

// LocaleMiddleware detects the user's preferred language and activates it in the request context.
type LocaleMiddleware struct{}

// NewLocaleMiddleware creates a new LocaleMiddleware.
func NewLocaleMiddleware() *LocaleMiddleware {
	return &LocaleMiddleware{}
}

func (m *LocaleMiddleware) Process(next middleware.Handler) middleware.Handler {
	return func(req *godjangohttp.Request) godjangohttp.Response {
		if !Config.UseI18n {
			return next(req)
		}

		lang := m.getLanguageFromRequest(req)

		// Activate language in context
		newReq := godjangohttp.WithValue(req, languageContextKey, lang)
		// Usually we'd want to expose LanguageCode. We'll store it in META for views to access or context.
		newReq.META["LANGUAGE_CODE"] = lang

		resp := next(newReq)

		// Set Content-Language header on response if not already set
		if httpResp, ok := resp.(*godjangohttp.HttpResponse); ok {
			if httpResp.Headers.Get("Content-Language") == "" {
				httpResp.Headers.Set("Content-Language", lang)
			}
		}

		return resp
	}
}

func (m *LocaleMiddleware) getLanguageFromRequest(req *godjangohttp.Request) string {
	// 1. Check URL path prefix (e.g., /en/...)
	path := req.URL.Path
	for _, lp := range Config.Languages {
		prefix := "/" + lp.Code + "/"
		if strings.HasPrefix(path, prefix) || path == "/"+lp.Code {
			return lp.Code
		}
	}

	// 2. Check Cookie/Session
	cookie, err := req.Cookie(Config.CookieName)
	if err == nil && cookie.Value != "" {
		if m.isSupportedLanguage(cookie.Value) {
			return cookie.Value
		}
	}

	// Assume Session fallback if configured in a full implementation...

	// 3. Check Accept-Language header
	acceptLang := req.Header.Get("Accept-Language")
	if acceptLang != "" {
		// Use golang.org/x/text/language to parse and match
		var supported []language.Tag
		for _, lp := range Config.Languages {
			t, err := language.Parse(lp.Code)
			if err == nil {
				supported = append(supported, t)
			}
		}

		if len(supported) > 0 {
			matcher := language.NewMatcher(supported)
			tags, _, _ := language.ParseAcceptLanguage(acceptLang)
			tag, _, _ := matcher.Match(tags...)
			return tag.String() // Might need formatting to match Config exactly
		}
	}

	return Config.LanguageCode
}

func (m *LocaleMiddleware) isSupportedLanguage(lang string) bool {
	for _, lp := range Config.Languages {
		if lp.Code == lang {
			return true
		}
	}
	return false
}
