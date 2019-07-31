package test_util

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strconv"
	"strings"
)

func DecodeHex(src []byte, dst []byte) error {
	offset, byteCount := DecodeHexOffsetAndLen(src)
	if len(dst) != byteCount {
		return errors.New("cannot decode hex, incorrect length")
	}
	_, err := hex.Decode(dst, src[offset:])
	return err
}

func DecodeHexOffsetAndLen(src []byte) (offset int, length int) {
	if src[0] == '0' && src[1] == 'x' {
		offset = 2
	}
	return offset, hex.DecodedLen(len(src) - offset)
}

func decodeHook(s reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t.Kind() == reflect.Slice && t.Elem().Kind() != reflect.Uint8 {
		return data, nil
	}
	if s.Kind() != reflect.String {
		return data, nil
	}
	strData := data.(string)
	if t.Kind() == reflect.Array && t.Elem().Kind() == reflect.Uint8 {
		res := reflect.New(t).Elem()
		sliceRes := res.Slice(0, t.Len()).Interface()
		err := DecodeHex([]byte(strData), sliceRes.([]byte))
		return res.Interface(), err
	}
	if t.Kind() == reflect.Uint64 {
		return strconv.ParseUint(strData, 10, 64)
	}
	if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		inBytes := []byte(strData)
		_, byteCount := DecodeHexOffsetAndLen(inBytes)
		res := make([]byte, byteCount, byteCount)
		err := DecodeHex([]byte(strData), res)
		return res, err
	}
	return data, nil
}

func UnmarshalYAML(unmarshal func(interface{}) error, getTyped func() interface{}) error {
	var raw interface{}
	// read raw YAML into parsed but untyped structure
	if err := unmarshal(&raw); err != nil {
		return err
	}

	typedData := getTyped()

	var md mapstructure.Metadata
	config := &mapstructure.DecoderConfig{
		DecodeHook:       decodeHook,
		Metadata:         &md,
		WeaklyTypedInput: false,
		ErrorUnused:      false,
		Result:           typedData,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	if err := decoder.Decode(raw); err != nil {
		return errors.New(fmt.Sprintf("cannot decode: %v", err))
	}
	if len(md.Unused) > 0 {
		return errors.New(fmt.Sprintf("unused keys: %s", strings.Join(md.Unused, ", ")))
	}
	return nil
}
