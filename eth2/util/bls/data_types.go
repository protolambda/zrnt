package bls

import "github.com/protolambda/zrnt/eth2/util/hex"

type BLSPubkey [48]byte
type BLSSignature [96]byte


func (v BLSPubkey) MarshalJSON() ([]byte, error)    { return hex.EncodeHex(v[:]) }
func (v BLSSignature) MarshalJSON() ([]byte, error) { return hex.EncodeHex(v[:]) }
func (v BLSPubkey) UnmarshalJSON(data []byte) error    { return hex.DecodeHex(data[1:len(data)-1], v[:]) }
func (v BLSSignature) UnmarshalJSON(data []byte) error { return hex.DecodeHex(data[1:len(data)-1], v[:]) }
func (v BLSPubkey) String() string    { return hex.EncodeHexStr(v[:]) }
func (v BLSSignature) String() string { return hex.EncodeHexStr(v[:]) }

type BLSDomain uint64
