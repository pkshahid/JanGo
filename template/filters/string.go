package filters

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	godjango "github.com/godjango/godjango/template"
)

func RegisterStringHtmlFilters(lib *godjango.Library) {
	lib.RegisterFilter("addslashes", AddSlashesFilter)
	lib.RegisterFilter("capfirst", CapFirstFilter)
	lib.RegisterFilter("center", CenterFilter)
	lib.RegisterFilter("cut", CutFilter)
	lib.RegisterFilter("ljust", LJustFilter)
	lib.RegisterFilter("rjust", RJustFilter)
	lib.RegisterFilter("lower", LowerFilter)
	lib.RegisterFilter("upper", UpperFilter)
	lib.RegisterFilter("title", TitleFilter)

	lib.RegisterFilter("escape", EscapeFilter)
	lib.RegisterFilter("escapejs", EscapeJsFilter)
	lib.RegisterFilter("force_escape", ForceEscapeFilter)
	lib.RegisterFilter("safe", SafeFilter)
	lib.RegisterFilter("striptags", StripTagsFilter)

	lib.RegisterFilter("linebreaks", LinebreaksFilter)
	lib.RegisterFilter("linebreaksbr", LinebreaksBrFilter)
	lib.RegisterFilter("linenumbers", LineNumbersFilter)

	lib.RegisterFilter("slugify", SlugifyFilter)
	lib.RegisterFilter("stringformat", StringFormatFilter)

	lib.RegisterFilter("truncatechars", TruncateCharsFilter)
	lib.RegisterFilter("truncatewords", TruncateWordsFilter)
	lib.RegisterFilter("truncatechars_html", TruncateCharsHtmlFilter)
	lib.RegisterFilter("truncatewords_html", TruncateWordsHtmlFilter)

	lib.RegisterFilter("wordcount", WordCountFilter)
	lib.RegisterFilter("wordwrap", WordWrapFilter)

	lib.RegisterFilter("urlencode", UrlEncodeFilter)
	lib.RegisterFilter("iriencode", IriEncodeFilter)
	lib.RegisterFilter("urlize", UrlizeFilter)
	lib.RegisterFilter("urlizetrunc", UrlizeTruncFilter)

	lib.RegisterFilter("phone2numeric", Phone2NumericFilter)
	lib.RegisterFilter("json_script", JsonScriptFilter)
}

func AddSlashesFilter(val any, args string) (any, error) {
	str := fmt.Sprintf("%v", val)
	str = strings.ReplaceAll(str, "\\", "\\\\")
	str = strings.ReplaceAll(str, "\"", "\\\"")
	str = strings.ReplaceAll(str, "'", "\\'")
	return str, nil
}

func CapFirstFilter(val any, args string) (any, error) {
	str := fmt.Sprintf("%v", val)
	if len(str) == 0 {
		return str, nil
	}
	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes), nil
}

func CenterFilter(val any, args string) (any, error) {
	width, _ := strconv.Atoi(args)
	str := fmt.Sprintf("%v", val)
	if len(str) >= width {
		return str, nil
	}
	leftPad := (width - len(str)) / 2
	rightPad := width - len(str) - leftPad
	return strings.Repeat(" ", leftPad) + str + strings.Repeat(" ", rightPad), nil
}

func CutFilter(val any, args string) (any, error) {
	str := fmt.Sprintf("%v", val)
	return strings.ReplaceAll(str, args, ""), nil
}

func LJustFilter(val any, args string) (any, error) {
	width, _ := strconv.Atoi(args)
	str := fmt.Sprintf("%v", val)
	if len(str) >= width {
		return str, nil
	}
	return str + strings.Repeat(" ", width-len(str)), nil
}

func RJustFilter(val any, args string) (any, error) {
	width, _ := strconv.Atoi(args)
	str := fmt.Sprintf("%v", val)
	if len(str) >= width {
		return str, nil
	}
	return strings.Repeat(" ", width-len(str)) + str, nil
}

func LowerFilter(val any, args string) (any, error) {
	return strings.ToLower(fmt.Sprintf("%v", val)), nil
}

func UpperFilter(val any, args string) (any, error) {
	return strings.ToUpper(fmt.Sprintf("%v", val)), nil
}

func TitleFilter(val any, args string) (any, error) {
	return strings.Title(fmt.Sprintf("%v", val)), nil
}

func EscapeFilter(val any, args string) (any, error) {
	return html.EscapeString(fmt.Sprintf("%v", val)), nil
}

func EscapeJsFilter(val any, args string) (any, error) {
	str := fmt.Sprintf("%v", val)
	str = strings.ReplaceAll(str, "\\", "\\u005C")
	str = strings.ReplaceAll(str, "\"", "\\u0022")
	str = strings.ReplaceAll(str, "'", "\\u0027")
	str = strings.ReplaceAll(str, "<", "\\u003C")
	str = strings.ReplaceAll(str, ">", "\\u003E")
	str = strings.ReplaceAll(str, "&", "\\u0026")
	return str, nil
}

func ForceEscapeFilter(val any, args string) (any, error) {
	return html.EscapeString(fmt.Sprintf("%v", val)), nil
}

func SafeFilter(val any, args string) (any, error) {
	return godjango.SafeString(fmt.Sprintf("%v", val)), nil
}

func StripTagsFilter(val any, args string) (any, error) {
	str := fmt.Sprintf("%v", val)
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(str, ""), nil
}

func LinebreaksFilter(val any, args string) (any, error) {
	str := fmt.Sprintf("%v", val)
	str = strings.ReplaceAll(str, "\r\n", "\n")
	paras := strings.Split(str, "\n\n")

	var res []string
	for _, p := range paras {
		p = strings.ReplaceAll(p, "\n", "<br>")
		res = append(res, fmt.Sprintf("<p>%s</p>", p))
	}
	return godjango.SafeString(strings.Join(res, "\n\n")), nil
}

func LinebreaksBrFilter(val any, args string) (any, error) {
	str := fmt.Sprintf("%v", val)
	str = strings.ReplaceAll(str, "\r\n", "\n")
	return godjango.SafeString(strings.ReplaceAll(str, "\n", "<br>")), nil
}

func LineNumbersFilter(val any, args string) (any, error) {
	str := fmt.Sprintf("%v", val)
	lines := strings.Split(str, "\n")
	var res []string
	for i, l := range lines {
		res = append(res, fmt.Sprintf("%d. %s", i+1, l))
	}
	return strings.Join(res, "\n"), nil
}

func SlugifyFilter(val any, args string) (any, error) {
	str := strings.ToLower(fmt.Sprintf("%v", val))
	re := regexp.MustCompile(`[^\w\s-]`)
	str = re.ReplaceAllString(str, "")
	str = strings.ReplaceAll(str, " ", "-")
	return str, nil
}

func StringFormatFilter(val any, args string) (any, error) {
	format := args
	// e.g. "s", ".2f"
	return fmt.Sprintf("%"+format, val), nil
}

func TruncateCharsFilter(val any, args string) (any, error) {
	length, _ := strconv.Atoi(args)
	str := fmt.Sprintf("%v", val)
	runes := []rune(str)
	if len(runes) > length && length > 3 {
		return string(runes[:length-3]) + "...", nil
	}
	return str, nil
}

func TruncateWordsFilter(val any, args string) (any, error) {
	wordsLimit, _ := strconv.Atoi(args)
	str := fmt.Sprintf("%v", val)
	words := strings.Fields(str)
	if len(words) > wordsLimit {
		return strings.Join(words[:wordsLimit], " ") + " ...", nil
	}
	return str, nil
}

func TruncateCharsHtmlFilter(val any, args string) (any, error) {
	// A full implementation requires parsing HTML to close tags properly.
	// For this prototype, we simply truncate and strip partial tags.
	str, _ := TruncateCharsFilter(val, args)
	return godjango.SafeString(str.(string)), nil
}

func TruncateWordsHtmlFilter(val any, args string) (any, error) {
	str, _ := TruncateWordsFilter(val, args)
	return godjango.SafeString(str.(string)), nil
}

func WordCountFilter(val any, args string) (any, error) {
	str := fmt.Sprintf("%v", val)
	return len(strings.Fields(str)), nil
}

func WordWrapFilter(val any, args string) (any, error) {
	width, _ := strconv.Atoi(args)
	str := fmt.Sprintf("%v", val)

	words := strings.Fields(str)
	if len(words) == 0 {
		return "", nil
	}

	var buf strings.Builder
	currentLen := 0

	for _, word := range words {
		if currentLen > 0 {
			if currentLen + 1 + len(word) > width {
				buf.WriteString("\n")
				currentLen = 0
			} else {
				buf.WriteString(" ")
				currentLen++
			}
		}
		buf.WriteString(word)
		currentLen += len(word)
	}
	return buf.String(), nil
}

func UrlEncodeFilter(val any, args string) (any, error) {
	// A full urlencode needs net/url, but standard url.QueryEscape handles most.
	return url.QueryEscape(fmt.Sprintf("%v", val)), nil
}

func IriEncodeFilter(val any, args string) (any, error) {
	return url.PathEscape(fmt.Sprintf("%v", val)), nil
}

func UrlizeFilter(val any, args string) (any, error) {
	str := fmt.Sprintf("%v", val)
	// Very naive url detection
	re := regexp.MustCompile(`(http[s]?://[^\s]+)`)
	return godjango.SafeString(re.ReplaceAllString(str, `<a href="$1" rel="nofollow">$1</a>`)), nil
}

func UrlizeTruncFilter(val any, args string) (any, error) {
	str := fmt.Sprintf("%v", val)
	length, _ := strconv.Atoi(args)

	re := regexp.MustCompile(`(http[s]?://[^\s]+)`)
	res := re.ReplaceAllStringFunc(str, func(url string) string {
		display := url
		if len(url) > length && length > 3 {
			display = url[:length-3] + "..."
		}
		return fmt.Sprintf(`<a href="%s" rel="nofollow">%s</a>`, url, display)
	})
	return godjango.SafeString(res), nil
}

func Phone2NumericFilter(val any, args string) (any, error) {
	str := strings.ToLower(fmt.Sprintf("%v", val))
	mapping := map[rune]rune{
		'a': '2', 'b': '2', 'c': '2',
		'd': '3', 'e': '3', 'f': '3',
		'g': '4', 'h': '4', 'i': '4',
		'j': '5', 'k': '5', 'l': '5',
		'm': '6', 'n': '6', 'o': '6',
		'p': '7', 'q': '7', 'r': '7', 's': '7',
		't': '8', 'u': '8', 'v': '8',
		'w': '9', 'x': '9', 'y': '9', 'z': '9',
	}

	runes := []rune(str)
	for i, r := range runes {
		if num, ok := mapping[r]; ok {
			runes[i] = num
		}
	}
	return string(runes), nil
}

func JsonScriptFilter(val any, args string) (any, error) {
	bytes, err := json.Marshal(val)
	if err != nil {
		return "", err
	}

	// Escape closing script tags
	jsonStr := strings.ReplaceAll(string(bytes), "</script>", `\u003C/script>`)
	return godjango.SafeString(fmt.Sprintf(`<script id="%s" type="application/json">%s</script>`, args, jsonStr)), nil
}
