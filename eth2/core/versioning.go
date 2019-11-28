package core

import (
	"fmt"
	. "github.com/protolambda/ztyp/view"
)

const VersionType = Bytes4

// 32 bits, not strictly an integer, hence represented as 4 bytes
// (bytes not necessarily corresponding to versions)
type Version [4]byte

func (v Version) ToUint32() uint32 {
	return uint32(v[0])<<24 | uint32(v[1])<<16 | uint32(v[2])<<8 | uint32(v[3])
}

type VersionProp ReadPropFn

func (p VersionProp) Version() (out Version, err error) {
	var v View
	v, err = p()
	if err != nil {
		return
	}
	b, ok := v.(SmallByteVecView)
	if !ok {
		err = fmt.Errorf("version is not a small byte vector: %v", b)
		return
	}
	copy(out[:], b)
	return
}
