package core

import (
	. "github.com/protolambda/ztyp/props"
	. "github.com/protolambda/ztyp/view"
)

type BLSPubkey [48]byte
var BLSPubkeyType = BasicVectorType(ByteType, 48)

type BLSSignature [96]byte
var BLSSignatureType = BasicVectorType(ByteType, 96)

// Mixed into a BLS domain to define its type
type BLSDomainType [4]byte

// BLS domain (8 bytes): fork version (32 bits) concatenated with BLS domain type (32 bits)
type BLSDomain [8]byte

func ComputeDomain(domainType BLSDomainType, forkVersion Version) (out BLSDomain) {
	copy(out[0:4], domainType[:])
	copy(out[4:8], forkVersion[:])
	return
}

// For pubkeys/signatures in state, a tree-representation is used.

type BLSPubkeyNode struct {
	*BasicVectorView
}

func NewBLSPubkeyNode() (b *BLSPubkeyNode) {
	return &BLSPubkeyNode{BasicVectorView: BLSPubkeyType.New()}
}

func (sig *BLSPubkeyNode) AsRaw() (out BLSPubkey) {
	_ = sig.IntoBytes(0, out[:])
	return
}

type BLSPubkeyReadProp BasicVectorReadProp

func (p BLSPubkeyReadProp) BLSPubkey() (*BLSPubkeyNode, error) {
	if v, err := BasicVectorReadProp(p).BasicVector(); err != nil {
		return nil, err
	} else {
		return &BLSPubkeyNode{BasicVectorView: v}, nil
	}
}


type BLSSignatureNode struct {
	*BasicVectorView
}

func NewBLSSignatureNode() (b *BLSSignatureNode) {
	return &BLSSignatureNode{BasicVectorView: BLSSignatureType.New()}
}

func (sig *BLSSignatureNode) AsRaw() (out BLSSignature) {
	_ = sig.IntoBytes(0, out[:])
	return
}

type BLSSignatureReadProp BasicVectorReadProp

func (p BLSSignatureReadProp) BLSSignature() (*BLSSignatureNode, error) {
	if v, err := BasicVectorReadProp(p).BasicVector(); err != nil {
		return nil, err
	} else {
		return &BLSSignatureNode{BasicVectorView: v}, nil
	}
}