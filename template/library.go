package template

// TagParser is a function that parses a specific template tag.
type TagParser func(parser *Parser, token Token) (Node, error)

// FilterFunc is a function that filters a template variable.
type FilterFunc func(val any, args string) (any, error)

// Library represents a collection of tags and filters.
type Library struct {
	Tags    map[string]TagParser
	Filters map[string]FilterFunc
}

// NewLibrary creates a new Library.
func NewLibrary() *Library {
	return &Library{
		Tags:    make(map[string]TagParser),
		Filters: make(map[string]FilterFunc),
	}
}

// RegisterTag registers a tag parser.
func (l *Library) RegisterTag(name string, parser TagParser) {
	l.Tags[name] = parser
}

// RegisterFilter registers a filter function.
func (l *Library) RegisterFilter(name string, filter FilterFunc) {
	l.Filters[name] = filter
}
