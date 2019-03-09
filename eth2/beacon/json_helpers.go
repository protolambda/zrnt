package beacon

import (
	"encoding/hex"
)

func encodeHex(v []byte) ([]byte, error) {
	out := make([]byte, hex.EncodedLen(len(v)) + 2)
	out[0] = '"'
	hex.Encode(out[1:], v)
	out[len(out) - 1] = '"'
	return out, nil
}

func (v Root) MarshalJSON() ([]byte, error)         { return encodeHex(v[:]) }
func (v Bytes32) MarshalJSON() ([]byte, error)      { return encodeHex(v[:]) }
func (v BLSPubkey) MarshalJSON() ([]byte, error)    { return encodeHex(v[:]) }
func (v BLSSignature) MarshalJSON() ([]byte, error) { return encodeHex(v[:]) }
