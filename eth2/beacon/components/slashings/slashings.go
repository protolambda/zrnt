package slashings

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/meta"
	. "github.com/protolambda/zrnt/eth2/core"
)

type SlashingReq interface {
	VersioningMeta
	ValidatorMeta
	ProposingMeta
	BalanceMeta
	ExitMeta
}

type SlashingsState struct {
	// Balances slashed at every withdrawal period
	Slashings [EPOCHS_PER_SLASHINGS_VECTOR]Gwei
}

func (state *SlashingsState) ResetSlashings(epoch Epoch) {
	state.Slashings[epoch%EPOCHS_PER_SLASHINGS_VECTOR] = 0
}

// Slash the validator with the given index.
func (state *SlashingsState) SlashValidator(meta SlashingReq, slashedIndex ValidatorIndex, whistleblowerIndex *ValidatorIndex) {
	currentEpoch := meta.Epoch()
	validator := meta.Validator(slashedIndex)
	meta.InitiateValidatorExit(slashedIndex)
	validator.Slashed = true
	validator.WithdrawableEpoch = currentEpoch + EPOCHS_PER_SLASHINGS_VECTOR
	slashedBalance := validator.EffectiveBalance
	state.Slashings[currentEpoch%EPOCHS_PER_SLASHINGS_VECTOR] += slashedBalance

	propIndex := meta.GetBeaconProposerIndex()
	if whistleblowerIndex == nil {
		whistleblowerIndex = &propIndex
	}
	whistleblowerReward := slashedBalance / WHISTLEBLOWER_REWARD_QUOTIENT
	proposerReward := whistleblowerReward / PROPOSER_REWARD_QUOTIENT
	meta.IncreaseBalance(propIndex, proposerReward)
	meta.IncreaseBalance(*whistleblowerIndex, whistleblowerReward-proposerReward)
	meta.DecreaseBalance(slashedIndex, whistleblowerReward)
}
