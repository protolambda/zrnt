package registry

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type Balances []Gwei

func (_ *Balances) Limit() uint64 {
	return VALIDATOR_REGISTRY_LIMIT
}

type BalancesState struct {
	Balances Balances
}

func (state *BalancesState) GetBalance(index ValidatorIndex) Gwei {
	return state.Balances[index]
}

func (state *BalancesState) IncreaseBalance(index ValidatorIndex, delta Gwei) {
	state.Balances[index] += delta
}

func (state *BalancesState) DecreaseBalance(index ValidatorIndex, delta Gwei) {
	currentBalance := state.Balances[index]
	// prevent underflow, clip to 0
	if currentBalance >= delta {
		state.Balances[index] -= delta
	} else {
		state.Balances[index] = 0
	}
}

func (state *BalancesState) ApplyDeltas(deltas *Deltas) {
	if len(deltas.Penalties) != len(state.Balances) || len(deltas.Rewards) != len(state.Balances) {
		panic("cannot apply deltas to balances list with different length")
	}
	for i := ValidatorIndex(0); i < ValidatorIndex(len(state.Balances)); i++ {
		state.IncreaseBalance(i, deltas.Rewards[i])
		state.DecreaseBalance(i, deltas.Penalties[i])
	}
}
