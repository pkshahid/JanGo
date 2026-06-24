package i18n

// LanguagePair maps a language code to its human-readable name.
type LanguagePair struct {
	Code string
	Name string
}

// Settings holds global configuration for internationalization.
type Settings struct {
	LanguageCode string
	UseI18n      bool
	LocalePaths  []string
	Languages    []LanguagePair
	CookieName   string
}

// Global default settings.
var Config = Settings{
	LanguageCode: "en-us",
	UseI18n:      true,
	LocalePaths:  []string{"locale"},
	Languages: []LanguagePair{
		{"en", "English"},
	},
	CookieName: "godjango_language",
}
