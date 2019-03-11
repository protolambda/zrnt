package debug_format

import (
	"encoding/json"
	"reflect"
)

func MarshalJSON(v interface{}, indent string) ([]byte, error) {
	return json.MarshalIndent(Encode(reflect.ValueOf(v)), "", indent)
}
