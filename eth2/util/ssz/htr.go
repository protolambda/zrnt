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

// When the hash function changed, also re-initialize the precomputed zero-hashes with this hash-function.
// These precomputed hashes are used to complete merkle-trees efficiently to a power of 2,
// without unnecessary duplicate hashing of zeroes, or hashes of, or higher order, up to 32.
func InitZeroHashes(hashFn hashing.HashFn) {
	htr.InitZeroHashes(htr.HashFn(hashFn))
}
