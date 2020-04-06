package beacon

import (
	. "github.com/protolambda/ztyp/view"
)

// Balances slashed at every withdrawal period
var SlashingsType = VectorType(GweiType, uint64(EPOCHS_PER_SLASHINGS_VECTOR))

type SlashingsView struct { *BasicVectorView }

func AsSlashings(v View, err error) (*SlashingsView, error) {
	c, err := AsBasicVector(v, err)
	return &SlashingsView{c}, nil
}

func (sl *SlashingsView) GetSlashingsValue(epoch Epoch) (Gwei, error) {
	i := uint64(epoch%EPOCHS_PER_SLASHINGS_VECTOR)
	return AsGwei(sl.Get(i))
}

func (sl *SlashingsView) ResetSlashings(epoch Epoch) error {
	i := uint64(epoch%EPOCHS_PER_SLASHINGS_VECTOR)
	return sl.Set(i, Uint64View(0))
}

func (sl *SlashingsView) AddSlashing(epoch Epoch, add Gwei) error {
	prev, err := sl.GetSlashingsValue(epoch)
	if err != nil {
		return err
	}
	i := uint64(epoch%EPOCHS_PER_SLASHINGS_VECTOR)
	return sl.Set(i, Uint64View(prev + add))
}

func (sl *SlashingsView) Total() (sum Gwei, err error) {
	for i := Epoch(0); i < EPOCHS_PER_SLASHINGS_VECTOR; i++ {
		v, err := sl.GetSlashingsValue(i)
		if err != nil {
			return 0, err
		}
		sum += v
	}
	return
}

// Slash the validator with the given index.
func (state *BeaconStateView) SlashValidator(epc *EpochsContext, slashedIndex ValidatorIndex, whistleblowerIndex *ValidatorIndex) error {
	slot, err := input.CurrentSlot()
	if err != nil {
		return err
	}
	currentEpoch := slot.ToEpoch()

	if err := input.InitiateValidatorExit(currentEpoch, slashedIndex); err != nil {
		return err
	}
	input.SlashAndDelayWithdraw(slashedIndex, currentEpoch + EPOCHS_PER_SLASHINGS_VECTOR)

	effectiveBalance, err := input.EffectiveBalance(slashedIndex)
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

	if err := input.DecreaseBalance(slashedIndex, effectiveBalance/MIN_SLASHING_PENALTY_QUOTIENT); err != nil {
		return err
	}

	propIndex, err := input.GetBeaconProposerIndex(slot)
	if err != nil {
		return err
	}
	if whistleblowerIndex == nil {
		whistleblowerIndex = &propIndex
	}
	whistleblowerReward := effectiveBalance / WHISTLEBLOWER_REWARD_QUOTIENT
	proposerReward := whistleblowerReward / PROPOSER_REWARD_QUOTIENT
	if err := input.IncreaseBalance(propIndex, proposerReward); err != nil {
		return err
	}
	if err := input.IncreaseBalance(*whistleblowerIndex, whistleblowerReward-proposerReward); err != nil {
		return err
	}
	return nil
}

func (state *BeaconStateView) ProcessEpochSlashings(epc *EpochsContext, process *EpochProcess) error {
	totalBalance, err := input.GetTotalStake()
	if err != nil {
		return err
	}

	slashings, err := state.Slashings()
	if err != nil {
		return err
	}

	slashingsSum, err := slashings.Total()
	if err != nil {
		return err
	}

	toSlash, err := input.GetIndicesToSlash()
	if err != nil {
		return err
	}
	for _, index := range toSlash {
		// Factored out from penalty numerator to avoid uint64 overflow
		slashedEffectiveBal, err := input.EffectiveBalance(index)
		if err != nil {
			return err
		}
		penaltyNumerator := slashedEffectiveBal / EFFECTIVE_BALANCE_INCREMENT
		if slashingsWeight := slashingsSum * 3; totalBalance < slashingsWeight {
			penaltyNumerator *= totalBalance
		} else {
			penaltyNumerator *= slashingsWeight
		}
		penalty := penaltyNumerator / totalBalance * EFFECTIVE_BALANCE_INCREMENT
		if err := input.DecreaseBalance(index, penalty); err != nil {
			return err
		}
	}
	return nil
}
