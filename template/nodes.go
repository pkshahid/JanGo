package template

import (
	"bytes"
	"fmt"
	"html"
	"strings"
)

// Node represents a parsed element of a template.
type Node interface {
	Render(ctx *Context) (string, error)
}

// NodeList is a collection of nodes.
type NodeList []Node

func (nl NodeList) Render(ctx *Context) (string, error) {
	var buf bytes.Buffer
	for _, n := range nl {
		res, err := n.Render(ctx)
		if err != nil {
			return "", err
		}
		buf.WriteString(res)
	}
	return buf.String(), nil
}

// TextNode renders static text.
type TextNode struct {
	Text string
}

func (n *TextNode) Render(ctx *Context) (string, error) {
	return n.Text, nil
}

// VariableNode renders a context variable, applying filters.
type VariableNode struct {
	Expression string
}

func (n *VariableNode) Render(ctx *Context) (string, error) {
	// Parse expression for variable and filters (e.g. `name|default:"Bob"|safe`)
	parts := strings.Split(n.Expression, "|")
	baseVar := strings.TrimSpace(parts[0])

	val := resolveVariable(baseVar, ctx)

	// Default missing variable is empty string (Django behavior unless configured)
	strVal := fmt.Sprintf("%v", val)
	if val == nil || val == "" {
		strVal = ""
	}

	safe := false

	// Apply filters sequentially
	for _, f := range parts[1:] {
		filterStr := strings.TrimSpace(f)
		filterName := filterStr
		var filterArg string

		if idx := strings.Index(filterStr, ":"); idx != -1 {
			filterName = filterStr[:idx]
			filterArg = filterStr[idx+1:]
			if strings.HasPrefix(filterArg, `"`) && strings.HasSuffix(filterArg, `"`) {
				filterArg = filterArg[1 : len(filterArg)-1]
			}
		}

		if filterName == "safe" {
			safe = true
		} else if filterName == "default" {
			if strVal == "" {
				strVal = filterArg
			}
		} else if filterName == "upper" {
			strVal = strings.ToUpper(strVal)
		} else if filterName == "lower" {
			strVal = strings.ToLower(strVal)
		}
		// Additional filters would be registered globally or via Libraries.
	}

	if !safe {
		strVal = html.EscapeString(strVal)
	}

	return strVal, nil
}

// BlockNode represents a built-in block like {% if %} or custom blocks.
type BlockNode struct {
	TagName  string
	TagArgs  string
	Children NodeList
	// Used for blocks that have alternative branches like {% else %}
	ElseChildren NodeList
}

func (n *BlockNode) Render(ctx *Context) (string, error) {
	switch n.TagName {
	case "if":
		val := resolveVariable(n.TagArgs, ctx)
		isTruthy := false
		if val != nil {
			if b, ok := val.(bool); ok {
				isTruthy = b
			} else if s, ok := val.(string); ok {
				isTruthy = len(s) > 0
			} else {
				isTruthy = true // simplifed truthy logic
			}
		}

		if isTruthy {
			return n.Children.Render(ctx)
		} else if n.ElseChildren != nil {
			return n.ElseChildren.Render(ctx)
		}
		return "", nil

	case "for":
		// Format: {% for item in items %}
		parts := strings.Split(n.TagArgs, " in ")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid for tag syntax: %s", n.TagArgs)
		}
		itemName := strings.TrimSpace(parts[0])
		listName := strings.TrimSpace(parts[1])

		val := resolveVariable(listName, ctx)
		if val == nil {
			return "", nil
		}

		var buf bytes.Buffer

		// Very simplified iteration over slices.
		// A full implementation would use reflect to iterate anything iterable.
		sliceVal := []any{}
		if s, ok := val.([]string); ok {
			for _, v := range s { sliceVal = append(sliceVal, v) }
		} else if s, ok := val.([]any); ok {
			sliceVal = s
		}

		ctx.Push(nil) // new scope
		defer ctx.Pop()

		for _, item := range sliceVal {
			ctx.Set(itemName, item)
			res, err := n.Children.Render(ctx)
			if err != nil {
				return "", err
			}
			buf.WriteString(res)
		}

		return buf.String(), nil

	default:
		// Unknown tag
		return "", nil
	}
}
