package exceptions

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/godjango/godjango/core/settings"
	godjangohttp "github.com/godjango/godjango/http"
)

//go:embed templates/*
var templatesFS embed.FS

type Frame struct {
	Function string
	File     string
	Line     int
	Snippet  string
	Vars     map[string]string // We'll stub this since go doesn't allow easy runtime local variable inspection
}

type DebugErrorData struct {
	ExceptionType  string
	ExceptionValue string
	Path           string
	Traceback      string
	Frames         []Frame
	GET            map[string]string
	POST           map[string]string
	Headers        map[string]string
	Settings       map[string]string
}

func RenderDebug500(req *godjangohttp.Request, err interface{}, traceback string) godjangohttp.Response {
	var errType, errValue string

	switch e := err.(type) {
	case error:
		errType = reflect.TypeOf(e).String()
		errValue = e.Error()
	case string:
		errType = "string"
		errValue = e
	default:
		errType = reflect.TypeOf(e).String()
		errValue = "Unknown error"
	}
	errType = strings.TrimPrefix(errType, "*")

	getData := make(map[string]string)
	for k, v := range req.GET {
		if len(v) > 0 { getData[k] = v[0] }
	}

	postData := make(map[string]string)
	if req.POST != nil {
		for k, v := range req.POST {
			if len(v) > 0 {
				if strings.Contains(strings.ToLower(k), "password") || strings.Contains(strings.ToLower(k), "secret") {
					postData[k] = "********"
				} else {
					postData[k] = v[0]
				}
			}
		}
	}

	headerData := make(map[string]string)
	for k, v := range req.Header {
		if len(v) > 0 {
			if strings.EqualFold(k, "Authorization") || strings.EqualFold(k, "Cookie") {
				headerData[k] = "********"
			} else {
				headerData[k] = v[0]
			}
		}
	}

	settingsData := dumpSettingsMasked()
	frames := parseTraceback(traceback)

	data := DebugErrorData{
		ExceptionType:  errType,
		ExceptionValue: errValue,
		Path:           req.Path,
		Traceback:      traceback,
		Frames:         frames,
		GET:            getData,
		POST:           postData,
		Headers:        headerData,
		Settings:       settingsData,
	}

	tmpl, parseErr := template.ParseFS(templatesFS, "templates/debug_500.html")
	if parseErr != nil {
		return godjangohttp.NewHttpResponse("Server Error\n\n"+traceback, http.StatusInternalServerError)
	}

	var buf bytes.Buffer
	if exeErr := tmpl.Execute(&buf, data); exeErr != nil {
		return godjangohttp.NewHttpResponse("Server Error\n\n"+traceback, http.StatusInternalServerError)
	}

	resp := godjangohttp.NewHttpResponse(buf.String(), http.StatusInternalServerError)
	resp.Headers.Set("Content-Type", "text/html; charset=utf-8")
	return resp
}

func parseTraceback(traceback string) []Frame {
	lines := strings.Split(traceback, "\n")
	var frames []Frame

	// Start from 1 to skip "goroutine X [running]:"
	for i := 1; i < len(lines); i += 2 {
		if i+1 >= len(lines) {
			break
		}

		funcName := lines[i]
		fileLine := strings.TrimSpace(lines[i+1])

		// Expected format: /path/to/file.go:123 +0x456
		parts := strings.Split(fileLine, " ")
		if len(parts) == 0 {
			continue
		}

		fileLineParts := strings.Split(parts[0], ":")
		if len(fileLineParts) != 2 {
			continue
		}

		file := fileLineParts[0]
		lineStr := fileLineParts[1]
		lineNum, err := strconv.Atoi(lineStr)
		if err != nil {
			continue
		}

		snippet := extractSnippet(file, lineNum)

		// In Go we cannot natively reflect local variables from the stack frame at runtime
		// without DWARF debugging information (like delve uses).
		// We'll provide a placeholder message so the UI satisfies the requirement visually
		// while adhering to Go's technical limits.
		vars := map[string]string{
			"info": "Local variable introspection requires external debuggers (e.g. dlv) in Go.",
		}

		frames = append(frames, Frame{
			Function: funcName,
			File:     file,
			Line:     lineNum,
			Snippet:  snippet,
			Vars:     vars,
		})
	}
	return frames
}

func extractSnippet(file string, targetLine int) string {
	content, err := os.ReadFile(file)
	if err != nil {
		return "Source code not available."
	}

	lines := strings.Split(string(content), "\n")
	start := targetLine - 5
	if start < 0 { start = 0 }
	end := targetLine + 5
	if end > len(lines) { end = len(lines) }

	var sb strings.Builder
	for i := start; i < end; i++ {
		prefix := "  "
		if i+1 == targetLine {
			prefix = ">>"
		}
		sb.WriteString(fmt.Sprintf("%s %4d: %s\n", prefix, i+1, lines[i]))
	}
	return sb.String()
}

func dumpSettingsMasked() map[string]string {
	s := settings.Get()
	val := reflect.ValueOf(s).Elem()
	typ := val.Type()

	result := make(map[string]string)

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		name := field.Name

		if !field.IsExported() {
			continue
		}

		v := val.Field(i)
		vStr := fmt.Sprintf("%v", v.Interface())

		lowerName := strings.ToLower(name)
		if strings.Contains(lowerName, "secret") ||
		   strings.Contains(lowerName, "password") ||
		   strings.Contains(lowerName, "key") {
			vStr = "********************"
		}

		result[name] = vStr
	}

	return result
}
