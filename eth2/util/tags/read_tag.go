package tags

import (
	"reflect"
	"strings"
)

// HasFlag checks if the given struct-field is tagged
// with the given flag (comma separated), in the given tag namespace.
// Example:
// 	MyField: bool `foobar:"abc,def"`
// Namespace: "foobar", flags: "abc", "def"
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
