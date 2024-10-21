package phase0

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

// Slash the validator with the given index.
func SlashValidator(spec *common.Spec, epc *common.EpochsContext, state common.BeaconState,
	slashedIndex common.ValidatorIndex, whistleblowerIndex *common.ValidatorIndex) error {

	currentEpoch := epc.CurrentEpoch.Epoch
	if err := InitiateValidatorExit(spec, epc, state, slashedIndex); err != nil {
		return err
	}
	vals, err := state.Validators()
	if err != nil {
		return err
	}
	v, err := vals.Validator(slashedIndex)
	if err != nil {
		return err
	}
	if err := v.MakeSlashed(); err != nil {
		return err
	}
	prevWithdrawalEpoch, err := v.WithdrawableEpoch()
	if err != nil {
		return err
	}
	withdrawalEpoch := currentEpoch + spec.MIN_SLASHING_WITHDRAWABLE_DELAY
	if withdrawalEpoch > prevWithdrawalEpoch {
		if err := v.SetWithdrawableEpoch(withdrawalEpoch); err != nil {
			return err
		}
	}

	effectiveBalance, err := v.EffectiveBalance()
	if err != nil {
		return err
	}

	settings := state.ForkSettings(spec)
	bals, err := state.Balances()
	if err != nil {
		return err
	}
	if err := common.DecreaseBalance(bals, slashedIndex, effectiveBalance/common.Gwei(settings.MinSlashingPenaltyQuotient)); err != nil {
		return err
	}

	slot, err := state.Slot()
	if err != nil {
		return err
	}
	propIndex, err := epc.GetBeaconProposer(slot)
	if err != nil {
		return err
	}
	if whistleblowerIndex == nil {
		whistleblowerIndex = &propIndex
	}
	whistleblowerReward := effectiveBalance / common.Gwei(spec.WHISTLEBLOWER_REWARD_QUOTIENT)
	proposerReward := settings.CalcProposerShare(whistleblowerReward)
	if err := common.IncreaseBalance(bals, propIndex, proposerReward); err != nil {
		return err
	}
	if err := common.IncreaseBalance(bals, *whistleblowerIndex, whistleblowerReward-proposerReward); err != nil {
		return err
	}
	return nil
}
