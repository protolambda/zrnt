package beacon

import (
	"encoding/hex"
	"errors"
)

func encodeHexStr(v []byte) string {
	out := make([]byte, hex.EncodedLen(len(v)))
	hex.Encode(out, v)
	return "0x"+string(out)
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

func (v Root) UnmarshalJSON(data []byte) error         { return decodeHex(data[1:len(data)-1], v[:]) }
func (v Bytes32) UnmarshalJSON(data []byte) error      { return decodeHex(data[1:len(data)-1], v[:]) }
func (v BLSPubkey) UnmarshalJSON(data []byte) error    { return decodeHex(data[1:len(data)-1], v[:]) }
func (v BLSSignature) UnmarshalJSON(data []byte) error { return decodeHex(data[1:len(data)-1], v[:]) }

func (v Root) String() string         { return encodeHexStr(v[:]) }
func (v Bytes32) String() string      { return encodeHexStr(v[:]) }
func (v BLSPubkey) String() string    { return encodeHexStr(v[:]) }
func (v BLSSignature) String() string { return encodeHexStr(v[:]) }
