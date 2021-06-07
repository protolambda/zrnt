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
