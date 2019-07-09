package components

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type SlashingsState struct {
	// Balances slashed at every withdrawal period
	Slashings [EPOCHS_PER_SLASHINGS_VECTOR]Gwei
}

func (state *SlashingsState) ResetSlashings(epoch Epoch) {
	state.Slashings[epoch%EPOCHS_PER_SLASHINGS_VECTOR] = 0
}

// Slash the validator with the given index.
func (state *BeaconState) SlashValidator(slashedIndex ValidatorIndex, whistleblowerIndex *ValidatorIndex) {
	currentEpoch := state.Epoch()
	validator := state.Validators[slashedIndex]
	state.InitiateValidatorExit(slashedIndex)
	validator.Slashed = true
	validator.WithdrawableEpoch = currentEpoch + EPOCHS_PER_SLASHINGS_VECTOR
	slashedBalance := validator.EffectiveBalance
	state.Slashings[currentEpoch%EPOCHS_PER_SLASHINGS_VECTOR] += slashedBalance

	propIndex := state.GetBeaconProposerIndex()
	if whistleblowerIndex == nil {
		whistleblowerIndex = &propIndex
	}
	whistleblowerReward := slashedBalance / WHISTLEBLOWER_REWARD_QUOTIENT
	proposerReward := whistleblowerReward / PROPOSER_REWARD_QUOTIENT
	state.Balances.IncreaseBalance(propIndex, proposerReward)
	state.Balances.IncreaseBalance(*whistleblowerIndex, whistleblowerReward-proposerReward)
	state.Balances.DecreaseBalance(slashedIndex, whistleblowerReward)
}
