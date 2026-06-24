package tags

import (
	"fmt"
	"strings"
	"time"

	godjango "github.com/pkshahid/JanGo/template"
)

// RegisterUtilTags registers with, comment, cycle, now, spaceless, verbatim, lorem.
func RegisterUtilTags(lib *godjango.Library) {
	lib.RegisterTag("with", WithParser)
	lib.RegisterTag("comment", CommentParser)
	lib.RegisterTag("cycle", CycleParser)
	lib.RegisterTag("now", NowParser)
	lib.RegisterTag("spaceless", SpacelessParser)
	lib.RegisterTag("verbatim", VerbatimParser)
	lib.RegisterTag("lorem", LoremParser)
}

func WithParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	parts := strings.Split(token.Contents, " ")
	kwargs := make(map[string]string)

	for _, p := range parts[1:] {
		if strings.Contains(p, "=") {
			kv := strings.SplitN(p, "=", 2)
			kwargs[kv[0]] = kv[1]
		}
	}

	children, err := parser.Parse([]string{"endwith"})
	if err != nil {
		return nil, err
	}
	parser.NextToken() // consume endwith

	return &WithNode{Kwargs: kwargs, Nodes: children}, nil
}

type WithNode struct {
	Kwargs map[string]string
	Nodes  godjango.NodeList
}

func (n *WithNode) Render(ctx *godjango.Context) (string, error) {
	ctx.Push(nil)
	defer ctx.Pop()

	for k, expr := range n.Kwargs {
		val := ctx.Resolve(expr)
		if val == "" || val == nil {
			ctx.Set(k, strings.Trim(expr, `"'`))
		} else {
			ctx.Set(k, val)
		}
	}

	return n.Nodes.Render(ctx)
}

func CommentParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	_, err := parser.Parse([]string{"endcomment"})
	if err != nil {
		return nil, err
	}
	parser.NextToken() // consume endcomment
	return &godjango.TextNode{Text: ""}, nil
}

func CycleParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	parts := strings.Split(token.Contents, " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("cycle requires arguments")
	}

	// Remove quotes
	var cycleValues []string
	for _, p := range parts[1:] {
		cycleValues = append(cycleValues, strings.Trim(p, `"'`))
	}

	// Memory allocation based pointer to isolate state globally for THIS specific AST node.
	return &CycleNode{Values: cycleValues, nodeID: fmt.Sprintf("cycle_%p", &cycleValues)}, nil
}

type CycleNode struct {
	Values []string
	nodeID string
}

func (n *CycleNode) Render(ctx *godjango.Context) (string, error) {
	if len(n.Values) == 0 {
		return "", nil
	}

	stateKey := "_cycle_state_" + n.nodeID
	idx := 0
	if saved, _ := ctx.GetRenderState(stateKey).(int); saved > 0 {
		idx = saved
	}

	val := n.Values[idx]
	ctx.SetRenderState(stateKey, (idx+1)%len(n.Values))
	return val, nil
}

func NowParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	parts := strings.SplitN(token.Contents, " ", 2)
	format := "Y-m-d" // Default pseudo format
	if len(parts) > 1 {
		format = strings.Trim(parts[1], `"'`)
	}
	return &NowNode{Format: format}, nil
}

type NowNode struct {
	Format string
}

func (n *NowNode) Render(ctx *godjango.Context) (string, error) {
	// Simple mapping of Django date format to Go format
	// This is NOT comprehensive, just a prototype mapping
	goFmt := n.Format
	goFmt = strings.ReplaceAll(goFmt, "Y", "2006")
	goFmt = strings.ReplaceAll(goFmt, "m", "01")
	goFmt = strings.ReplaceAll(goFmt, "d", "02")
	goFmt = strings.ReplaceAll(goFmt, "H", "15")
	goFmt = strings.ReplaceAll(goFmt, "i", "04")
	goFmt = strings.ReplaceAll(goFmt, "s", "05")

	return time.Now().Format(goFmt), nil
}

func SpacelessParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	children, err := parser.Parse([]string{"endspaceless"})
	if err != nil {
		return nil, err
	}
	parser.NextToken()
	return &SpacelessNode{Nodes: children}, nil
}

type SpacelessNode struct {
	Nodes godjango.NodeList
}

func (n *SpacelessNode) Render(ctx *godjango.Context) (string, error) {
	res, err := n.Nodes.Render(ctx)
	if err != nil {
		return "", err
	}

	// Strip spaces between HTML tags.
	// Since Go strings are immutable and doing character level scans is sometimes tricky
	// for a generic template that might output plain text and nodes, a simple split/trim works
	// for the prototype of spaceless.
	var buf strings.Builder

	for i := 0; i < len(res); i++ {
		if res[i] == '<' {
			buf.WriteByte(res[i])
		} else if res[i] == '>' {
			buf.WriteByte(res[i])
			// look ahead for next '<'
			j := i + 1
			onlyWhitespace := true
			for j < len(res) {
				if res[j] == '<' {
					break
				}
				if res[j] != ' ' && res[j] != '\n' && res[j] != '\t' && res[j] != '\r' {
					onlyWhitespace = false
					break
				}
				j++
			}
			if onlyWhitespace && j < len(res) && res[j] == '<' {
				// Skip the whitespace
				i = j - 1
			}
		} else {
			buf.WriteByte(res[i])
		}
	}

	return buf.String(), nil
}

func VerbatimParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	// Our Lexer doesn't natively support verbatim blocks, it just tokenizes everything.
	// To implement verbatim using AST, we reconstruct the tokens into text until we hit endverbatim.
	var buf strings.Builder

	for {
		tok := parser.NextToken()
		if tok == nil {
			return nil, fmt.Errorf("unclosed verbatim block")
		}

		if tok.Type == godjango.TokenBlock && strings.HasPrefix(tok.Contents, "endverbatim") {
			break
		}

		switch tok.Type {
		case godjango.TokenText:
			buf.WriteString(tok.Contents)
		case godjango.TokenVar:
			buf.WriteString("{{ ")
			buf.WriteString(tok.Contents)
			buf.WriteString(" }}")
		case godjango.TokenBlock:
			buf.WriteString("{% ")
			buf.WriteString(tok.Contents)
			buf.WriteString(" %}")
		case godjango.TokenComment:
			buf.WriteString("{# ")
			buf.WriteString(tok.Contents)
			buf.WriteString(" #}")
		}
	}

	return &godjango.TextNode{Text: buf.String()}, nil
}

func LoremParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	// Very simple implementation of lorem
	return &godjango.TextNode{Text: "Lorem ipsum dolor sit amet, consectetur adipiscing elit."}, nil
}
