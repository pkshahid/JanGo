package template

// Context represents a stack of dictionaries for template rendering.
type Context struct {
	dicts []map[string]any
	// Internal render state
	renderState map[string]any
}

// NewContext creates a new context with an initial dictionary.
func NewContext(data map[string]any) *Context {
	if data == nil {
		data = make(map[string]any)
	}
	return &Context{
		dicts:       []map[string]any{data},
		renderState: make(map[string]any),
	}
}

// Push adds a new dictionary to the stack.
func (c *Context) Push(data map[string]any) {
	if data == nil {
		data = make(map[string]any)
	}
	c.dicts = append(c.dicts, data)
}

// Pop removes the top dictionary from the stack.
func (c *Context) Pop() map[string]any {
	if len(c.dicts) == 0 {
		return nil
	}
	top := c.dicts[len(c.dicts)-1]
	c.dicts = c.dicts[:len(c.dicts)-1]
	return top
}

// Get retrieves a value from the context, searching from top to bottom.
func (c *Context) Get(key string) (any, bool) {
	for i := len(c.dicts) - 1; i >= 0; i-- {
		if val, ok := c.dicts[i][key]; ok {
			return val, true
		}
	}
	return nil, false
}

// Set sets a value in the top dictionary of the stack.
func (c *Context) Set(key string, value any) {
	if len(c.dicts) == 0 {
		c.dicts = append(c.dicts, make(map[string]any))
	}
	c.dicts[len(c.dicts)-1][key] = value
}

// Flatten combines all dictionaries in the stack into a single map.
// Values higher in the stack take precedence.
func (c *Context) Flatten() map[string]any {
	flat := make(map[string]any)
	for i := 0; i < len(c.dicts); i++ {
		for k, v := range c.dicts[i] {
			flat[k] = v
		}
	}
	return flat
}
