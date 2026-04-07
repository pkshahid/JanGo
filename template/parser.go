package template

import (
	"fmt"
	"strings"
)

// Parser builds an AST from tokens.
type Parser struct {
	tokens []Token
	pos    int
	engine *Engine
	tags   map[string]TagParser
}

// NewParser creates a new parser.
func NewParser(tokens []Token, engine *Engine) *Parser {
	p := &Parser{
		tokens: tokens,
		engine: engine,
		tags:   make(map[string]TagParser),
	}

	if engine != nil {
		for _, lib := range engine.Builtins {
			for name, parser := range lib.Tags {
				p.tags[name] = parser
			}
		}
	}

	return p
}

// NextToken consumes and returns the next token.
func (p *Parser) NextToken() *Token {
	if p.pos >= len(p.tokens) {
		return nil
	}
	tok := &p.tokens[p.pos]
	p.pos++
	return tok
}

// PrependToken puts a token back into the stream.
func (p *Parser) PrependToken(tok Token) {
	p.pos--
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

			// Check if we hit a stop token
			for _, stop := range stopAt {
				if tagName == stop {
					p.PrependToken(tok)
					return nodes, nil
				}
			}

			if tagParser, ok := p.tags[tagName]; ok {
				node, err := tagParser(p, tok)
				if err != nil {
					return nil, err
				}
				nodes = append(nodes, node)
			} else {
				// Fallback to legacy or unknown handling
				nodes = append(nodes, &BlockNode{
					TagName: tagName,
					TagArgs: tok.Contents,
				})
			}
		}
	}

	if len(stopAt) > 0 {
		return nil, fmt.Errorf("unclosed tag, expected one of: %v", stopAt)
	}

	return nodes, nil
}
