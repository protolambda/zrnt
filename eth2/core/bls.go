package core

import (
	"fmt"
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

type BLSPubkeyReadProp ReadPropFn

func (p BLSPubkeyReadProp) BLSPubkey() (*BLSPubkey, error) {
	v, err := p()
	if err != nil {
		return nil, err
	}
	pub, ok := v.(*BasicVectorView)
	if ok {
		return nil, fmt.Errorf("not a uint64 view: %v", v)
	}
	return &BLSPubkey{BasicVectorView: pub}, nil
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

// Mixed into a BLS domain to define its type
type BLSDomainType [4]byte

// BLS domain (8 bytes): fork version (32 bits) concatenated with BLS domain type (32 bits)
type BLSDomain [8]byte

func ComputeDomain(domainType BLSDomainType, forkVersion Version) (out BLSDomain) {
	copy(out[0:4], domainType[:])
	copy(out[4:8], forkVersion[:])
	return
}
