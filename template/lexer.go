package template

import (
	"strings"
)

// TokenType represents the type of a template token.
type TokenType int

const (
	TokenText TokenType = iota
	TokenVar
	TokenBlock
	TokenComment
)

// Token represents a lexical token in a template.
type Token struct {
	Type     TokenType
	Contents string
	Line     int
}

// Lexer breaks a template string into tokens.
type Lexer struct {
	input  string
	pos    int
	line   int
	Tokens []Token
}

// NewLexer creates a new lexer.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
		line:  1,
	}
}

// Lex scans the input and produces tokens.
func (l *Lexer) Lex() []Token {
	for l.pos < len(l.input) {
		startVar := strings.Index(l.input[l.pos:], "{{")
		startBlock := strings.Index(l.input[l.pos:], "{%")
		startComment := strings.Index(l.input[l.pos:], "{#")

		// Find the closest tag
		start := -1
		tagType := TokenText

		findClosest := func(idx int, t TokenType) {
			if idx != -1 {
				idx += l.pos
				if start == -1 || idx < start {
					start = idx
					tagType = t
				}
			}
		}

		findClosest(startVar, TokenVar)
		findClosest(startBlock, TokenBlock)
		findClosest(startComment, TokenComment)

		if start == -1 {
			// No more tags, the rest is text
			text := l.input[l.pos:]
			l.addToken(TokenText, text)
			break
		}

		// Add preceding text
		if start > l.pos {
			text := l.input[l.pos:start]
			l.addToken(TokenText, text)
		}

		// Find end of tag
		var endStr string
		switch tagType {
		case TokenVar:
			endStr = "}}"
		case TokenBlock:
			endStr = "%}"
		case TokenComment:
			endStr = "#}"
		}

		end := strings.Index(l.input[start+2:], endStr)
		if end == -1 {
			// Unclosed tag, treat rest as text
			l.addToken(TokenText, l.input[start:])
			break
		}

		end += start + 2
		contents := strings.TrimSpace(l.input[start+2 : end])
		l.addToken(tagType, contents)
		l.pos = end + 2
	}
	return l.Tokens
}

func (l *Lexer) addToken(t TokenType, contents string) {
	if t == TokenText && contents == "" {
		return
	}
	l.Tokens = append(l.Tokens, Token{
		Type:     t,
		Contents: contents,
		Line:     l.line,
	})
	l.line += strings.Count(contents, "\n")
}
