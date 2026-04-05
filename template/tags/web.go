package tags

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/godjango/godjango/core/settings"
	godjangohttp "github.com/godjango/godjango/http"
	"github.com/godjango/godjango/http/urls"
	godjango "github.com/godjango/godjango/template"
)

// RegisterWebTags registers url, static, csrf_token.
func RegisterWebTags(lib *godjango.Library) {
	lib.RegisterTag("url", UrlParser)
	lib.RegisterTag("static", StaticParser)
	lib.RegisterTag("csrf_token", CsrfTokenParser)
}

func UrlParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	parts := strings.Split(token.Contents, " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("url tag requires view name")
	}

	viewName := strings.Trim(parts[1], `"'`)

	kwargs := make(map[string]string)
	for _, p := range parts[2:] {
		if strings.Contains(p, "=") {
			kv := strings.SplitN(p, "=", 2)
			kwargs[kv[0]] = kv[1]
		}
	}

	return &UrlNode{ViewName: viewName, Kwargs: kwargs}, nil
}

type UrlNode struct {
	ViewName string
	Kwargs   map[string]string
}

func (n *UrlNode) Render(ctx *godjango.Context) (string, error) {
	resolvedKwargs := make(map[string]any)
	for k, v := range n.Kwargs {
		val := ctx.Resolve(v)
		// If resolution fails or returns empty, might be a string literal, but ctx.Resolve handles quotes
		if val == "" || val == nil {
			resolvedKwargs[k] = v // fallback to original string
		} else {
			resolvedKwargs[k] = val
		}
	}

	urlStr := urls.Reverse(n.ViewName, resolvedKwargs)
	if urlStr == "" {
		return "", fmt.Errorf("Reverse for '%s' not found", n.ViewName)
	}
	return urlStr, nil
}

func StaticParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	parts := strings.Split(token.Contents, " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("static tag requires a path")
	}

	path := parts[1] // The resolver handles quoted strings

	return &StaticNode{PathExpr: path}, nil
}

type StaticNode struct {
	PathExpr string
}

func (n *StaticNode) Render(ctx *godjango.Context) (string, error) {
	s := settings.Get()
	staticUrl := s.STATIC_URL
	if staticUrl == "" {
		staticUrl = "/static/"
	}

	pathVal := ctx.Resolve(n.PathExpr)
	pathStr := fmt.Sprintf("%v", pathVal)

	return staticUrl + pathStr, nil
}

func CsrfTokenParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	return &CsrfTokenNode{}, nil
}

type CsrfTokenNode struct{}

func (n *CsrfTokenNode) Render(ctx *godjango.Context) (string, error) {
	// Need to get the request from context
	val, ok := ctx.Get("request")
	if !ok {
		return "", nil // csrf_token requires request
	}

	req, ok := val.(*godjangohttp.Request)
	if !ok {
		return "", nil
	}

	token := req.META["CSRF_TOKEN"]
	if token == "" {
		return "", nil
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf(`<input type="hidden" name="csrfmiddlewaretoken" value="%s">`, token))
	return buf.String(), nil
}
