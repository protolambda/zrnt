package ssz

import (
	"github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zssz"
	"github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/types"
)

// HashTreeRoot the given value.
//
// The value is expected to be a pointer to a type matching the SSZ definition that is used.
// Example:
//  sszTyp := zssz.GetSSZ((*MyStruct)(nil))
//  value := MyStruct{A: 123, B: false, C: []byte{42,13,37}}
//  root := HashTreeRoot(&value, sszTyp)
func HashTreeRoot(value interface{}, sszTyp types.SSZ) core.Root {
	hFn := hashing.GetHashFn()
	return zssz.HashTreeRoot(htr.HashFn(hFn), value, sszTyp)
}

// SigningRoot the given value. If sszTyp is not a container,
// this function will just return the result of a regular HashTreeRoot.
//
// The value is expected to be a pointer to a type matching the SSZ definition that is used.
// Example:
//  sszTyp := zssz.GetSSZ((*MyStruct)(nil))
//  value := MyStruct{A: 123, B: false, C: []byte{42,13,37}}
//  root := SigningRoot(&value, sszTyp)
func SigningRoot(value interface{}, sszTyp types.SSZ) core.Root {
	hFn := hashing.GetHashFn()
	signedSSZ, ok := sszTyp.(types.SignedSSZ)
	if !ok {
		// resort to Hash-tree-root, if the type is not something that can be signed
		return zssz.HashTreeRoot(htr.HashFn(hFn), value, sszTyp)
	}
	return zssz.SigningRoot(htr.HashFn(hFn), value, signedSSZ)
}
