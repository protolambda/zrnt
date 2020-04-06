package beacon

import (
	"errors"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

type Balances []Gwei

func (_ *Balances) Limit() uint64 {
	return VALIDATOR_REGISTRY_LIMIT
}

var RegistryBalancesSSZ = zssz.GetSSZ((*Balances)(nil))

var RegistryBalancesType = BasicListType(GweiType, VALIDATOR_REGISTRY_LIMIT)

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
	return v.SetBalance(index, delta)
}

func (state *BeaconStateView) ApplyDeltas(deltas *Deltas) error {
	balances, err := state.Balances()
	if err != nil {
		return err
	}
	balLen, err := balances.Length()
	if err != nil {
		return err
	}
	if uint64(len(deltas.Penalties)) != balLen || uint64(len(deltas.Rewards)) != balLen {
		return errors.New("cannot apply deltas to balances list with different length")
	}
	// TODO: can be optimized a lot, make a new tree in one go
	for i := ValidatorIndex(0); i < ValidatorIndex(balLen); i++ {
		balances.IncreaseBalance(i, deltas.Rewards[i])
		balances.DecreaseBalance(i, deltas.Penalties[i])
	}
	return nil
}
