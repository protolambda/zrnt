package slashings

import (
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	. "github.com/protolambda/ztyp/props"
	. "github.com/protolambda/ztyp/view"
)

// Balances slashed at every withdrawal period
var SlashingsType = VectorType(GweiType, uint64(EPOCHS_PER_SLASHINGS_VECTOR))

type Slashings struct { *BasicVectorView }

func (sl *Slashings) GetSlashingsValue(epoch Epoch) (Gwei, error) {
	prev, err := sl.Get(uint64(epoch%EPOCHS_PER_SLASHINGS_VECTOR))
	if err != nil {
		return 0, err
	} else if prevSlashed, ok := prev.(Uint64View); !ok {
		return 0, fmt.Errorf("cannot read previous slashed stake as Uint64: %v", prev)
	} else {
		return Gwei(prevSlashed), nil
	}
}

func (sl *Slashings) ResetSlashings(epoch Epoch) error {
	return sl.Set(uint64(epoch%EPOCHS_PER_SLASHINGS_VECTOR), Uint64View(0))
}
func (sl *Slashings) AddSlashing(epoch Epoch, add Gwei) error {
	prev, err := sl.GetSlashingsValue(epoch)
	if err != nil {
		return err
	}
	return sl.Set(uint64(epoch%EPOCHS_PER_SLASHINGS_VECTOR), Uint64View(prev + add))
}

func (sl *Slashings) Total() (sum Gwei, err error) {
	for i := Epoch(0); i < EPOCHS_PER_SLASHINGS_VECTOR; i++ {
		v, err := sl.GetSlashingsValue(i)
		if err != nil {
			return 0, err
		}
		sum += v
	}
	return
}

type SlashingsProp ReadPropFn

func (p SlashingsProp) Slashings() (*Slashings, error) {
	if v, err := p(); err != nil {
		return nil, err
	} else if f, ok := v.(*Slashings); !ok {
		return nil, fmt.Errorf("not a fork view: %v", v)
	} else {
		return f, nil
	}
}

type SlashingsEpochProcess interface {
	ProcessEpochSlashings() error
}

type SlashingFeature struct {
	State SlashingsProp
	Meta  interface {
		meta.Versioning
		meta.Proposers
		meta.Balance
		meta.Staking
		meta.EffectiveBalances
		meta.Slashing
		meta.Exits
	}
}

// Slash the validator with the given index.
func (f *SlashingFeature) SlashValidator(slashedIndex ValidatorIndex, whistleblowerIndex *ValidatorIndex) error {
	slot, err := f.Meta.CurrentSlot()
	if err != nil {
		return err
	}
	currentEpoch := slot.ToEpoch()

	if err := f.Meta.InitiateValidatorExit(currentEpoch, slashedIndex); err != nil {
		return err
	}
	f.Meta.SlashAndDelayWithdraw(slashedIndex, currentEpoch + EPOCHS_PER_SLASHINGS_VECTOR)

	effectiveBalance, err := f.Meta.EffectiveBalance(slashedIndex)
	if err != nil {
		return err
	}

	slashings, err := f.State.Slashings()
	if err != nil {
		return err
	}
	if err := slashings.AddSlashing(currentEpoch, effectiveBalance); err != nil {
		return err
	}

	if err := f.Meta.DecreaseBalance(slashedIndex, effectiveBalance/MIN_SLASHING_PENALTY_QUOTIENT); err != nil {
		return err
	}

	propIndex, err := f.Meta.GetBeaconProposerIndex(slot)
	if err != nil {
		return err
	}
	if whistleblowerIndex == nil {
		whistleblowerIndex = &propIndex
	}
	whistleblowerReward := effectiveBalance / WHISTLEBLOWER_REWARD_QUOTIENT
	proposerReward := whistleblowerReward / PROPOSER_REWARD_QUOTIENT
	if err := f.Meta.IncreaseBalance(propIndex, proposerReward); err != nil {
		return err
	}
	if err := f.Meta.IncreaseBalance(*whistleblowerIndex, whistleblowerReward-proposerReward); err != nil {
		return err
	}
	return nil
}

func (f *SlashingFeature) ProcessEpochSlashings() error {
	totalBalance, err := f.Meta.GetTotalStake()
	if err != nil {
		return err
	}

	slashings, err := f.State.Slashings()
	if err != nil {
		return err
	}

	slashingsSum, err := slashings.Total()
	if err != nil {
		return err
	}

	currentEpoch, err := f.Meta.CurrentEpoch()
	if err != nil {
		return err
	}
	withdrawableEpoch := currentEpoch + (EPOCHS_PER_SLASHINGS_VECTOR / 2)

	toSlash, err := f.Meta.GetIndicesToSlash(withdrawableEpoch)
	if err != nil {
		return err
	}
	for _, index := range toSlash {
		// Factored out from penalty numerator to avoid uint64 overflow
		slashedEffectiveBal, err := f.Meta.EffectiveBalance(index)
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
		if err := f.Meta.DecreaseBalance(index, penalty); err != nil {
			return err
		}
	}
	return nil
}
