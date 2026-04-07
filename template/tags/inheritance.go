package tags

import (
	"fmt"
	"strings"

	godjango "github.com/godjango/godjango/template"
)

// RegisterInheritanceTags registers block, extends, include.
func RegisterInheritanceTags(lib *godjango.Library) {
	lib.RegisterTag("block", BlockParser)
	lib.RegisterTag("extends", ExtendsParser)
	lib.RegisterTag("include", IncludeParser)
	lib.RegisterTag("load", LoadParser)
}

func BlockParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	parts := strings.Split(token.Contents, " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid block tag syntax")
	}
	blockName := parts[1]

	children, err := parser.Parse([]string{"endblock"})
	if err != nil {
		return nil, err
	}
	parser.NextToken() // consume endblock

	return &BlockInheritNode{Name: blockName, Nodes: children}, nil
}

type BlockInheritNode struct {
	Name  string
	Nodes godjango.NodeList
}

func (n *BlockInheritNode) Render(ctx *godjango.Context) (string, error) {
	// If a child template overwrote this block, render that instead.
	// We track blocks in context.renderState["blocks"]
	blocksMap, _ := ctx.GetRenderState("blocks").(map[string]godjango.NodeList)
	if blocksMap != nil {
		if overrideNodes, ok := blocksMap[n.Name]; ok {
			// Provide block.super to the context
			superContent, _ := n.Nodes.Render(ctx)
			// Push a new context layer for block.super so it's accessible via {{ block.super }}
			ctx.Push(map[string]any{
				"block": map[string]any{
					"super": godjango.SafeString(superContent),
				},
			})
			res, err := overrideNodes.Render(ctx)
			ctx.Pop()
			return res, err
		}
	}
	return n.Nodes.Render(ctx)
}

func ExtendsParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	parts := strings.Split(token.Contents, " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid extends tag syntax")
	}
	parentName := strings.Trim(parts[1], `"'`)

	// Rest of the template nodes are either blocks or ignored text
	children, err := parser.Parse(nil)
	if err != nil {
		return nil, err
	}

	return &ExtendsNode{ParentName: parentName, Nodes: children}, nil
}

type ExtendsNode struct {
	ParentName string
	Nodes      godjango.NodeList
}

func (n *ExtendsNode) Render(ctx *godjango.Context) (string, error) {
	// Collect child blocks
	blocks := make(map[string]godjango.NodeList)
	for _, node := range n.Nodes {
		if blockNode, ok := node.(*BlockInheritNode); ok {
			blocks[blockNode.Name] = blockNode.Nodes
		}
	}

	// Merge into context state
	existingBlocks, _ := ctx.GetRenderState("blocks").(map[string]godjango.NodeList)
	if existingBlocks == nil {
		existingBlocks = make(map[string]godjango.NodeList)
	}
	for k, v := range blocks {
		if _, exists := existingBlocks[k]; !exists {
			existingBlocks[k] = v
		}
	}
	ctx.SetRenderState("blocks", existingBlocks)

	// Fetch and render parent
	engine := ctx.GetEngine()
	if engine == nil {
		return "", fmt.Errorf("extends requires template engine in context")
	}
	parentTmpl, err := engine.GetTemplate(n.ParentName)
	if err != nil {
		return "", err
	}

	return parentTmpl.Render(ctx)
}

func IncludeParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	parts := strings.Split(token.Contents, " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid include tag syntax")
	}
	tmplName := strings.Trim(parts[1], `"'`)

	// Parse 'with' kwargs if present
	kwargs := make(map[string]string)
	only := false

	for i := 2; i < len(parts); i++ {
		if parts[i] == "with" {
			continue
		} else if parts[i] == "only" {
			only = true
		} else if strings.Contains(parts[i], "=") {
			kv := strings.SplitN(parts[i], "=", 2)
			kwargs[kv[0]] = kv[1]
		}
	}

	return &IncludeNode{TemplateName: tmplName, Kwargs: kwargs, Only: only}, nil
}

type IncludeNode struct {
	TemplateName string
	Kwargs       map[string]string
	Only         bool
}

func (n *IncludeNode) Render(ctx *godjango.Context) (string, error) {
	engine := ctx.GetEngine()
	if engine == nil {
		return "", fmt.Errorf("include requires template engine")
	}

	tmpl, err := engine.GetTemplate(n.TemplateName)
	if err != nil {
		return "", err
	}

	var incCtx *godjango.Context
	if n.Only {
		incCtx = godjango.NewContext(nil)
	} else {
		// Clone top level state or just push
		ctx.Push(nil)
		incCtx = ctx
	}

	// Resolve and apply kwargs
	for k, expr := range n.Kwargs {
		incCtx.Set(k, ctx.Resolve(expr))
	}

	res, err := tmpl.Render(incCtx)

	if !n.Only {
		ctx.Pop()
	}

	return res, err
}

func LoadParser(parser *godjango.Parser, token godjango.Token) (godjango.Node, error) {
	// Django load just loads tag libraries. Since our engine parser resolves tags during Parse,
	// a load tag inside the template would need to inject parsers into the current Parser.
	// Let's implement dynamic library loading here.

	parts := strings.Split(token.Contents, " ")
	engine := parser.GetEngine()

	for _, libName := range parts[1:] {
		if lib, ok := engine.Libraries[libName]; ok {
			for name, tagParser := range lib.Tags {
				parser.RegisterTag(name, tagParser)
			}
		}
	}

	// Doesn't render anything
	return &godjango.TextNode{Text: ""}, nil
}
