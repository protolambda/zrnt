package hex

import (
	hexenc "encoding/hex"
)

func EncodeHexStr(v []byte) string {
	out := make([]byte, hexenc.EncodedLen(len(v)))
	hexenc.Encode(out, v)
	return "0x"+string(out)
}

// encodes with prefix
func EncodeHex(v []byte) ([]byte, error) {
	return []byte(EncodeHexStr(v)), nil
}
