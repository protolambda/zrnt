package core

import (
	"github.com/protolambda/zrnt/eth2/util/hex"
)

type Root [32]byte

func (v *Root) String() string { return hex.EncodeHexStr((*v)[:]) }

type Bytes []byte

func (v *Bytes) String() string { return hex.EncodeHexStr(*v) }
