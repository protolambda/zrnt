package beacon

import "github.com/protolambda/zrnt/eth2/util/math"

type Deltas struct {
	// element for each validator in registry
	Rewards []Gwei
	// element for each validator in registry
	Penalties []Gwei
}

func NewDeltas(validatorCount uint64) *Deltas {
	return &Deltas{
		Rewards:   make([]Gwei, validatorCount, validatorCount),
		Penalties: make([]Gwei, validatorCount, validatorCount),
	}
}

func (deltas *Deltas) Add(other *Deltas) {
	if len(deltas.Penalties) != len(other.Penalties) ||
		len(deltas.Rewards) != len(other.Rewards) {
		panic("cannot add other deltas, lengths do not match")
	}
	for i := 0; i < len(deltas.Rewards); i++ {
		deltas.Rewards[i] += other.Rewards[i]
	}
	for i := 0; i < len(deltas.Penalties); i++ {
		deltas.Penalties[i] += other.Penalties[i]
	}
}

type Valuator interface {
	GetBaseReward(index ValidatorIndex) Gwei
	GetInactivityPenalty(index ValidatorIndex) Gwei
	IsNotFinalizing() bool
}

type DeltasCalculator func(state *BeaconState, v Valuator) *Deltas

type DefaultValuator struct {
	adjustedQuotient     uint64
	previousTotalBalance Gwei
	currentTotalBalance  Gwei
	epochsSinceFinality  Epoch
	state                *BeaconState
}

func NewDefaultValuator(state *BeaconState) *DefaultValuator {
	v := &DefaultValuator{state: state}
	v.previousTotalBalance = state.GetTotalBalanceOf(
		state.ValidatorRegistry.GetActiveValidatorIndices(state.Epoch() - 1))
	v.currentTotalBalance = state.GetTotalBalanceOf(
		state.ValidatorRegistry.GetActiveValidatorIndices(state.Epoch()))
	v.adjustedQuotient = math.IntegerSquareroot(uint64(v.previousTotalBalance)) / BASE_REWARD_QUOTIENT
	v.epochsSinceFinality = state.Epoch() + 1 - state.FinalizedEpoch
	return v
}

func (v *DefaultValuator) GetBaseReward(index ValidatorIndex) Gwei {
	// TODO: this could be precomputed for performance
	return v.state.GetEffectiveBalance(index) / Gwei(v.adjustedQuotient) / 5
}

func (v *DefaultValuator) GetInactivityPenalty(index ValidatorIndex) Gwei {
	extra := Gwei(0)
	if v.epochsSinceFinality > 4 {
		extra = v.state.GetEffectiveBalance(index) * Gwei(v.epochsSinceFinality) / INACTIVITY_PENALTY_QUOTIENT / 2
	}
	return v.GetBaseReward(index) + extra
}

func (v *DefaultValuator) IsNotFinalizing() bool {
	return v.epochsSinceFinality > 4
}
