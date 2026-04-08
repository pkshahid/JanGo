package i18n

import (
	"context"
)

// Gettext translates a singular string.
// Note: Django's gettext acts on the current thread's active language.
// In Go, we must explicitly pass the context.
func Gettext(ctx context.Context, message string) string {
	lang := GetLanguage(ctx)
	if translated := lookup(lang, "", message, 1); translated != "" {
		return translated
	}
	return message
}

// Ngettext translates a plural string.
func Ngettext(ctx context.Context, singular, plural string, n int) string {
	lang := GetLanguage(ctx)
	if translated := lookup(lang, "", singular, n); translated != "" {
		return translated
	}
	if n == 1 {
		return singular
	}
	return plural
}

// Pgettext translates a string with context.
func Pgettext(ctx context.Context, ctxt, message string) string {
	lang := GetLanguage(ctx)
	if translated := lookup(lang, ctxt, message, 1); translated != "" {
		return translated
	}
	return message
}

// Npgettext translates a plural string with context.
func Npgettext(ctx context.Context, ctxt, singular, plural string, n int) string {
	lang := GetLanguage(ctx)
	if translated := lookup(lang, ctxt, singular, n); translated != "" {
		return translated
	}
	if n == 1 {
		return singular
	}
	return plural
}

// LazyString represents a string that is translated at the moment of rendering.
type LazyString struct {
	Message string
	Context string
	Plural  string
	N       int
	IsPlural bool
}

// String evaluates the lazy string using a given context.
func (ls LazyString) String(ctx context.Context) string {
	if ls.IsPlural {
		if ls.Context != "" {
			return Npgettext(ctx, ls.Context, ls.Message, ls.Plural, ls.N)
		}
		return Ngettext(ctx, ls.Message, ls.Plural, ls.N)
	}

	if ls.Context != "" {
		return Pgettext(ctx, ls.Context, ls.Message)
	}
	return Gettext(ctx, ls.Message)
}

// GettextLazy returns a LazyString for singular translation.
func GettextLazy(message string) LazyString {
	return LazyString{Message: message}
}

// PgettextLazy returns a LazyString for contextual translation.
func PgettextLazy(ctxt, message string) LazyString {
	return LazyString{Context: ctxt, Message: message}
}
