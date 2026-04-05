package filters

import (
	"fmt"
	"strings"
	"time"

	godjango "github.com/godjango/godjango/template"
)

func RegisterDateTimeFilters(lib *godjango.Library) {
	lib.RegisterFilter("date", DateFilter)
	lib.RegisterFilter("time", TimeFilter)
	lib.RegisterFilter("timesince", TimeSinceFilter)
	lib.RegisterFilter("timeuntil", TimeUntilFilter)
	lib.RegisterFilter("filesizeformat", FileSizeFormatFilter)
	lib.RegisterFilter("pluralize", PluralizeFilter)
	lib.RegisterFilter("pprint", PprintFilter)
}

func parseTimeVal(val any) (time.Time, bool) {
	if t, ok := val.(time.Time); ok {
		return t, true
	}
	if t, ok := val.(*time.Time); ok && t != nil {
		return *t, true
	}
	return time.Time{}, false
}

func formatDjangoTime(t time.Time, format string) string {
	// Prototype map of common Django formats to Go format
	goFmt := format
	goFmt = strings.ReplaceAll(goFmt, "Y", "2006")
	goFmt = strings.ReplaceAll(goFmt, "y", "06")
	goFmt = strings.ReplaceAll(goFmt, "m", "01")
	goFmt = strings.ReplaceAll(goFmt, "n", "1")
	goFmt = strings.ReplaceAll(goFmt, "F", "January")
	goFmt = strings.ReplaceAll(goFmt, "M", "Jan")
	goFmt = strings.ReplaceAll(goFmt, "d", "02")
	goFmt = strings.ReplaceAll(goFmt, "j", "2")
	goFmt = strings.ReplaceAll(goFmt, "l", "Monday")
	goFmt = strings.ReplaceAll(goFmt, "D", "Mon")

	goFmt = strings.ReplaceAll(goFmt, "H", "15")
	goFmt = strings.ReplaceAll(goFmt, "G", "15") // no leading zero
	goFmt = strings.ReplaceAll(goFmt, "h", "03")
	goFmt = strings.ReplaceAll(goFmt, "g", "3")
	goFmt = strings.ReplaceAll(goFmt, "i", "04")
	goFmt = strings.ReplaceAll(goFmt, "s", "05")
	goFmt = strings.ReplaceAll(goFmt, "a", "pm")
	goFmt = strings.ReplaceAll(goFmt, "A", "PM")

	return t.Format(goFmt)
}

func DateFilter(val any, args string) (any, error) {
	if args == "" {
		args = "Y-m-d"
	}
	t, ok := parseTimeVal(val)
	if !ok {
		return "", nil
	}
	return formatDjangoTime(t, args), nil
}

func TimeFilter(val any, args string) (any, error) {
	if args == "" {
		args = "g:i a"
	}
	t, ok := parseTimeVal(val)
	if !ok {
		return "", nil
	}
	return formatDjangoTime(t, args), nil
}

func TimeSinceFilter(val any, args string) (any, error) {
	t, ok := parseTimeVal(val)
	if !ok {
		return "", nil
	}

	now := time.Now()
	// Django timesince allows a second argument to compare against
	// We'll skip parsing the second arg for this prototype.

	diff := now.Sub(t)
	if diff < 0 {
		return "0 minutes", nil
	}

	days := int(diff.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%d days", days), nil
	}
	hours := int(diff.Hours())
	if hours > 0 {
		return fmt.Sprintf("%d hours", hours), nil
	}
	minutes := int(diff.Minutes())
	return fmt.Sprintf("%d minutes", minutes), nil
}

func TimeUntilFilter(val any, args string) (any, error) {
	t, ok := parseTimeVal(val)
	if !ok {
		return "", nil
	}

	now := time.Now()
	diff := t.Sub(now)
	if diff < 0 {
		return "0 minutes", nil
	}

	days := int(diff.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%d days", days), nil
	}
	hours := int(diff.Hours())
	if hours > 0 {
		return fmt.Sprintf("%d hours", hours), nil
	}
	minutes := int(diff.Minutes())
	return fmt.Sprintf("%d minutes", minutes), nil
}

func FileSizeFormatFilter(val any, args string) (any, error) {
	var bytes int64
	fmt.Sscanf(fmt.Sprintf("%v", val), "%d", &bytes)

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d bytes", bytes), nil
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp]), nil
}

func PluralizeFilter(val any, args string) (any, error) {
	var count int

	slice := getSlice(val)
	if slice != nil {
		count = len(slice)
	} else {
		fmt.Sscanf(fmt.Sprintf("%v", val), "%d", &count)
	}

	parts := strings.Split(args, ",")
	singular := ""
	plural := "s"

	if len(parts) == 1 && parts[0] != "" {
		plural = parts[0]
	} else if len(parts) == 2 {
		singular = parts[0]
		plural = parts[1]
	}

	if count == 1 || count == -1 {
		return singular, nil
	}
	return plural, nil
}

func PprintFilter(val any, args string) (any, error) {
	return godjango.SafeString(fmt.Sprintf("<pre>%#v</pre>", val)), nil
}
