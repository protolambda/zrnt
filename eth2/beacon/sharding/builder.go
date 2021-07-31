package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Builder struct {
	Pubkey common.BLSPubkey `json:"pubkey" yaml:"pubkey"`
}

func (v *Builder) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&v.Pubkey)
}

func (v *Builder) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&v.Pubkey)
}

func (a *Builder) ByteLength() uint64 {
	return BuilderType.TypeByteLength()
}

func (*Builder) FixedLength() uint64 {
	return BuilderType.TypeByteLength()
}

func (v *Builder) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&v.Pubkey)
}

func (v *Builder) View() *BuilderView {
	c, _ := BuilderType.FromFields(
		common.ViewPubkey(&v.Pubkey),
	)
	return &BuilderView{c}
}

var BuilderType = ContainerType("Builder", []FieldDef{
	{"pubkey", common.BLSPubkeyType},
})

const (
	_builderPubkey = iota
)

type BuilderView struct {
	*ContainerView
}

//var _ common.Builder = (*BuilderView)(nil)

func NewBuilderView() *BuilderView {
	return &BuilderView{ContainerView: BuilderType.New()}
}

func AsBuilder(v View, err error) (*BuilderView, error) {
	c, err := AsContainer(v, err)
	return &BuilderView{c}, err
}

func (v *BuilderView) Pubkey() (common.BLSPubkey, error) {
	return common.AsBLSPubkey(v.Get(_builderPubkey))
}
