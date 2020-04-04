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

type SlashingsProp PropFn

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

type SlashingProcessInput interface {
	meta.Versioning
	meta.Proposers
	meta.Balance
	meta.Staking
	meta.EffectiveBalances
	meta.Slashing
	meta.SlashingTask
	meta.Exits
}

// Slash the validator with the given index.
func (state *SlashingsProp) SlashValidator(input SlashingProcessInput, slashedIndex ValidatorIndex, whistleblowerIndex *ValidatorIndex) error {
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

func (state *SlashingsProp) ProcessEpochSlashings(input SlashingProcessInput) error {
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
