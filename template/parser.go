package template

import (
	"fmt"
	"strings"
)

// Parser builds an AST from tokens.
type Parser struct {
	tokens []Token
	pos    int
}

// NewParser creates a new parser.
func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

// Parse runs the parser until it runs out of tokens or hits a stop token.
func (p *Parser) Parse(stopAt []string) (NodeList, error) {
	var nodes NodeList

	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		p.pos++

		switch tok.Type {
		case TokenText:
			nodes = append(nodes, &TextNode{Text: tok.Contents})
		case TokenVar:
			nodes = append(nodes, &VariableNode{Expression: tok.Contents})
		case TokenComment:
			// Ignore
		case TokenBlock:
			parts := strings.SplitN(tok.Contents, " ", 2)
			tagName := parts[0]
			tagArgs := ""
			if len(parts) > 1 {
				tagArgs = parts[1]
			}

			// Check if we hit a stop token (e.g. {% endif %})
			for _, stop := range stopAt {
				if tagName == stop {
					p.pos-- // step back so the caller can process it if needed
					return nodes, nil
				}
			}

			// Handle built-in blocks with bodies
			if tagName == "if" {
				children, err := p.Parse([]string{"else", "endif"})
				if err != nil {
					return nil, err
				}

				var elseChildren NodeList
				// Look at the stop token
				if p.pos < len(p.tokens) && p.tokens[p.pos].Contents == "else" {
					p.pos++ // consume else
					elseChildren, err = p.Parse([]string{"endif"})
					if err != nil {
						return nil, err
					}
				}

				if p.pos < len(p.tokens) && p.tokens[p.pos].Contents == "endif" {
					p.pos++ // consume endif
				} else {
					return nil, fmt.Errorf("unclosed if tag")
				}

				nodes = append(nodes, &BlockNode{
					TagName:      tagName,
					TagArgs:      tagArgs,
					Children:     children,
					ElseChildren: elseChildren,
				})
			} else if tagName == "for" {
				children, err := p.Parse([]string{"endfor"})
				if err != nil {
					return nil, err
				}

				if p.pos < len(p.tokens) && p.tokens[p.pos].Contents == "endfor" {
					p.pos++
				} else {
					return nil, fmt.Errorf("unclosed for tag")
				}

				nodes = append(nodes, &BlockNode{
					TagName:  tagName,
					TagArgs:  tagArgs,
					Children: children,
				})
			} else {
				// Ignore unknown tags for now or treat as simple block node
			}
		}
	}

	if len(stopAt) > 0 {
		return nil, fmt.Errorf("unclosed tag, expected one of: %v", stopAt)
	}

	return nodes, nil
}
