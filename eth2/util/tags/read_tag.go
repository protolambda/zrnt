package tags

import (
	"reflect"
	"strings"
)

func HasFlag(vt *reflect.StructField, namespace string, flag string) bool {
	if flag == "" {
		return false
	}
	tag, ok := vt.Tag.Lookup(namespace)
	if !ok {
		return false
	}
	if len(tag) == 0 {
		return false
	}

	// look through the tag to find a flag (comma separated)
	for tag != "" {
		var next string
		i := strings.Index(tag, ",")
		if i >= 0 {
			tag, next = tag[:i], tag[i+1:]
		}
		if tag == flag {
			return true
		}
		tag = next
	}

	return false
}
