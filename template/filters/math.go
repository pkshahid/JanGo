package filters

import (
	"fmt"
	"strconv"
	"strings"

	godjango "github.com/pkshahid/JanGo/template"
)

func RegisterMathLogicFilters(lib *godjango.Library) {
	lib.RegisterFilter("add", AddFilter)
	lib.RegisterFilter("divisibleby", DivisibleByFilter)
	lib.RegisterFilter("floatformat", FloatFormatFilter)
	lib.RegisterFilter("get_digit", GetDigitFilter)
	lib.RegisterFilter("default", DefaultFilter)
	lib.RegisterFilter("default_if_none", DefaultIfNoneFilter)
	lib.RegisterFilter("yesno", YesNoFilter)
}

func AddFilter(val any, args string) (any, error) {
	if val == nil {
		return args, nil
	}

	valStr := fmt.Sprintf("%v", val)

	// Try parsing as float
	valF, errV := strconv.ParseFloat(valStr, 64)
	argF, errA := strconv.ParseFloat(args, 64)

	if errV == nil && errA == nil {
		sum := valF + argF
		// If both are integers, return int string
		if float64(int(valF)) == valF && float64(int(argF)) == argF {
			return fmt.Sprintf("%d", int(sum)), nil
		}
		return fmt.Sprintf("%g", sum), nil
	}

	// If float parsing fails, concatenate strings
	return valStr + args, nil
}

func DivisibleByFilter(val any, args string) (any, error) {
	valStr := fmt.Sprintf("%v", val)
	valInt, errV := strconv.Atoi(valStr)
	argInt, errA := strconv.Atoi(args)

	if errV != nil || errA != nil || argInt == 0 {
		return false, nil
	}

	return valInt%argInt == 0, nil
}

func FloatFormatFilter(val any, args string) (any, error) {
	valStr := fmt.Sprintf("%v", val)
	valF, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return "", nil
	}

	places := -1
	if args != "" {
		if p, err := strconv.Atoi(args); err == nil {
			places = p
		}
	}

	if places == -1 {
		// Default: 1 decimal if unrounded, 0 if integer
		if float64(int(valF)) == valF {
			return fmt.Sprintf("%d", int(valF)), nil
		}
		return fmt.Sprintf("%.1f", valF), nil
	}

	if places < 0 {
		places = -places
		// Negative means only show if it has a decimal part
		if float64(int(valF)) == valF {
			return fmt.Sprintf("%d", int(valF)), nil
		}
	}

	format := fmt.Sprintf("%%.%df", places)
	return fmt.Sprintf(format, valF), nil
}

func GetDigitFilter(val any, args string) (any, error) {
	valStr := fmt.Sprintf("%v", val)
	_, err := strconv.Atoi(valStr)
	if err != nil {
		return valStr, nil
	}

	argInt, err := strconv.Atoi(args)
	if err != nil || argInt < 1 {
		return valStr, nil
	}

	if argInt > len(valStr) {
		return "0", nil
	}

	// 1-indexed from the right
	idx := len(valStr) - argInt
	return string(valStr[idx]), nil
}

func DefaultFilter(val any, args string) (any, error) {
	if val == nil || val == "" || val == false {
		return args, nil
	}
	// Check for zero-length slices/maps
	str := fmt.Sprintf("%v", val)
	if str == "[]" || str == "map[]" {
		return args, nil
	}
	return val, nil
}

func DefaultIfNoneFilter(val any, args string) (any, error) {
	if val == nil {
		return args, nil
	}
	return val, nil
}

func YesNoFilter(val any, args string) (any, error) {
	parts := strings.Split(args, ",")
	yes, no, maybe := "yes", "no", "maybe"
	if len(parts) > 0 {
		yes = parts[0]
	}
	if len(parts) > 1 {
		no = parts[1]
	}
	if len(parts) > 2 {
		maybe = parts[2]
	} else {
		maybe = no
	}

	if val == nil {
		return maybe, nil
	}

	var isTrue bool
	if b, ok := val.(bool); ok {
		isTrue = b
	} else if s, ok := val.(string); ok {
		isTrue = len(s) > 0
	} else {
		str := fmt.Sprintf("%v", val)
		isTrue = str != "0" && str != "" && str != "[]" && str != "map[]"
	}

	if isTrue {
		return yes, nil
	}
	return no, nil
}
