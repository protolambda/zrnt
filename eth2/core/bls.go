package core

import (
	. "github.com/protolambda/ztyp/props"
	. "github.com/protolambda/ztyp/view"
)

type BLSPubkeyBytes [48]byte
var BLSPubkeyType = BasicVectorType(ByteType, 48)

type BLSPubkey struct {
	*BasicVectorView
}

func NewBLSPubkey() (b *BLSPubkey) {
	return &BLSPubkey{BasicVectorView: BLSPubkeyType.New()}
}

func (sig *BLSPubkey) Bytes() (out BLSPubkeyBytes) {
	_ = sig.IntoBytes(0, out[:])
	return
}

type BLSPubkeyReadProp BasicVectorReadProp

func (p BLSPubkeyReadProp) BLSPubkey() (*BLSPubkey, error) {
	if v, err := BasicVectorReadProp(p).BasicVector(); err != nil {
		return nil, err
	} else {
		return &BLSPubkey{BasicVectorView: v}, nil
	}
}


type BLSSignatureBytes [96]byte
var BLSSignatureType = BasicVectorType(ByteType, 96)

type BLSSignature struct {
	*BasicVectorView
}

func NewBLSSignature() (b *BLSSignature) {
	return &BLSSignature{BasicVectorView: BLSSignatureType.New()}
}

func (sig *BLSSignature) Bytes() (out BLSSignatureBytes) {
	_ = sig.IntoBytes(0, out[:])
	return
}

type BLSSignatureReadProp BasicVectorReadProp

func (p BLSSignatureReadProp) BLSSignature() (*BLSSignature, error) {
	if v, err := BasicVectorReadProp(p).BasicVector(); err != nil {
		return nil, err
	} else {
		return &BLSSignature{BasicVectorView: v}, nil
	}
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
