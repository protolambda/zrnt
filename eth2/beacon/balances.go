package beacon

import (
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Balances []Gwei

func (a *Balances) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, Gwei(0))
		return &(*a)[i]
	}, GweiType.TypeByteLength(), spec.VALIDATOR_REGISTRY_LIMIT)
}

func (a Balances) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, GweiType.TypeByteLength(), uint64(len(a)))
}

func (a Balances) ByteLength(spec *Spec) (out uint64) {
	return uint64(len(a)) * GweiType.TypeByteLength()
}

func (a *Balances) FixedLength(spec *Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li Balances) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	length := uint64(len(li))
	return hFn.Uint64ListHTR(func(i uint64) uint64 {
		return uint64(li[i])
	}, length, spec.VALIDATOR_REGISTRY_LIMIT)
}

func (c *Phase0Config) RegistryBalances() *BasicListTypeDef {
	return BasicListType(GweiType, c.VALIDATOR_REGISTRY_LIMIT)
}

type RegistryBalancesView struct {
	*BasicListView
}

func AsRegistryBalances(v View, err error) (*RegistryBalancesView, error) {
	c, err := AsBasicList(v, err)
	return &RegistryBalancesView{c}, err
}

func (v *RegistryBalancesView) GetBalance(index ValidatorIndex) (Gwei, error) {
	return AsGwei(v.Get(uint64(index)))
}

func (v *RegistryBalancesView) SetBalance(index ValidatorIndex, bal Gwei) error {
	return v.Set(uint64(index), Uint64View(bal))
}

func (v *RegistryBalancesView) IncreaseBalance(index ValidatorIndex, delta Gwei) error {
	bal, err := v.GetBalance(index)
	if err != nil {
		return err
	}
	bal += delta
	return v.SetBalance(index, bal)
}

func (v *RegistryBalancesView) DecreaseBalance(index ValidatorIndex, delta Gwei) error {
	bal, err := v.GetBalance(index)
	if err != nil {
		return err
	}
	// prevent underflow, clip to 0
	if bal >= delta {
		bal -= delta
	} else {
		bal = 0
	}
	return v.SetBalance(index, bal)
}
