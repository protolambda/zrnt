package slashings

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type SlashingsState struct {
	// Balances slashed at every withdrawal period
	Slashings [EPOCHS_PER_SLASHINGS_VECTOR]Gwei
}

func (state *SlashingsState) ResetSlashings(epoch Epoch) {
	state.Slashings[epoch%EPOCHS_PER_SLASHINGS_VECTOR] = 0
}

type SlashingFeature struct {
	State *SlashingsState
	Meta  interface {
		meta.Versioning
		meta.Validators
		meta.Proposers
		meta.Balance
		meta.Staking
		meta.EffectiveBalances
		meta.Slashing
		meta.Exits
	}
}

// Slash the validator with the given index.
func (f *SlashingFeature) SlashValidator(slashedIndex ValidatorIndex, whistleblowerIndex *ValidatorIndex) {
	slot := f.Meta.CurrentSlot()
	currentEpoch := slot.ToEpoch()

	validator := f.Meta.Validator(slashedIndex)
	f.Meta.InitiateValidatorExit(currentEpoch, slashedIndex)
	validator.Slashed = true
	validator.WithdrawableEpoch = currentEpoch + EPOCHS_PER_SLASHINGS_VECTOR

	f.State.Slashings[currentEpoch%EPOCHS_PER_SLASHINGS_VECTOR] = validator.EffectiveBalance

	f.Meta.DecreaseBalance(slashedIndex, validator.EffectiveBalance/MIN_SLASHING_PENALTY_QUOTIENT)

	propIndex := f.Meta.GetBeaconProposerIndex(slot)
	if whistleblowerIndex == nil {
		whistleblowerIndex = &propIndex
	}
	whistleblowerReward := validator.EffectiveBalance / WHISTLEBLOWER_REWARD_QUOTIENT
	proposerReward := whistleblowerReward / PROPOSER_REWARD_QUOTIENT
	f.Meta.IncreaseBalance(propIndex, proposerReward)
	f.Meta.IncreaseBalance(*whistleblowerIndex, whistleblowerReward-proposerReward)
}

func (f *SlashingFeature) ProcessEpochSlashings() {
	currentEpoch := f.Meta.CurrentEpoch()
	totalBalance := f.Meta.GetTotalStakedBalance(currentEpoch)

	epochIndex := currentEpoch % EPOCHS_PER_SLASHINGS_VECTOR
	// Compute slashed balances in the current epoch
	slashings := f.State.Slashings[(epochIndex+1)%EPOCHS_PER_SLASHINGS_VECTOR]

	withdrawableEpoch := currentEpoch + (EPOCHS_PER_SLASHINGS_VECTOR / 2)

	for _, index := range f.Meta.GetIndicesToSlash(withdrawableEpoch) {
		penaltyNumerator := f.Meta.EffectiveBalance(index) / EFFECTIVE_BALANCE_INCREMENT
		if slashingsWeight := slashings * 3; totalBalance < slashingsWeight {
			penaltyNumerator *= totalBalance
		} else {
			penaltyNumerator *= slashingsWeight
		}
		penalty := penaltyNumerator / totalBalance * EFFECTIVE_BALANCE_INCREMENT
		f.Meta.DecreaseBalance(index, penalty)
	}
}
