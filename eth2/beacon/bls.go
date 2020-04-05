package beacon

import (
	"bytes"
	"github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zssz"
	"github.com/protolambda/zssz/htr"
	. "github.com/protolambda/ztyp/view"
)

type BLSPubkey [48]byte
var BLSPubkeyType = BasicVectorType(ByteType, 48)

type BLSSignature [96]byte
var BLSSignatureType = BasicVectorType(ByteType, 96)

type BLSPubkeyMessagePair struct {
	PK      BLSPubkey
	Message []byte
}

// Mixed into a BLS domain to define its type
type BLSDomainType [4]byte

// BLS domain (8 bytes): fork version (32 bits) concatenated with BLS domain type (32 bits)
type BLSDomain [32]byte

func ComputeDomain(domainType BLSDomainType, forkVersion Version, genesisValidatorsRoot Root) (out BLSDomain) {
	copy(out[0:4], domainType[:])
	forkDataRoot := ComputeForkDataRoot(forkVersion, genesisValidatorsRoot)
	copy(out[4:32], forkDataRoot[0:28])
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

// For pubkeys/signatures in state, a tree-representation is used. (TODO: cache optimized deserialized/parsed bls points)

type BLSPubkeyView struct {
	*BasicVectorView
}

func AsBLSPubkey(v View, err error) (BLSPubkey, error) {
	if err != nil {
		return BLSPubkey{}, err
	}
	bv, err := AsBasicVector(v, nil)
	if err != nil {
		return BLSPubkey{}, err
	}
	pub := BLSPubkeyView{BasicVectorView: bv}
	var out BLSPubkey
	buf := bytes.NewBuffer(out[:])
	if err := pub.Serialize(buf); err != nil {
		return BLSPubkey{}, nil
	}
	copy(out[:], buf.Bytes())
	return out, nil
}

type BLSSignatureView struct {
	*BasicVectorView
}

func AsBLSSignature(v View, err error) (BLSSignature, error) {
	if err != nil {
		return BLSSignature{}, err
	}
	bv, err := AsBasicVector(v, nil)
	if err != nil {
		return BLSSignature{}, err
	}
	pub := BLSSignatureView{BasicVectorView: bv}
	var out BLSSignature
	buf := bytes.NewBuffer(out[:])
	if err := pub.Serialize(buf); err != nil {
		return BLSSignature{}, nil
	}
	copy(out[:], buf.Bytes())
	return out, nil
}
