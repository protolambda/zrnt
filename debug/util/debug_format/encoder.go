package debug_format

import (
	"encoding/hex"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"reflect"
	"regexp"
	"strings"
)

func Encode(v reflect.Value) interface{} {
	encoded := encodeValue(v)
	// add "hash_tree_root" to top level map
	if asOrdMap, ok := encoded.(OrderedMap); ok {
		coreHashTreeRoot := ssz.HashTreeRoot(v)
		encodedHashTreeRoot := make([]byte, hex.EncodedLen(len(coreHashTreeRoot)))
		hex.Encode(encodedHashTreeRoot, coreHashTreeRoot[:])
		encoded = append(asOrdMap, OrderedMapEntry{"hash_tree_root", string(encodedHashTreeRoot)})
	}
	return encoded
}

func encodeValue(v reflect.Value) interface{} {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		return encodeToMap(v)
	case reflect.Slice, reflect.Array:
		if v.Type().Elem().Kind() == reflect.Struct {
			return encodeToStructList(v)
		}
		// return an explicit empty list for length 0 elements (may be nil slice)
		if v.Len() == 0 {
			return make([]interface{}, 0)
		}
		return v.Interface()
	default:
		return v.Interface()
	}
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func encodeToStructList(v reflect.Value) []interface{} {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	items := v.Len()
	out := make([]interface{}, 0, items)
	for i := 0; i < items; i++ {
		out = append(out, encodeValue(v.Index(i)))
	}
	return out
}

func encodeToMap(v reflect.Value) OrderedMap {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	fields := v.NumField()
	out := make(OrderedMap, 0, fields)
	for i := 0; i < fields; i++ {
		f := v.Field(i)
		// ignore unexported struct fields
		if !f.CanSet() {
			continue
		}
		// Transform the name to the python formatting: snake case
		name := toSnakeCase(v.Type().Field(i).Name)
		out = append(out, OrderedMapEntry{name, encodeValue(f)})
		fieldHashTreeRoot := ssz.HashTreeRoot(f.Interface())
		encodedHashTreeRoot := make([]byte, hex.EncodedLen(len(fieldHashTreeRoot)))
		hex.Encode(encodedHashTreeRoot, fieldHashTreeRoot[:])
		out = append(out, OrderedMapEntry{name + "_hash_tree_root", string(encodedHashTreeRoot)})
	}
	return out
}
