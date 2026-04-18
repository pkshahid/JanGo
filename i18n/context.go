package i18n

import (
	"context"
)

type languageContextKeyType int

const languageContextKey languageContextKeyType = iota

// Activate returns a new context with the given language code activated.
func Activate(ctx context.Context, langCode string) context.Context {
	return context.WithValue(ctx, languageContextKey, langCode)
}

// GetLanguage retrieves the current active language code from the context.
func GetLanguage(ctx context.Context) string {
	if ctx == nil {
		return Config.LanguageCode
	}
	if lang, ok := ctx.Value(languageContextKey).(string); ok && lang != "" {
		return lang
	}
	return Config.LanguageCode
}

// GlobalOverride provides a way to globally override language per thread/goroutine.
// In Go, since there's no true thread-local storage, this returns a function
// that can be deferred if you maintain state in a context pointer, or it just modifies
// a pointer to a context.
// For thread safety, language MUST be passed via context.
// We provide Override for scoped changes within a function.
func Override(ctx context.Context, langCode string) (context.Context, func() context.Context) {
	origCtx := ctx
	newCtx := Activate(ctx, langCode)
	restore := func() context.Context {
		return origCtx
	}
	return newCtx, restore
}
