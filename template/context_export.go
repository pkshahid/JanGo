package template

// Resolve allows external packages (like tags) to resolve variables using the private resolver.
func (c *Context) Resolve(path string) any {
	return resolveVariable(path, c)
}
