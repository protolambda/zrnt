package debug_format

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type OrderedMapEntry struct {
	Key   string
	Value interface{}
}
type OrderedMap []OrderedMapEntry

func (ordMap OrderedMap) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	length := len(ordMap)
	count := 0
	for _, kv := range ordMap {
		jsonValue, err := json.Marshal(kv.Value)
		if err != nil {
			return nil, err
		}
		buffer.WriteString(fmt.Sprintf("\"%s\":%s", kv.Key, string(jsonValue)))
		count++
		if count < length {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("}")
	return buffer.Bytes(), nil
}
