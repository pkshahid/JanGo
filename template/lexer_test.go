package template

import (
	"testing"
)

func TestLexer(t *testing.T) {
	input := `Hello {{ name }}!
{% if user %}
	{# This is a comment #}
	<p>Welcome, {{ user.name }}</p>
{% endif %}`

	lexer := NewLexer(input)
	tokens := lexer.Lex()

	expected := []struct {
		Type     TokenType
		Contents string
	}{
		{TokenText, "Hello "},
		{TokenVar, "name"},
		{TokenText, "!\n"},
		{TokenBlock, "if user"},
		{TokenText, "\n\t"},
		{TokenComment, "This is a comment"},
		{TokenText, "\n\t<p>Welcome, "},
		{TokenVar, "user.name"},
		{TokenText, "</p>\n"},
		{TokenBlock, "endif"},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.Type {
			t.Errorf("Token %d: Expected type %v, got %v", i, exp.Type, tokens[i].Type)
		}
		if tokens[i].Contents != exp.Contents {
			t.Errorf("Token %d: Expected contents %q, got %q", i, exp.Contents, tokens[i].Contents)
		}
	}
}
