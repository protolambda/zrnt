package spec_testing

import (
	"regexp"
	"strings"
)

func decodeValue(v interface{}) interface{} {
	switch tv := v.(type) {
	case map[string]interface{}:
		return decodeStrKeyMap(tv)
	case map[interface{}]interface{}:
		return decodeMap(tv)
	case []interface{}:
		return decodeList(tv)
	default:
		return v
	}
}

var matchSnake = regexp.MustCompile("_([a-z0-9])")

func toPascalCase(str string) string {
	snake := []byte(matchSnake.ReplaceAllStringFunc(str, func(v string) string {
		return strings.ToUpper(string(v[1]))
	}))
	snake[0] = strings.ToUpper(string(snake[0]))[0]
	return string(snake)
}

func decodeList(v []interface{}) []interface{} {
	// return an explicit empty list for length 0 elements (may be nil slice)
	if len(v) == 0 {
		return make([]interface{}, 0)
	}
	items := len(v)
	out := make([]interface{}, 0, items)
	for i := 0; i < items; i++ {
		out = append(out, decodeValue(v[i]))
	}
	return out
}

func decodeStrKeyMap(v map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range v {
		out[toPascalCase(k)] = decodeValue(v)
	}
	return out
}

func decodeMap(v map[interface{}]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range v {
		kStr, ok := k.(string)
		if !ok {
			panic("cannot encode maps with non-string keys")
		}
		name := toPascalCase(kStr)
		out[name] = decodeValue(v)
	}
	return out
}
