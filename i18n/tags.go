package i18n

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkshahid/JanGo/template"
)

var Library *template.Library

func init() {
	Library = template.NewLibrary()
	Library.RegisterTag("trans", TransTag)
	Library.RegisterTag("blocktrans", BlockTransTag)
	Library.RegisterTag("get_current_language", GetCurrentLanguageTag)
	Library.RegisterTag("get_available_languages", GetAvailableLanguagesTag)
}

// unquote unquotes a template string literal.
func unquoteTemplateStr(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[0] == s[len(s)-1] {
		return s[1 : len(s)-1]
	}
	return s
}

// TransTag handles {% trans "message" %}
func TransTag(parser *template.Parser, token template.Token) (template.Node, error) {
	// Example: {% trans "Hello" %} or {% trans "Hello" as var %}
	args := strings.Fields(token.Contents)
	if len(args) < 2 {
		return nil, fmt.Errorf("trans tag requires at least one argument")
	}

	msg := unquoteTemplateStr(args[1])
	asVar := ""
	if len(args) >= 4 && args[2] == "as" {
		asVar = args[3]
	}

	return &TransNode{Message: msg, AsVar: asVar}, nil
}

type TransNode struct {
	Message string
	AsVar   string
}

func (n *TransNode) Render(ctx *template.Context) (string, error) {
	// Extract the Go context to get the active language.
	// We assume template.Context holds a reference to the request or a standard context.Context.
	// For this mock, we fetch context from request if available.
	var reqCtx context.Context
	reqRaw, ok := ctx.Get("request")
	if ok {
		// Assuming request has a Context() method or is a context itself
		if req, isReq := reqRaw.(interface{ Context() context.Context }); isReq {
			reqCtx = req.Context()
		}
	}

	translated := Gettext(reqCtx, n.Message)

	if n.AsVar != "" {
		ctx.Set(n.AsVar, translated)
		return "", nil
	}
	return translated, nil
}

// BlockTransTag handles {% blocktrans %}...{% plural %}...{% endblocktrans %}
func BlockTransTag(parser *template.Parser, token template.Token) (template.Node, error) {
	node := &BlockTransNode{}

	// Parse tokens until endblocktrans or plural
	singularNodes, err := parser.Parse([]string{"plural", "endblocktrans"})
	if err != nil {
		return nil, err
	}
	node.Singular = singularNodes

	nextToken := parser.NextToken()
	if nextToken != nil && nextToken.Contents == "plural" {
		pluralNodes, err := parser.Parse([]string{"endblocktrans"})
		if err != nil {
			return nil, err
		}
		node.Plural = pluralNodes
		parser.NextToken() // consume endblocktrans
	}

	// We'd parse arguments like count=... from the initial token here
	// For simplicity, we just stub this in this implementation
	return node, nil
}

type BlockTransNode struct {
	Singular template.NodeList
	Plural   template.NodeList
	CountVar string
}

func (n *BlockTransNode) Render(ctx *template.Context) (string, error) {
	var reqCtx context.Context
	reqRaw, ok := ctx.Get("request")
	if ok {
		if req, isReq := reqRaw.(interface{ Context() context.Context }); isReq {
			reqCtx = req.Context()
		}
	}

	// Render singular to get the message id
	// Real Django evaluates variables inside, but we simplify
	singularStr, _ := n.Singular.Render(ctx)

	if len(n.Plural) > 0 {
		pluralStr, _ := n.Plural.Render(ctx)
		count := 1 // Normally extracted via n.CountVar
		return Ngettext(reqCtx, singularStr, pluralStr, count), nil
	}

	return Gettext(reqCtx, singularStr), nil
}

// GetCurrentLanguageTag handles {% get_current_language as LANGUAGE_CODE %}
func GetCurrentLanguageTag(parser *template.Parser, token template.Token) (template.Node, error) {
	args := strings.Fields(token.Contents)
	if len(args) == 3 && args[1] == "as" {
		return &GetCurrentLanguageNode{AsVar: args[2]}, nil
	}
	return nil, fmt.Errorf("Invalid get_current_language syntax")
}

type GetCurrentLanguageNode struct {
	AsVar string
}

func (n *GetCurrentLanguageNode) Render(ctx *template.Context) (string, error) {
	var reqCtx context.Context
	reqRaw, ok := ctx.Get("request")
	if ok {
		if req, isReq := reqRaw.(interface{ Context() context.Context }); isReq {
			reqCtx = req.Context()
		}
	}
	lang := GetLanguage(reqCtx)
	ctx.Set(n.AsVar, lang)
	return "", nil
}

// GetAvailableLanguagesTag handles {% get_available_languages as LANGUAGES %}
func GetAvailableLanguagesTag(parser *template.Parser, token template.Token) (template.Node, error) {
	args := strings.Fields(token.Contents)
	if len(args) == 3 && args[1] == "as" {
		return &GetAvailableLanguagesNode{AsVar: args[2]}, nil
	}
	return nil, fmt.Errorf("Invalid get_available_languages syntax")
}

type GetAvailableLanguagesNode struct {
	AsVar string
}

func (n *GetAvailableLanguagesNode) Render(ctx *template.Context) (string, error) {
	ctx.Set(n.AsVar, Config.Languages)
	return "", nil
}
