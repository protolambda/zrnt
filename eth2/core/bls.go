package core

import (
	"github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zssz"
	"github.com/protolambda/zssz/htr"
)

type BLSPubkey [48]byte

type BLSSignature [96]byte

type BLSPubkeyMessagePair struct {
	PK      BLSPubkey
	Message []byte
}

// Mixed into a BLS domain to define its type
type BLSDomainType [4]byte

// BLS domain (8 bytes): fork version (32 bits) concatenated with BLS domain type (32 bits)
type BLSDomain [8]byte

func ComputeDomain(domainType BLSDomainType, forkVersion Version) (out BLSDomain) {
	copy(out[0:4], domainType[:])
	copy(out[4:8], forkVersion[:])
	return
}

type SigningRoot struct {
	ObjectRoot Root
	Domain     BLSDomain
}

var SigningRootSSZ = zssz.GetSSZ((*SigningRoot)(nil))

func ComputeSigningRoot(msgRoot Root, dom BLSDomain) Root {
	withDomain := SigningRoot{
		ObjectRoot: msgRoot,
		Domain:     dom,
	}
	hFn := hashing.GetHashFn()
	return zssz.HashTreeRoot(htr.HashFn(hFn), &withDomain, SigningRootSSZ)
}
