package slashings

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
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
	meta.InitiateValidatorExit(currentEpoch, slashedIndex)
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

type EpochSlashingReq interface {
	VersioningMeta
	StakingMeta
	EffectiveBalanceMeta
	BalanceMeta
	SlashingMeta
}

func (state *SlashingsState) ProcessEpochSlashings(meta EpochSlashingReq) {
	currentEpoch := meta.Epoch()
	totalBalance := meta.GetTotalStakedBalance(currentEpoch)

	epochIndex := currentEpoch % EPOCHS_PER_SLASHINGS_VECTOR
	// Compute slashed balances in the current epoch
	slashings := state.Slashings[(epochIndex+1)%EPOCHS_PER_SLASHINGS_VECTOR]

	withdrawableEpoch := currentEpoch+(EPOCHS_PER_SLASHINGS_VECTOR/2)

	for _, index := range meta.GetIndicesToSlash(withdrawableEpoch) {
		penaltyNumerator := meta.EffectiveBalance(index) / EFFECTIVE_BALANCE_INCREMENT
		if slashingsWeight := slashings * 3; totalBalance < slashingsWeight {
			penaltyNumerator *= totalBalance
		} else {
			penaltyNumerator *= slashingsWeight
		}
		penalty := penaltyNumerator / totalBalance * EFFECTIVE_BALANCE_INCREMENT
		meta.DecreaseBalance(index, penalty)
	}
}
