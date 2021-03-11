package common

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Eth1Address [20]byte

func (p Eth1Address) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(p[:])), nil
}

func (p Eth1Address) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *Eth1Address) UnmarshalText(text []byte) error {
	if p == nil {
		return errors.New("cannot decode into nil Eth1Address")
	}
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	if len(text) != 40 {
		return fmt.Errorf("unexpected length string '%s'", string(text))
	}
	_, err := hex.Decode(p[:], text)
	return err
}

type Eth1Data struct {
	// Hash-tree-root of DepositData tree.
	DepositRoot  Root         `json:"deposit_root" yaml:"deposit_root"`
	DepositCount DepositIndex `json:"deposit_count" yaml:"deposit_count"`
	BlockHash    Root         `json:"block_hash" yaml:"block_hash"`
}

func (b *Eth1Data) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&b.DepositRoot, &b.DepositCount, &b.BlockHash)
}

func (a *Eth1Data) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(a.DepositRoot, a.DepositCount, a.BlockHash)
}

func (a *Eth1Data) ByteLength() uint64 {
	return Eth1DataType.TypeByteLength()
}

func (a *Eth1Data) FixedLength() uint64 {
	return Eth1DataType.TypeByteLength()
}

func (b *Eth1Data) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(b.DepositRoot, b.DepositCount, b.BlockHash)
}

func (dat *Eth1Data) View() *Eth1DataView {
	depRv := RootView(dat.DepositRoot)
	blockRv := RootView(dat.BlockHash)
	c, _ := Eth1DataType.FromFields(&depRv, Uint64View(dat.DepositCount), &blockRv)
	return &Eth1DataView{c}
}

var Eth1DataType = ContainerType("Eth1Data", []FieldDef{
	{"deposit_root", RootType},
	{"deposit_count", Uint64Type},
	{"block_hash", Bytes32Type},
})

type Eth1DataView struct{ *ContainerView }

func AsEth1Data(v View, err error) (*Eth1DataView, error) {
	c, err := AsContainer(v, err)
	return &Eth1DataView{c}, err
}

func (v *Eth1DataView) DepositRoot() (Root, error) {
	return AsRoot(v.Get(0))
}

func (v *Eth1DataView) SetDepositRoot(r Root) error {
	rv := RootView(r)
	return v.Set(0, &rv)
}

func (v *Eth1DataView) DepositCount() (DepositIndex, error) {
	return AsDepositIndex(v.Get(1))
}

func (v *Eth1DataView) DepositIndex() (DepositIndex, error) {
	return AsDepositIndex(v.Get(2))
}
