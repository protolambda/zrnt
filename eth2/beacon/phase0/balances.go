package phase0

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Balances []common.Gwei

func (a *Balances) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, common.Gwei(0))
		return &(*a)[i]
	}, common.GweiType.TypeByteLength(), spec.VALIDATOR_REGISTRY_LIMIT)
}

func (a Balances) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, common.GweiType.TypeByteLength(), uint64(len(a)))
}

func (a Balances) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(len(a)) * common.GweiType.TypeByteLength()
}

func (a *Balances) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li Balances) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.Uint64ListHTR(func(i uint64) uint64 {
		return uint64(li[i])
	}, length, spec.VALIDATOR_REGISTRY_LIMIT)
}

func RegistryBalancesType(spec *common.Spec) *BasicListTypeDef {
	return BasicListType(common.GweiType, spec.VALIDATOR_REGISTRY_LIMIT)
}

type RegistryBalancesView struct {
	*BasicListView
}

func AsRegistryBalances(v View, err error) (*RegistryBalancesView, error) {
	c, err := AsBasicList(v, err)
	return &RegistryBalancesView{c}, err
}

func (v *RegistryBalancesView) GetBalance(index common.ValidatorIndex) (common.Gwei, error) {
	return common.AsGwei(v.Get(uint64(index)))
}

func (v *RegistryBalancesView) SetBalance(index common.ValidatorIndex, bal common.Gwei) error {
	return v.Set(uint64(index), Uint64View(bal))
}

func (v *RegistryBalancesView) IncreaseBalance(index common.ValidatorIndex, delta common.Gwei) error {
	bal, err := v.GetBalance(index)
	if err != nil {
		return err
	}
	bal += delta
	return v.SetBalance(index, bal)
}

func (v *RegistryBalancesView) DecreaseBalance(index common.ValidatorIndex, delta common.Gwei) error {
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

func (v *RegistryBalancesView) AllBalances() ([]common.Gwei, error) {
	var out []common.Gwei
	balIter := v.ReadonlyIter()
	for i := common.ValidatorIndex(0); true; i++ {
		el, ok, err := balIter.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		balance, err := common.AsGwei(el, nil)
		if err != nil {
			return nil, err
		}
		out = append(out, balance)
	}
	return out, nil
}
