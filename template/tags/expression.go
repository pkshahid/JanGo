package tags

import (
	godjango "github.com/pkshahid/JanGo/template"
	"strings"
)

func evaluateExpression(expr string, ctx *godjango.Context) bool {
	expr = strings.TrimSpace(expr)

	// Check logical AND/OR/NOT
	if strings.HasPrefix(expr, "not ") {
		return !evaluateExpression(expr[4:], ctx)
	}

	if parts := strings.Split(expr, " and "); len(parts) > 1 {
		for _, p := range parts {
			if !evaluateExpression(p, ctx) {
				return false
			}
		}
		return true
	}

	if parts := strings.Split(expr, " or "); len(parts) > 1 {
		for _, p := range parts {
			if evaluateExpression(p, ctx) {
				return true
			}
		}
		return false
	}

	// Check operators
	operators := []string{"==", "!=", "<=", ">=", "<", ">", " in ", " not in "}
	var op string
	var lhs, rhs string

	for _, o := range operators {
		if idx := strings.Index(expr, o); idx != -1 {
			op = strings.TrimSpace(o)
			lhs = strings.TrimSpace(expr[:idx])
			rhs = strings.TrimSpace(expr[idx+len(o):])
			break
		}
	}

	if op != "" {
		lVal := ctx.Resolve(lhs)
		rVal := ctx.Resolve(rhs)

		switch op {
		case "==":
			return lVal == rVal
		case "!=":
			return lVal != rVal
			// Note: complex operator logic omitted for brevity in prototype.
			// A full implementation would compare ints, floats, and strings properly.
		}
		return false
	}

	// Simple truthiness
	val := ctx.Resolve(expr)
	if val == nil {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	if s, ok := val.(string); ok {
		return len(s) > 0
	}
	return true
}
