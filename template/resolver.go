package template

import (
	"reflect"
	"strings"
)

// resolveVariable traverses a variable path (e.g., "user.profile.name") in the context.
func resolveVariable(path string, ctx *Context) any {
	// Handle strings literals directly
	if (strings.HasPrefix(path, "\"") && strings.HasSuffix(path, "\"")) ||
		(strings.HasPrefix(path, "'") && strings.HasSuffix(path, "'")) {
		return path[1 : len(path)-1]
	}

	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return ""
	}

	// Find the base variable in the context
	val, ok := ctx.Get(parts[0])
	if !ok {
		return ""
	}

	for _, part := range parts[1:] {
		val = resolvePart(val, part)
		if val == nil {
			return ""
		}
	}

	return val
}

func resolvePart(obj any, part string) any {
	if obj == nil {
		return nil
	}

	val := reflect.ValueOf(obj)

	// Dereference pointers
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	// 1. Try Map
	if val.Kind() == reflect.Map {
		key := reflect.ValueOf(part)
		if val.Type().Key() == key.Type() {
			res := val.MapIndex(key)
			if res.IsValid() {
				return res.Interface()
			}
		}
		return nil
	}

	// 2. Try Struct Field
	if val.Kind() == reflect.Struct {
		field := val.FieldByName(part)
		// If not found, try Title case for exported fields
		if !field.IsValid() {
			titlePart := strings.Title(part)
			field = val.FieldByName(titlePart)
		}

		if field.IsValid() && field.CanInterface() {
			return field.Interface()
		}
	}

	// 3. Try Method
	method := val.MethodByName(part)
	if !method.IsValid() {
		titlePart := strings.Title(part)
		method = val.MethodByName(titlePart)
	}

	// Only call methods with no arguments and a single return value (or value + error)
	if method.IsValid() && method.Type().NumIn() == 0 {
		numOut := method.Type().NumOut()
		if numOut == 1 || numOut == 2 {
			results := method.Call(nil)
			// Handle error if second return value
			if numOut == 2 && !results[1].IsNil() {
				return nil
			}
			return results[0].Interface()
		}
	}

	return nil
}
