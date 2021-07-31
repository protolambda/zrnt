package altair

import (
	"bytes"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/bitfields"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/conv"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

// SyncCommitteeSubnetBits is formatted as a serialized SSZ bitvector,
// with trailing zero bits if length does not align with byte length.
type SyncCommitteeSubnetBits []byte

func (li SyncCommitteeSubnetBits) View(spec *common.Spec) *SyncCommitteeSubnetBitsView {
	v, _ := SyncCommitteeSubnetBitsType(spec).Deserialize(codec.NewDecodingReader(bytes.NewReader(li), uint64(len(li))))
	return &SyncCommitteeSubnetBitsView{v.(*BitVectorView)}
}

func (li *SyncCommitteeSubnetBits) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.BitVector((*[]byte)(li), spec.SYNC_COMMITTEE_SIZE/common.SYNC_COMMITTEE_SUBNET_COUNT)
}

func (a SyncCommitteeSubnetBits) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.BitVector(a[:])
}

func (a SyncCommitteeSubnetBits) ByteLength(spec *common.Spec) uint64 {
	return (spec.SYNC_COMMITTEE_SIZE + 7) / 8
}

func (a *SyncCommitteeSubnetBits) FixedLength(spec *common.Spec) uint64 {
	return (spec.SYNC_COMMITTEE_SIZE + 7) / 8
}

func (li SyncCommitteeSubnetBits) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.BitVectorHTR(li)
}

func (cb SyncCommitteeSubnetBits) MarshalText() ([]byte, error) {
	return conv.BytesMarshalText(cb[:])
}

func (cb *SyncCommitteeSubnetBits) UnmarshalText(text []byte) error {
	return conv.DynamicBytesUnmarshalText((*[]byte)(cb), text)
}

func (cb SyncCommitteeSubnetBits) String() string {
	return conv.BytesString(cb[:])
}

func (cb SyncCommitteeSubnetBits) GetBit(i uint64) bool {
	return bitfields.GetBit(cb, i)
}

func (cb SyncCommitteeSubnetBits) SetBit(i uint64, v bool) {
	bitfields.SetBit(cb, i, v)
}

type SyncCommitteeSubnetBitsView struct {
	*BitVectorView
}

func AsSyncCommitteeSubnetBits(v View, err error) (*SyncCommitteeSubnetBitsView, error) {
	c, err := AsBitVector(v, err)
	return &SyncCommitteeSubnetBitsView{c}, err
}

func (v *SyncCommitteeSubnetBitsView) Raw(spec *common.Spec) (SyncCommitteeSubnetBits, error) {
	byteLen := int((spec.SYNC_COMMITTEE_SIZE/common.SYNC_COMMITTEE_SUBNET_COUNT + 7) / 8)
	var buf bytes.Buffer
	buf.Grow(byteLen)
	if err := v.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		return nil, err
	}
	out := SyncCommitteeSubnetBits(buf.Bytes())
	if len(out) != byteLen {
		return nil, fmt.Errorf("failed to convert sync committee tree bits view to raw bits")
	}
	return out, nil
}

func SyncCommitteeSubnetBitsType(spec *common.Spec) *BitVectorTypeDef {
	return BitVectorType(spec.SYNC_COMMITTEE_SIZE / common.SYNC_COMMITTEE_SUBNET_COUNT)
}

// SyncCommitteeBits is formatted as a serialized SSZ bitvector,
// with trailing zero bits if length does not align with byte length.
type SyncCommitteeBits []byte

func (li SyncCommitteeBits) View(spec *common.Spec) *SyncCommitteeBitsView {
	v, _ := SyncCommitteeBitsType(spec).Deserialize(codec.NewDecodingReader(bytes.NewReader(li), uint64(len(li))))
	return &SyncCommitteeBitsView{v.(*BitVectorView)}
}

func (li *SyncCommitteeBits) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.BitVector((*[]byte)(li), spec.SYNC_COMMITTEE_SIZE)
}

func (a SyncCommitteeBits) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.BitVector(a[:])
}

func (a SyncCommitteeBits) ByteLength(spec *common.Spec) uint64 {
	return (spec.SYNC_COMMITTEE_SIZE + 7) / 8
}

func (a *SyncCommitteeBits) FixedLength(spec *common.Spec) uint64 {
	return (spec.SYNC_COMMITTEE_SIZE + 7) / 8
}

func (li SyncCommitteeBits) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.BitVectorHTR(li)
}

func (cb SyncCommitteeBits) MarshalText() ([]byte, error) {
	return conv.BytesMarshalText(cb[:])
}

func (cb *SyncCommitteeBits) UnmarshalText(text []byte) error {
	return conv.DynamicBytesUnmarshalText((*[]byte)(cb), text)
}

func (cb SyncCommitteeBits) String() string {
	return conv.BytesString(cb[:])
}

func (cb SyncCommitteeBits) GetBit(i uint64) bool {
	return bitfields.GetBit(cb, i)
}

func (cb SyncCommitteeBits) SetBit(i uint64, v bool) {
	bitfields.SetBit(cb, i, v)
}

type SyncCommitteeBitsView struct {
	*BitVectorView
}

func AsSyncCommitteeBits(v View, err error) (*SyncCommitteeBitsView, error) {
	c, err := AsBitVector(v, err)
	return &SyncCommitteeBitsView{c}, err
}

func (v *SyncCommitteeBitsView) Raw(spec *common.Spec) (SyncCommitteeBits, error) {
	byteLen := int((spec.SYNC_COMMITTEE_SIZE + 7) / 8)
	var buf bytes.Buffer
	buf.Grow(byteLen)
	if err := v.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		return nil, err
	}
	out := SyncCommitteeBits(buf.Bytes())
	if len(out) != byteLen {
		return nil, fmt.Errorf("failed to convert sync committee tree bits view to raw bits")
	}
	return out, nil
}

func SyncCommitteeBitsType(spec *common.Spec) *BitVectorTypeDef {
	return BitVectorType(spec.SYNC_COMMITTEE_SIZE)
}
