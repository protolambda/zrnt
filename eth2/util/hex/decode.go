package hex

import (
	hexenc "encoding/hex"
	"errors"
)

func DecodeHex(src []byte, dst []byte) error {
	offset := 0
	if src[0] == '0' && src[1] == 'x' {
		offset = 2
	}
	byteCount := hexenc.DecodedLen(len(src) - offset)
	if len(dst) != byteCount {
		return errors.New("cannot decode hex, incorrect length")
	}
	_, err := hexenc.Decode(dst, src[offset:])
	return err
}
