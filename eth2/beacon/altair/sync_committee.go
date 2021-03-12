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

// SyncCommitteeBits is formatted as a serialized SSZ bitvector,
// with trailing zero bits if length does not align with byte length.
type SyncCommitteeBits []byte

func (li SyncCommitteeBits) View(spec *common.Spec) *SyncCommitteeBitsView {
	v, _ := SyncCommitteeBitsType(spec).Deserialize(codec.NewDecodingReader(bytes.NewReader(li), uint64(len(li))))
	return &SyncCommitteeBitsView{v.(*BitVectorView)}
}

func (li *SyncCommitteeBits) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.BitList((*[]byte)(li), spec.SYNC_COMMITTEE_SIZE)
}

func (a SyncCommitteeBits) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.BitList(a[:])
}

func (a SyncCommitteeBits) ByteLength(spec *common.Spec) uint64 {
	return uint64(len(a))
}

func (a *SyncCommitteeBits) FixedLength(*common.Spec) uint64 {
	return 0
}

func (li SyncCommitteeBits) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.BitListHTR(li, spec.SYNC_COMMITTEE_SIZE)
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

func SyncCommitteePubkeysType(spec *common.Spec) VectorTypeDef {
	return VectorType(common.BLSPubkeyType, spec.SYNC_COMMITTEE_SIZE)
}

type SyncCommitteePubkeysView struct {
	*ComplexVectorView
}

func AsSyncCommitteePubkeys(v View, err error) (*SyncCommitteePubkeysView, error) {
	c, err := AsComplexVector(v, err)
	return &SyncCommitteePubkeysView{c}, err
}

func SyncCommitteePubkeyAggregatesType(spec *common.Spec) *ComplexVectorTypeDef {
	return ComplexVectorType(common.BLSPubkeyType, spec.SYNC_COMMITTEE_SIZE/spec.SYNC_SUBCOMMITTEE_SIZE)
}

type SyncCommitteePubkeyAggregatesView struct {
	*ComplexVectorView
}

func AsSyncCommitteePubkeyAggregates(v View, err error) (*SyncCommitteePubkeyAggregatesView, error) {
	c, err := AsComplexVector(v, err)
	return &SyncCommitteePubkeyAggregatesView{c}, err
}

func SyncCommitteeType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("SyncCommittee", []FieldDef{
		{"pubkeys", SyncCommitteePubkeysType(spec)},
		{"pubkey_aggregates", SyncCommitteePubkeyAggregatesType(spec)},
	})
}

type SyncCommitteeView struct {
	*ContainerView
}

func AsSyncCommittee(v View, err error) (*SyncCommitteeView, error) {
	c, err := AsContainer(v, err)
	return &SyncCommitteeView{c}, err
}
