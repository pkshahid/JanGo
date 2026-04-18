package i18n

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Translation maps msgid to msgstr.
// Plurals are mapped as `msgid|context|plural_index` or we can use a struct.
type Translation struct {
	MsgId       string
	MsgIdPlural string
	Context     string
	MsgStrs     []string // Index 0 is singular, 1+ are plurals
}

// Catalog holds translations for all languages.
type Catalog struct {
	mu           sync.RWMutex
	translations map[string]map[string]*Translation // lang -> msgid(with context) -> Translation
}

var globalCatalog = &Catalog{
	translations: make(map[string]map[string]*Translation),
}

// MakeKey generates a key for looking up a translation.
func MakeKey(context, msgid string) string {
	if context == "" {
		return msgid
	}
	return context + "\x04" + msgid
}

// LoadPaths parses all .po files in the configured paths.
func LoadPaths(paths []string) error {
	globalCatalog.mu.Lock()
	defer globalCatalog.mu.Unlock()

	for _, p := range paths {
		err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".po") {
				// e.g., locale/es/LC_MESSAGES/django.po
				parts := strings.Split(filepath.ToSlash(path), "/")
				if len(parts) >= 3 {
					lang := parts[len(parts)-3]
					parsePOFile(lang, path)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// parsePOFile reads a PO file and adds it to the catalog.
func parsePOFile(lang, path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	if _, ok := globalCatalog.translations[lang]; !ok {
		globalCatalog.translations[lang] = make(map[string]*Translation)
	}
	dict := globalCatalog.translations[lang]

	scanner := bufio.NewScanner(file)

	var msgctxt, msgid, msgidPlural string
	var msgstrs []string

	saveEntry := func() {
		if msgid != "" {
			key := MakeKey(msgctxt, msgid)
			t := &Translation{
				MsgId:       msgid,
				MsgIdPlural: msgidPlural,
				Context:     msgctxt,
			}
			// Copy msgstrs
			t.MsgStrs = make([]string, len(msgstrs))
			copy(t.MsgStrs, msgstrs)

			// Fill empty plural forms if missing
			if len(t.MsgStrs) == 0 {
				t.MsgStrs = []string{""}
			}
			dict[key] = t
		}
		// Reset
		msgctxt, msgid, msgidPlural = "", "", ""
		msgstrs = nil
	}

	state := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "msgctxt ") {
			saveEntry()
			msgctxt = unquote(strings.TrimPrefix(line, "msgctxt "))
			state = "msgctxt"
		} else if strings.HasPrefix(line, "msgid ") {
			if state != "msgctxt" {
				saveEntry()
			}
			msgid = unquote(strings.TrimPrefix(line, "msgid "))
			state = "msgid"
		} else if strings.HasPrefix(line, "msgid_plural ") {
			msgidPlural = unquote(strings.TrimPrefix(line, "msgid_plural "))
			state = "msgid_plural"
		} else if strings.HasPrefix(line, "msgstr ") {
			str := unquote(strings.TrimPrefix(line, "msgstr "))
			msgstrs = append(msgstrs, str) // Usually just index 0
			state = "msgstr"
		} else if strings.HasPrefix(line, "msgstr[") {
			// Extract index
			idxEnd := strings.Index(line, "]")
			if idxEnd > 7 {
				// we assume indices come in order 0, 1, 2...
				str := unquote(line[idxEnd+2:])
				msgstrs = append(msgstrs, str)
			}
			state = "msgstr_plural"
		} else if strings.HasPrefix(line, "\"") && strings.HasSuffix(line, "\"") {
			// Continuation line
			val := unquote(line)
			switch state {
			case "msgctxt": msgctxt += val
			case "msgid": msgid += val
			case "msgid_plural": msgidPlural += val
			case "msgstr": msgstrs[0] += val
			case "msgstr_plural":
				if len(msgstrs) > 0 {
					msgstrs[len(msgstrs)-1] += val
				}
			}
		}
	}
	saveEntry()
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		// handle simple escapes
		s = s[1:len(s)-1]
		s = strings.ReplaceAll(s, "\\n", "\n")
		s = strings.ReplaceAll(s, "\\\"", "\"")
		return s
	}
	return s
}

// lookup finds a translation string.
func lookup(lang, context, msgid string, n int) string {
	globalCatalog.mu.RLock()
	defer globalCatalog.mu.RUnlock()

	dict, ok := globalCatalog.translations[lang]
	if !ok {
		// Fallback to base language if a locale like en-us is missing but en exists
		parts := strings.Split(lang, "-")
		if len(parts) > 1 {
			dict, ok = globalCatalog.translations[parts[0]]
		}
	}

	if !ok {
		return ""
	}

	key := MakeKey(context, msgid)
	t, ok := dict[key]
	if !ok {
		return ""
	}

	// Plural rule logic (simplified)
	idx := 0
	if t.MsgIdPlural != "" {
		idx = getPluralIndex(lang, n)
	}

	if idx < len(t.MsgStrs) && t.MsgStrs[idx] != "" {
		return t.MsgStrs[idx]
	}
	return ""
}

// getPluralIndex returns the index of the plural form for a given language.
func getPluralIndex(lang string, n int) int {
	// A real implementation would parse the Plural-Forms header from the .po file
	// or use golang.org/x/text/feature/plural.
	// We'll use a simplified set of common rules.
	langBase := strings.Split(lang, "-")[0]

	switch langBase {
	case "ja", "zh", "ko":
		return 0 // Only 1 form
	case "fr", "pt":
		if n > 1 {
			return 1
		}
		return 0
	case "ru", "uk", "be", "hr", "sr":
		if n%10 == 1 && n%100 != 11 {
			return 0
		} else if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
			return 1
		}
		return 2
	case "pl":
		if n == 1 {
			return 0
		} else if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
			return 1
		}
		return 2
	case "ar":
		if n == 0 { return 0 }
		if n == 1 { return 1 }
		if n == 2 { return 2 }
		if n%100 >= 3 && n%100 <= 10 { return 3 }
		if n%100 >= 11 && n%100 <= 99 { return 4 }
		return 5
	default:
		// Germanic languages (en, de, nl, sv, da) + es, it, etc.
		if n != 1 {
			return 1
		}
		return 0
	}
}

// For testing purposes
func AddTranslation(lang, msgid, msgstr string) {
	globalCatalog.mu.Lock()
	defer globalCatalog.mu.Unlock()
	if _, ok := globalCatalog.translations[lang]; !ok {
		globalCatalog.translations[lang] = make(map[string]*Translation)
	}
	globalCatalog.translations[lang][msgid] = &Translation{
		MsgId:   msgid,
		MsgStrs: []string{msgstr},
	}
}
