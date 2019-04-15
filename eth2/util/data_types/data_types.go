package data_types

import "github.com/protolambda/zrnt/eth2/util/hex"

type Root [32]byte

func (v *Root) MarshalJSON() ([]byte, error)    { return hex.EncodeHex((*v)[:]) }
func (v *Root) UnmarshalJSON(data []byte) error { return hex.DecodeHex(data[1:len(data)-1], (*v)[:]) }
func (v *Root) String() string                  { return hex.EncodeHexStr((*v)[:]) }
