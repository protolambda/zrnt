package beacon

import (
	"encoding/hex"
	"errors"
)

func encodeHexStr(v []byte) string {
	out := make([]byte, hex.EncodedLen(len(v))+4)
	out[0] = '"'
	out[1] = '0'
	out[2] = 'x'
	hex.Encode(out[1:], v)
	out[len(out)-1] = '"'
	return string(out)
}

// encodes with prefix
func encodeHex(v []byte) ([]byte, error) {
	return []byte(encodeHexStr(v)), nil
}

func decodeHex(src []byte, dst []byte) error {
	offset := 0
	if src[0] == '0' && src[1] == 'x' {
		offset = 2
	}
	byteCount := hex.DecodedLen(len(src) - offset)
	if len(dst) != byteCount {
		return errors.New("cannot decode hex, incorrect length")
	}
	_, err := hex.Decode(dst, src[offset:])
	return err
}

func (v Root) MarshalJSON() ([]byte, error)         { return encodeHex(v[:]) }
func (v Bytes32) MarshalJSON() ([]byte, error)      { return encodeHex(v[:]) }
func (v BLSPubkey) MarshalJSON() ([]byte, error)    { return encodeHex(v[:]) }
func (v BLSSignature) MarshalJSON() ([]byte, error) { return encodeHex(v[:]) }

func (v Root) UnmarshalJSON(data []byte) error         { return decodeHex(data, v[:]) }
func (v Bytes32) UnmarshalJSON(data []byte) error      { return decodeHex(data, v[:]) }
func (v BLSPubkey) UnmarshalJSON(data []byte) error    { return decodeHex(data, v[:]) }
func (v BLSSignature) UnmarshalJSON(data []byte) error { return decodeHex(data, v[:]) }

func (v Root) MarshalYAML() (interface{}, error)         { return encodeHexStr(v[:]), nil }
func (v Bytes32) MarshalYAML() (interface{}, error)      { return encodeHexStr(v[:]), nil }
func (v BLSPubkey) MarshalYAML() (interface{}, error)    { return encodeHexStr(v[:]), nil }
func (v BLSSignature) MarshalYAML() (interface{}, error) { return encodeHexStr(v[:]), nil }

func decodeYAMLStyle(v []byte, unmarshal func(interface{}) error) error {
	var data string
	if err := unmarshal(&data); err != nil {
		return err
	}
	return decodeHex([]byte(data), v[:])
}

func (v Root) UnmarshalYAML(unmarshal func(interface{}) error) error         { return decodeYAMLStyle(v[:], unmarshal) }
func (v Bytes32) UnmarshalYAML(unmarshal func(interface{}) error) error      { return decodeYAMLStyle(v[:], unmarshal) }
func (v BLSPubkey) UnmarshalYAML(unmarshal func(interface{}) error) error    { return decodeYAMLStyle(v[:], unmarshal) }
func (v BLSSignature) UnmarshalYAML(unmarshal func(interface{}) error) error { return decodeYAMLStyle(v[:], unmarshal) }

func (v Root) String() string         { return encodeHexStr(v[:]) }
func (v Bytes32) String() string      { return encodeHexStr(v[:]) }
func (v BLSPubkey) String() string    { return encodeHexStr(v[:]) }
func (v BLSSignature) String() string { return encodeHexStr(v[:]) }
