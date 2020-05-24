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
type BLSDomain [32]byte

// A digest of the current fork data
type ForkDigest [4]byte

type ForkData struct {
	CurrentVersion        Version
	GenesisValidatorsRoot Root
}

var ForkDataSSZ = zssz.GetSSZ((*ForkData)(nil))

func ComputeForkDataRoot(currentVersion Version, genesisValidatorsRoot Root) Root {
	data := ForkData{
		CurrentVersion:        currentVersion,
		GenesisValidatorsRoot: genesisValidatorsRoot,
	}
	hFn := hashing.GetHashFn()
	return zssz.HashTreeRoot(htr.HashFn(hFn), &data, ForkDataSSZ)
}

func CompureForkDigest(currentVersion Version, genesisValidatorsRoot Root) ForkDigest {
	var digest ForkDigest
	dataRoot := ComputeForkDataRoot(currentVersion, genesisValidatorsRoot)
	copy(digest[:], dataRoot[:4])
	return digest
}

func ComputeDomain(domainType BLSDomainType, forkVersion Version, genesisValidatorsRoot Root) (out BLSDomain) {
	copy(out[0:4], domainType[:])
	forkDataRoot := ComputeForkDataRoot(forkVersion, genesisValidatorsRoot)
	copy(out[4:32], forkDataRoot[0:28])
	return
}

type SigningData struct {
	ObjectRoot Root
	Domain     BLSDomain
}

var SigningDataSSZ = zssz.GetSSZ((*SigningData)(nil))

func ComputeSigningRoot(msgRoot Root, dom BLSDomain) Root {
	withDomain := SigningData{
		ObjectRoot: msgRoot,
		Domain:     dom,
	}
	hFn := hashing.GetHashFn()
	return zssz.HashTreeRoot(htr.HashFn(hFn), &withDomain, SigningDataSSZ)
}
