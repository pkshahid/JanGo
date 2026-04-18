package tags

import (
	"fmt"
	"strings"

	godjango "github.com/pkshahid/JanGo/template"
)

// RegisterLogicTags registers if, for, etc into a library.
func RegisterLogicTags(lib *godjango.Library) {
	lib.RegisterTag("if", IfParser)
	lib.RegisterTag("for", ForParser)
}

func IfParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	parts := strings.SplitN(token.Contents, " ", 2)
	expression := ""
	if len(parts) > 1 {
		expression = parts[1]
	}

	children, err := parser.Parse([]string{"else", "elif", "endif"})
	if err != nil {
		return nil, err
	}

	node := &IfNode{
		Conditions: []IfCondition{{Expr: expression, Nodes: children}},
	}

	for {
		nextTok := parser.NextToken()
		if nextTok == nil {
			return nil, fmt.Errorf("unclosed if block")
		}

		tagParts := strings.SplitN(nextTok.Contents, " ", 2)
		tagName := tagParts[0]

		if tagName == "endif" {
			break
		} else if tagName == "elif" {
			elifExpr := ""
			if len(tagParts) > 1 {
				elifExpr = tagParts[1]
			}
			elifChildren, err := parser.Parse([]string{"else", "elif", "endif"})
			if err != nil {
				return nil, err
			}
			node.Conditions = append(node.Conditions, IfCondition{Expr: elifExpr, Nodes: elifChildren})
		} else if tagName == "else" {
			elseChildren, err := parser.Parse([]string{"endif"})
			if err != nil {
				return nil, err
			}
			node.ElseNodes = elseChildren

			// Must end with endif now
			endifTok := parser.NextToken()
			if endifTok == nil || !strings.HasPrefix(endifTok.Contents, "endif") {
				return nil, fmt.Errorf("expected endif after else")
			}
			break
		}
	}

	return node, nil
}

type IfCondition struct {
	Expr  string
	Nodes godjango.NodeList
}

type IfNode struct {
	Conditions []IfCondition
	ElseNodes  godjango.NodeList
}

func (n *IfNode) Render(ctx *godjango.Context) (string, error) {
	for _, cond := range n.Conditions {
		// Mock expression evaluation.
		// A full implementation parses boolean expressions with 'and', 'or', '==', etc.
		// We'll do a simple truthiness check on a single variable for prototype.
		// Basic truthiness / operator logic parser
		isTruthy := evaluateExpression(cond.Expr, ctx)
		if isTruthy {
			return cond.Nodes.Render(ctx)
		}
	}

	if n.ElseNodes != nil {
		return n.ElseNodes.Render(ctx)
	}

	return "", nil
}

func ForParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	parts := strings.SplitN(token.Contents, " ", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid for tag")
	}

	forArgs := strings.Split(parts[1], " in ")
	if len(forArgs) != 2 {
		return nil, fmt.Errorf("invalid for tag args")
	}

	itemName := strings.TrimSpace(forArgs[0])
	listName := strings.TrimSpace(forArgs[1])

	children, err := parser.Parse([]string{"empty", "endfor"})
	if err != nil {
		return nil, err
	}

	node := &ForNode{
		ItemName: itemName,
		ListName: listName,
		Children: children,
	}

	nextTok := parser.NextToken()
	if nextTok != nil && strings.HasPrefix(nextTok.Contents, "empty") {
		emptyChildren, err := parser.Parse([]string{"endfor"})
		if err != nil {
			return nil, err
		}
		node.EmptyNodes = emptyChildren
		parser.NextToken() // consume endfor
	}

	return node, nil
}

type ForNode struct {
	ItemName   string
	ListName   string
	Children   godjango.NodeList
	EmptyNodes godjango.NodeList
}

func (n *ForNode) Render(ctx *godjango.Context) (string, error) {
	val := ctx.Resolve(n.ListName)

	// Convert to slice
	var sliceVal []any
	if s, ok := val.([]string); ok {
		for _, v := range s { sliceVal = append(sliceVal, v) }
	} else if s, ok := val.([]any); ok {
		sliceVal = s
	} else if s, ok := val.([]int); ok {
		for _, v := range s { sliceVal = append(sliceVal, v) }
	}

	if len(sliceVal) == 0 && n.EmptyNodes != nil {
		return n.EmptyNodes.Render(ctx)
	}

	var buf strings.Builder
	ctx.Push(nil)
	defer ctx.Pop()

	length := len(sliceVal)
	for i, item := range sliceVal {
		ctx.Set(n.ItemName, item)

		// Get parent forloop
		var parentloop any
		if p, ok := ctx.Get("forloop"); ok {
			parentloop = p
		}

		// Set forloop variables
		forloop := map[string]any{
			"counter":    i + 1,
			"counter0":   i,
			"first":      i == 0,
			"last":       i == length-1,
			"parentloop": parentloop,
		}
		ctx.Set("forloop", forloop)

		res, err := n.Children.Render(ctx)
		if err != nil {
			return "", err
		}
		buf.WriteString(res)
	}

	return buf.String(), nil
}
