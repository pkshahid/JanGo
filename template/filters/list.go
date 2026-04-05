package filters

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"

	godjango "github.com/godjango/godjango/template"
)

func RegisterListSeqFilters(lib *godjango.Library) {
	lib.RegisterFilter("dictsort", DictSortFilter)
	lib.RegisterFilter("dictsortreversed", DictSortReversedFilter)
	lib.RegisterFilter("first", FirstFilter)
	lib.RegisterFilter("last", LastFilter)
	lib.RegisterFilter("random", RandomFilter)
	lib.RegisterFilter("join", JoinFilter)
	lib.RegisterFilter("slice", SliceFilter)
	lib.RegisterFilter("make_list", MakeListFilter)
	lib.RegisterFilter("unordered_list", UnorderedListFilter)
	lib.RegisterFilter("length", LengthFilter)
	lib.RegisterFilter("length_is", LengthIsFilter)
	lib.RegisterFilter("safeseq", SafeSeqFilter)
}

func getSlice(val any) []any {
	v := reflect.ValueOf(val)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() { return nil }
		v = v.Elem()
	}

	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		res := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			res[i] = v.Index(i).Interface()
		}
		return res
	}
	return nil
}

func getMapKeyValue(m any, key string) string {
	v := reflect.ValueOf(m)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() { return "" }
		v = v.Elem()
	}
	if v.Kind() == reflect.Map {
		val := v.MapIndex(reflect.ValueOf(key))
		if val.IsValid() {
			return fmt.Sprintf("%v", val.Interface())
		}
	} else if v.Kind() == reflect.Struct {
		field := v.FieldByName(key)
		if !field.IsValid() {
			field = v.FieldByName(strings.Title(key))
		}
		if field.IsValid() && field.CanInterface() {
			return fmt.Sprintf("%v", field.Interface())
		}
	}
	return ""
}

func DictSortFilter(val any, args string) (any, error) {
	slice := getSlice(val)
	if slice == nil {
		return val, nil
	}

	sort.SliceStable(slice, func(i, j int) bool {
		v1 := getMapKeyValue(slice[i], args)
		v2 := getMapKeyValue(slice[j], args)
		return v1 < v2
	})
	return slice, nil
}

func DictSortReversedFilter(val any, args string) (any, error) {
	slice := getSlice(val)
	if slice == nil {
		return val, nil
	}

	sort.SliceStable(slice, func(i, j int) bool {
		v1 := getMapKeyValue(slice[i], args)
		v2 := getMapKeyValue(slice[j], args)
		return v1 > v2
	})
	return slice, nil
}

func FirstFilter(val any, args string) (any, error) {
	slice := getSlice(val)
	if len(slice) > 0 {
		return slice[0], nil
	}
	return "", nil
}

func LastFilter(val any, args string) (any, error) {
	slice := getSlice(val)
	if len(slice) > 0 {
		return slice[len(slice)-1], nil
	}
	return "", nil
}

func RandomFilter(val any, args string) (any, error) {
	slice := getSlice(val)
	if len(slice) > 0 {
		idx := rand.Intn(len(slice))
		return slice[idx], nil
	}
	return "", nil
}

func JoinFilter(val any, args string) (any, error) {
	slice := getSlice(val)
	if slice == nil {
		return val, nil
	}
	var strs []string
	for _, v := range slice {
		strs = append(strs, fmt.Sprintf("%v", v))
	}
	return strings.Join(strs, args), nil
}

func SliceFilter(val any, args string) (any, error) {
	// args = "start:end"
	slice := getSlice(val)
	if slice == nil {
		// handle string
		str := fmt.Sprintf("%v", val)
		runes := []rune(str)
		start, end := parseSliceArgs(args, len(runes))
		if start < 0 || start > len(runes) || end < 0 || end > len(runes) || start > end {
			return str, nil // bounds error
		}
		return string(runes[start:end]), nil
	}

	start, end := parseSliceArgs(args, len(slice))
	if start < 0 || start > len(slice) || end < 0 || end > len(slice) || start > end {
		return slice, nil
	}
	return slice[start:end], nil
}

func parseSliceArgs(args string, length int) (int, int) {
	parts := strings.Split(args, ":")
	start, end := 0, length
	if len(parts) > 0 && parts[0] != "" {
		fmt.Sscanf(parts[0], "%d", &start)
	}
	if len(parts) > 1 && parts[1] != "" {
		fmt.Sscanf(parts[1], "%d", &end)
	}

	if start < 0 { start = length + start }
	if end < 0 { end = length + end }
	if start < 0 { start = 0 }
	if end < 0 { end = 0 }

	return start, end
}

func MakeListFilter(val any, args string) (any, error) {
	str := fmt.Sprintf("%v", val)
	var res []string
	for _, r := range []rune(str) {
		res = append(res, string(r))
	}
	return res, nil
}

func UnorderedListFilter(val any, args string) (any, error) {
	// Simplistic mapping for recursive lists.
	// A full implementation requires tracking nesting.
	slice := getSlice(val)
	if len(slice) == 0 {
		return "", nil
	}
	var buf strings.Builder
	for _, v := range slice {
		buf.WriteString(fmt.Sprintf("\t<li>%v</li>\n", v))
	}
	return godjango.SafeString(buf.String()), nil
}

func LengthFilter(val any, args string) (any, error) {
	slice := getSlice(val)
	if slice != nil {
		return len(slice), nil
	}
	return len(fmt.Sprintf("%v", val)), nil
}

func LengthIsFilter(val any, args string) (any, error) {
	length, _ := LengthFilter(val, "")
	lInt := length.(int)

	argInt := -1
	fmt.Sscanf(args, "%d", &argInt)
	return lInt == argInt, nil
}

func SafeSeqFilter(val any, args string) (any, error) {
	slice := getSlice(val)
	if slice == nil {
		return godjango.SafeString(fmt.Sprintf("%v", val)), nil
	}
	for i, v := range slice {
		slice[i] = godjango.SafeString(fmt.Sprintf("%v", v))
	}
	return slice, nil
}
