package registry

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type Balances []Gwei

func (bals Balances) IncreaseBalance(index ValidatorIndex, delta Gwei) {
	bals[index] += delta
}

func (bals Balances) DecreaseBalance(index ValidatorIndex, delta Gwei) {
	currentBalance := bals[index]
	// prevent underflow, clip to 0
	if currentBalance >= delta {
		bals[index] -= delta
	} else {
		bals[index] = 0
	}
}

func (bals Balances) ApplyDeltas(deltas *Deltas) {
	if len(deltas.Penalties) != len(bals) || len(deltas.Rewards) != len(bals) {
		panic("cannot apply deltas to balances list with different length")
	}
	for i := ValidatorIndex(0); i < ValidatorIndex(len(bals)); i++ {
		bals.IncreaseBalance(i, deltas.Rewards[i])
		bals.DecreaseBalance(i, deltas.Penalties[i])
	}
}
