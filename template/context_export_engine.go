package template

// GetRenderState fetches internal engine state data.
func (c *Context) GetRenderState(key string) any {
	if c.renderState == nil {
		return nil
	}
	return c.renderState[key]
}

// SetRenderState saves internal engine state data.
func (c *Context) SetRenderState(key string, val any) {
	if c.renderState == nil {
		c.renderState = make(map[string]any)
	}
	c.renderState[key] = val
}

// GetEngine fetches the Engine pointer if attached.
func (c *Context) GetEngine() *Engine {
	if eng, ok := c.GetRenderState("engine").(*Engine); ok {
		return eng
	}
	return nil
}

// GetEngine fetches Engine attached to parser
func (p *Parser) GetEngine() *Engine {
	return p.engine
}

// RegisterTag allows a tag to register tags to the parser dynamically (e.g. {% load %})
func (p *Parser) RegisterTag(name string, parser TagParser) {
	p.tags[name] = parser
}
