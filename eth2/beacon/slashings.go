package beacon

import (
	. "github.com/protolambda/ztyp/view"
)

// Balances slashed at every withdrawal period
var SlashingsType = VectorType(GweiType, uint64(EPOCHS_PER_SLASHINGS_VECTOR))

type SlashingsView struct{ *BasicVectorView }

func AsSlashings(v View, err error) (*SlashingsView, error) {
	c, err := AsBasicVector(v, err)
	return &SlashingsView{c}, nil
}

func (sl *SlashingsView) GetSlashingsValue(epoch Epoch) (Gwei, error) {
	i := uint64(epoch % EPOCHS_PER_SLASHINGS_VECTOR)
	return AsGwei(sl.Get(i))
}

func (sl *SlashingsView) ResetSlashings(epoch Epoch) error {
	i := uint64(epoch % EPOCHS_PER_SLASHINGS_VECTOR)
	return sl.Set(i, Uint64View(0))
}

func (sl *SlashingsView) AddSlashing(epoch Epoch, add Gwei) error {
	prev, err := sl.GetSlashingsValue(epoch)
	if err != nil {
		return err
	}
	i := uint64(epoch % EPOCHS_PER_SLASHINGS_VECTOR)
	return sl.Set(i, Uint64View(prev+add))
}

func (sl *SlashingsView) Total() (sum Gwei, err error) {
	iter := sl.ReadonlyIter()
	for {
		el, ok, err := iter.Next()
		if err != nil {
			return 0, err
		}
		if !ok {
			break
		}
		value, err := AsGwei(el, nil)
		if err != nil {
			return 0, err
		}
		sum += value
	}
	return
}

// Slash the validator with the given index.
func (state *BeaconStateView) SlashValidator(epc *EpochsContext, slashedIndex ValidatorIndex, whistleblowerIndex *ValidatorIndex) error {
	currentEpoch := epc.CurrentEpoch.Epoch
	if err := state.InitiateValidatorExit(epc, slashedIndex); err != nil {
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
	withdrawalEpoch := currentEpoch + EPOCHS_PER_SLASHINGS_VECTOR
	if withdrawalEpoch > prevWithdrawalEpoch {
		if err := v.SetWithdrawableEpoch(withdrawalEpoch); err != nil {
			return err
		}
	}

	effectiveBalance, err := v.EffectiveBalance()
	if err != nil {
		return err
	}

	slashings, err := state.Slashings()
	if err != nil {
		return err
	}
	if err := slashings.AddSlashing(currentEpoch, effectiveBalance); err != nil {
		return err
	}

	bals, err := state.Balances()
	if err != nil {
		return err
	}
	if err := bals.DecreaseBalance(slashedIndex, effectiveBalance/MIN_SLASHING_PENALTY_QUOTIENT); err != nil {
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
	whistleblowerReward := effectiveBalance / WHISTLEBLOWER_REWARD_QUOTIENT
	proposerReward := whistleblowerReward / PROPOSER_REWARD_QUOTIENT
	if err := bals.IncreaseBalance(propIndex, proposerReward); err != nil {
		return err
	}
	if err := bals.IncreaseBalance(*whistleblowerIndex, whistleblowerReward-proposerReward); err != nil {
		return err
	}
	return nil
}

func (state *BeaconStateView) ProcessEpochSlashings(epc *EpochsContext, process *EpochProcess) error {
	totalBalance := process.TotalActiveStake

	slashings, err := state.Slashings()
	if err != nil {
		return err
	}

	slashingsSum, err := slashings.Total()
	if err != nil {
		return err
	}

	bals, err := state.Balances()
	if err != nil {
		return err
	}
	for _, index := range process.IndicesToSlash {
		// Factored out from penalty numerator to avoid uint64 overflow
		slashedEffectiveBal := process.Statuses[index].Validator.EffectiveBalance
		penaltyNumerator := slashedEffectiveBal / EFFECTIVE_BALANCE_INCREMENT
		if slashingsWeight := slashingsSum * 3; totalBalance < slashingsWeight {
			penaltyNumerator *= totalBalance
		} else {
			penaltyNumerator *= slashingsWeight
		}
		penalty := penaltyNumerator / totalBalance * EFFECTIVE_BALANCE_INCREMENT

		if err := bals.DecreaseBalance(index, penalty); err != nil {
			return err
		}
	}
	return nil
}
