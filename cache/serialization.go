package cache

import (
	"encoding/json"
	"reflect"
	"time"
)

type serializedWrapper struct {
	TypeName string          `json:"type"`
	Data     json.RawMessage `json:"data"`
}

var typeRegistry = map[string]reflect.Type{
	"time.Time": reflect.TypeOf(time.Time{}),
}

// RegisterType allows registering a custom type for deserialization when stored as `any`.
func RegisterType(name string, v any) {
	typeRegistry[name] = reflect.TypeOf(v)
}

func getTypeName(v any) string {
	if v == nil {
		return ""
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.String()
}

// Serialize converts any value into a JSON byte slice, preserving type information if registered.
func Serialize(value any) ([]byte, error) {
	if value == nil {
		return json.Marshal(nil)
	}

	// For simple types, just marshal them directly if we don't need strict struct preservation.
	// But to be safe, we wrap structs.
	t := reflect.TypeOf(value)
	if t.Kind() == reflect.Struct || (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct) {
		typeName := getTypeName(value)
		if _, ok := typeRegistry[typeName]; ok {
			data, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			wrapper := serializedWrapper{
				TypeName: typeName,
				Data:     data,
			}
			return json.Marshal(wrapper)
		}
	}

	// Otherwise marshal directly
	return json.Marshal(value)
}

// Deserialize converts JSON bytes back into an interface{}.
// If it's a known wrapped type, it instantiates the correct struct.
func Deserialize(data []byte) (any, error) {
	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}

	// Try to unmarshal into wrapper first
	var wrapper serializedWrapper
	if err := json.Unmarshal(data, &wrapper); err == nil && wrapper.TypeName != "" {
		if t, ok := typeRegistry[wrapper.TypeName]; ok {
			val := reflect.New(t).Interface()
			if err := json.Unmarshal(wrapper.Data, val); err == nil {
				// Return the value itself (dereferenced)
				return reflect.ValueOf(val).Elem().Interface(), nil
			}
		}
	}

	// Fallback to standard generic unmarshal
	var res any
	err := json.Unmarshal(data, &res)
	return res, err
}
