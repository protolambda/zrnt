package common

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type CachedPubkey = bls.CachedPubkey

type BLSPubkey = bls.BLSPubkey

func ViewPubkey(pub *BLSPubkey) *BLSPubkeyView {
	v, _ := BLSPubkeyType.Deserialize(codec.NewDecodingReader(bytes.NewReader(pub[:]), 48))
	return &BLSPubkeyView{v.(*BasicVectorView)}
}

var BLSPubkeyType = BasicVectorType(ByteType, 48)

type BLSSignature = bls.BLSSignature

func ViewSignature(sig *BLSSignature) *BLSSignatureView {
	v, _ := BLSSignatureType.Deserialize(codec.NewDecodingReader(bytes.NewReader(sig[:]), 48))
	return &BLSSignatureView{v.(*BasicVectorView)}
}

var BLSSignatureType = BasicVectorType(ByteType, 96)

// Mixed into a BLS domain to define its type
type BLSDomainType [4]byte

func (dt BLSDomainType) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(dt[:])), nil
}

func (dt BLSDomainType) String() string {
	return "0x" + hex.EncodeToString(dt[:])
}

func (dt *BLSDomainType) Deserialize(dr *codec.DecodingReader) error {
	_, err := dr.Read(dt[:])
	return err
}

func (dt *BLSDomainType) Serialize(w *codec.EncodingWriter) error {
	return w.Write(dt[:])
}

func (dt *BLSDomainType) ByteLength() uint64 {
	return 4
}

func (dt *BLSDomainType) FixedLength() uint64 {
	return 4
}

func (dt BLSDomainType) HashTreeRoot(hFn tree.HashFn) Root {
	var out Root
	copy(out[:4], dt[:])
	return out
}

func (dt *BLSDomainType) UnmarshalText(text []byte) error {
	if dt == nil {
		return errors.New("cannot decode into nil BLSDomainType")
	}
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	if len(text) != 8 {
		return fmt.Errorf("unexpected length string '%s'", string(text))
	}
	_, err := hex.Decode(dt[:], text)
	return err
}

// Sometimes a beacon state is not available, or too much for what it is good for.
// Functions that just need a specific BLS domain can use this function.
type BLSDomainFn func(typ BLSDomainType, epoch Epoch) (BLSDomain, error)

// BLS domain (8 bytes): fork version (32 bits) concatenated with BLS domain type (32 bits)
type BLSDomain [32]byte

func (dom *BLSDomain) Deserialize(dr *codec.DecodingReader) error {
	_, err := dr.Read(dom[:])
	return err
}

func (dom *BLSDomain) Serialize(w *codec.EncodingWriter) error {
	return w.Write(dom[:])
}

func (a *BLSDomain) ByteLength() uint64 {
	return 32
}

func (a *BLSDomain) FixedLength() uint64 {
	return 32
}

func (dom BLSDomain) HashTreeRoot(hFn tree.HashFn) Root {
	return Root(dom) // just convert to root type (no hashing involved)
}

func (dom BLSDomain) String() string {
	return "0x" + hex.EncodeToString(dom[:])
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

func (d *SigningData) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&d.ObjectRoot, &d.Domain)
}

func (d *SigningData) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&d.ObjectRoot, &d.Domain)
}

func (a *SigningData) ByteLength() uint64 {
	return 32 + 32
}

func (a *SigningData) FixedLength() uint64 {
	return 32 + 32
}

func (d *SigningData) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(d.ObjectRoot, d.Domain)
}

func ComputeSigningRoot(msgRoot Root, dom BLSDomain) Root {
	withDomain := SigningData{
		ObjectRoot: msgRoot,
		Domain:     dom,
	}
	return withDomain.HashTreeRoot(tree.GetHashFn())
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
	buf := bytes.NewBuffer(out[:0])
	if err := pub.Serialize(codec.NewEncodingWriter(buf)); err != nil {
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
	buf := bytes.NewBuffer(out[:0])
	if err := pub.Serialize(codec.NewEncodingWriter(buf)); err != nil {
		return BLSSignature{}, nil
	}
	copy(out[:], buf.Bytes())
	return out, nil
}

// TODO: BLSPoint can be customized to have more bls-specific functionality and checks (e.g. modulus, not full 256 bits)
type BLSPoint = Root

const BLSPointType = RootType
