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
	*SlashingsState
	Meta interface {
		meta.VersioningMeta
		meta.ValidatorMeta
		meta.ProposingMeta
		meta.BalanceMeta
		meta.StakingMeta
		meta.EffectiveBalanceMeta
		meta.SlashingMeta
		meta.ExitMeta
	}
}

// Slash the validator with the given index.
func (state *SlashingFeature) SlashValidator(slashedIndex ValidatorIndex, whistleblowerIndex *ValidatorIndex) {
	slot := state.Meta.CurrentSlot()
	currentEpoch := slot.ToEpoch()
	validator := state.Meta.Validator(slashedIndex)
	state.Meta.InitiateValidatorExit(currentEpoch, slashedIndex)
	validator.Slashed = true
	validator.WithdrawableEpoch = currentEpoch + EPOCHS_PER_SLASHINGS_VECTOR
	slashedBalance := validator.EffectiveBalance
	state.Slashings[currentEpoch%EPOCHS_PER_SLASHINGS_VECTOR] += slashedBalance

	propIndex := state.Meta.GetBeaconProposerIndex(slot)
	if whistleblowerIndex == nil {
		whistleblowerIndex = &propIndex
	}
	whistleblowerReward := slashedBalance / WHISTLEBLOWER_REWARD_QUOTIENT
	proposerReward := whistleblowerReward / PROPOSER_REWARD_QUOTIENT
	state.Meta.IncreaseBalance(propIndex, proposerReward)
	state.Meta.IncreaseBalance(*whistleblowerIndex, whistleblowerReward-proposerReward)
	state.Meta.DecreaseBalance(slashedIndex, whistleblowerReward)
}

func (state *SlashingFeature) ProcessEpochSlashings() {
	currentEpoch := state.Meta.CurrentEpoch()
	totalBalance := state.Meta.GetTotalStakedBalance(currentEpoch)

	epochIndex := currentEpoch % EPOCHS_PER_SLASHINGS_VECTOR
	// Compute slashed balances in the current epoch
	slashings := state.Slashings[(epochIndex+1)%EPOCHS_PER_SLASHINGS_VECTOR]

	withdrawableEpoch := currentEpoch+(EPOCHS_PER_SLASHINGS_VECTOR/2)

	for _, index := range state.Meta.GetIndicesToSlash(withdrawableEpoch) {
		penaltyNumerator := state.Meta.EffectiveBalance(index) / EFFECTIVE_BALANCE_INCREMENT
		if slashingsWeight := slashings * 3; totalBalance < slashingsWeight {
			penaltyNumerator *= totalBalance
		} else {
			penaltyNumerator *= slashingsWeight
		}
		penalty := penaltyNumerator / totalBalance * EFFECTIVE_BALANCE_INCREMENT
		state.Meta.DecreaseBalance(index, penalty)
	}
}
