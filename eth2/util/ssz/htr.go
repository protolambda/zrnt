package ssz

import (
	"github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zssz"
	"github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/types"
)

func HashTreeRoot(value interface{}, sszTyp types.SSZ) core.Root {
	hFn := hashing.GetHashFn()
	return zssz.HashTreeRoot(htr.HashFn(hFn), value, sszTyp)
}

func SigningRoot(value interface{}, sszTyp types.SSZ) core.Root {
	hFn := hashing.GetHashFn()
	signedSSZ, ok := sszTyp.(types.SignedSSZ)
	if !ok {
		// resort to Hash-tree-root, if the type is not something that can be signed
		return zssz.HashTreeRoot(htr.HashFn(hFn), value, sszTyp)
	}
	return zssz.SigningRoot(htr.HashFn(hFn), value, signedSSZ)
}
