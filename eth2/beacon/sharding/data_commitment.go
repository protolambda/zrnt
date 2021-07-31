package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type BLSCommitment = common.BLSPubkey

var BLSCommitmentType = common.BLSPubkeyType

var DataCommitmentType = ContainerType("DataCommitment", []FieldDef{
	{"point", BLSCommitmentType},
	{"length", Uint64Type},
})

type DataCommitmentView struct {
	*ContainerView
}

func AsDataCommitment(v View, err error) (*DataCommitmentView, error) {
	c, err := AsContainer(v, err)
	return &DataCommitmentView{c}, err
}

func (v *DataCommitmentView) Point() (BLSCommitment, error) {
	return common.AsBLSPubkey(v.Get(0))
}

func (v *DataCommitmentView) Length() (uint64, error) {
	l, err := AsUint64(v.Get(1))
	return uint64(l), err
}

func (v *DataCommitmentView) Raw() (*DataCommitment, error) {
	point, err := common.AsBLSPubkey(v.Get(0))
	if err != nil {
		return nil, err
	}
	length, err := AsUint64(v.Get(1))
	if err != nil {
		return nil, err
	}
	return &DataCommitment{Point: point, Length: length}, nil
}

type DataCommitment struct {
	// KZG10 commitment to the data
	Point BLSCommitment `json:"point" yaml:"point"`
	// Length of the data in samples
	Length Uint64View `json:"length" yaml:"length"`
}

func (d *DataCommitment) View() *DataCommitmentView {
	v, _ := AsDataCommitment(DataCommitmentType.FromFields(common.ViewPubkey(&d.Point), d.Length))
	return v
}

func (d *DataCommitment) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&d.Point, &d.Length)
}

func (d *DataCommitment) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&d.Point, &d.Length)
}

func (a *DataCommitment) ByteLength() uint64 {
	return DataCommitmentType.TypeByteLength()
}

func (a *DataCommitment) FixedLength() uint64 {
	return DataCommitmentType.TypeByteLength()
}

func (d *DataCommitment) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&d.Point, &d.Length)
}

var AttestedDataCommitmentType = ContainerType("AttestedDataCommitment", []FieldDef{
	{"commitment", DataCommitmentType},
	{"root", RootType},
	{"includer_index", common.ValidatorIndexType},
})

type AttestedDataCommitmentView struct {
	*ContainerView
}

func AsAttestedDataCommitment(v View, err error) (*AttestedDataCommitmentView, error) {
	c, err := AsContainer(v, err)
	return &AttestedDataCommitmentView{c}, err
}

func (v *AttestedDataCommitmentView) Commitment() (*DataCommitmentView, error) {
	return AsDataCommitment(v.Get(0))
}

func (v *AttestedDataCommitmentView) Root() (common.Root, error) {
	r, err := AsRoot(v.Get(1))
	return r, err
}

func (v *AttestedDataCommitmentView) IncluderIndex() (common.ValidatorIndex, error) {
	vi, err := common.AsValidatorIndex(v.Get(2))
	return vi, err
}

func (v *AttestedDataCommitmentView) Raw() (*AttestedDataCommitment, error) {
	cmt, err := AsDataCommitment(v.Get(0))
	if err != nil {
		return nil, err
	}
	rawCmt, err := cmt.Raw()
	if err != nil {
		return nil, err
	}
	r, err := AsRoot(v.Get(1))
	if err != nil {
		return nil, err
	}
	vi, err := common.AsValidatorIndex(v.Get(2))
	if err != nil {
		return nil, err
	}
	return &AttestedDataCommitment{Commitment: *rawCmt, Root: r, IncluderIndex: vi}, nil
}

type AttestedDataCommitment struct {
	// KZG10 commitment to the data, and length
	Commitment DataCommitment `json:"commitment" yaml:"commitment"`
	// hash_tree_root of the ShardBlobHeader (stored so that attestations can be checked against it)
	Root common.Root `json:"root" yaml:"root"`
	// The proposer who included the shard-header
	IncluderIndex common.ValidatorIndex `json:"includer_index" yaml:"includer_index"`
}

func (ad *AttestedDataCommitment) View() *DataCommitmentView {
	v, _ := AsDataCommitment(AttestedDataCommitmentType.FromFields(
		ad.Commitment.View(), (*RootView)(&ad.Root), Uint64View(ad.IncluderIndex)))
	return v
}

func (ad *AttestedDataCommitment) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&ad.Commitment, &ad.Root, &ad.IncluderIndex)
}

func (ad *AttestedDataCommitment) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&ad.Commitment, &ad.Root, &ad.IncluderIndex)
}

func (ad *AttestedDataCommitment) ByteLength() uint64 {
	return AttestedDataCommitmentType.TypeByteLength()
}

func (ad *AttestedDataCommitment) FixedLength() uint64 {
	return AttestedDataCommitmentType.TypeByteLength()
}

func (ad *AttestedDataCommitment) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&ad.Commitment, &ad.Root, &ad.IncluderIndex)
}
