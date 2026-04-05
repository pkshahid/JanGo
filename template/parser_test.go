package template

import (
	"testing"
)

func TestParserAndRendering(t *testing.T) {
	input := `Hello {{ name }}!
{% if user %}
	Welcome, {{ user.name }}
{% else %}
	Please log in.
{% endif %}

Items:
{% for item in items %}
- {{ item }}
{% endfor %}
`
	lexer := NewLexer(input)
	tokens := lexer.Lex()

	parser := NewParser(tokens)
	nodes, err := parser.Parse(nil)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	ctx := NewContext(map[string]any{
		"name": "Alice",
		"user": map[string]any{"name": "Alice"},
		"items": []any{"A", "B"},
	})

	result, err := nodes.Render(ctx)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	expected := `Hello Alice!

	Welcome, Alice


Items:

- A

- B

`
	if result != expected {
		t.Errorf("Render mismatch.\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestFiltersAndEscaping(t *testing.T) {
	input := `{{ val1|upper }}
{{ val2|default:"N/A" }}
{{ htmlVar }}
{{ htmlVar|safe }}`

	lexer := NewLexer(input)
	parser := NewParser(lexer.Lex())
	nodes, _ := parser.Parse(nil)

	ctx := NewContext(map[string]any{
		"val1": "test",
		"val2": "",
		"htmlVar": "<script>",
	})

	result, _ := nodes.Render(ctx)
	expected := `TEST
N/A
&lt;script&gt;
<script>`

	if result != expected {
		t.Errorf("Filter mismatch.\nExpected:\n%q\nGot:\n%q", expected, result)
	}
}
