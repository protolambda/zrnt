package beacon

import (
	"context"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

// SlashingsHistory is a EPOCHS_PER_SLASHINGS_VECTOR vector
type SlashingsHistory []Gwei

func (a *SlashingsHistory) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	if Epoch(len(*a)) != spec.EPOCHS_PER_SLASHINGS_VECTOR {
		// re-use space if available (for recycling old state objects)
		if Epoch(cap(*a)) >= spec.EPOCHS_PER_SLASHINGS_VECTOR {
			*a = (*a)[:spec.EPOCHS_PER_SLASHINGS_VECTOR]
		} else {
			*a = make([]Gwei, spec.EPOCHS_PER_SLASHINGS_VECTOR, spec.EPOCHS_PER_SLASHINGS_VECTOR)
		}
	}
	return dr.Vector(func(i uint64) codec.Deserializable {
		return &(*a)[i]
	}, GweiType.TypeByteLength(), uint64(spec.EPOCHS_PER_SLASHINGS_VECTOR))
}

func (a SlashingsHistory) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.Vector(func(i uint64) codec.Serializable {
		return &a[i]
	}, GweiType.TypeByteLength(), uint64(len(a)))
}

func (a SlashingsHistory) ByteLength(spec *Spec) (out uint64) {
	return uint64(len(a)) * GweiType.TypeByteLength()
}

func (a *SlashingsHistory) FixedLength(spec *Spec) uint64 {
	return uint64(spec.EPOCHS_PER_SLASHINGS_VECTOR) * GweiType.TypeByteLength()
}

func (li SlashingsHistory) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.Uint64VectorHTR(func(i uint64) uint64 {
		return uint64(li[i])
	}, uint64(len(li)))
}

// Balances slashed at every withdrawal period
func (c *Phase0Config) Slashings() VectorTypeDef {
	return VectorType(GweiType, uint64(c.EPOCHS_PER_SLASHINGS_VECTOR))
}

type SlashingsView struct{ *BasicVectorView }

func AsSlashings(v View, err error) (*SlashingsView, error) {
	c, err := AsBasicVector(v, err)
	return &SlashingsView{c}, nil
}

func (sl *SlashingsView) GetSlashingsValue(epoch Epoch) (Gwei, error) {
	i := uint64(epoch) % sl.VectorLength
	return AsGwei(sl.Get(i))
}

func (sl *SlashingsView) ResetSlashings(epoch Epoch) error {
	i := uint64(epoch) % sl.VectorLength
	return sl.Set(i, Uint64View(0))
}

func (sl *SlashingsView) AddSlashing(epoch Epoch, add Gwei) error {
	prev, err := sl.GetSlashingsValue(epoch)
	if err != nil {
		return err
	}
	i := uint64(epoch) % sl.VectorLength
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
func (spec *Spec) SlashValidator(epc *EpochsContext, state *BeaconStateView, slashedIndex ValidatorIndex, whistleblowerIndex *ValidatorIndex) error {
	currentEpoch := epc.CurrentEpoch.Epoch
	if err := spec.InitiateValidatorExit(epc, state, slashedIndex); err != nil {
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
	withdrawalEpoch := currentEpoch + spec.EPOCHS_PER_SLASHINGS_VECTOR
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
	if err := bals.DecreaseBalance(slashedIndex, effectiveBalance/Gwei(spec.MIN_SLASHING_PENALTY_QUOTIENT)); err != nil {
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
	whistleblowerReward := effectiveBalance / Gwei(spec.WHISTLEBLOWER_REWARD_QUOTIENT)
	proposerReward := whistleblowerReward / Gwei(spec.PROPOSER_REWARD_QUOTIENT)
	if err := bals.IncreaseBalance(propIndex, proposerReward); err != nil {
		return err
	}
	if err := bals.IncreaseBalance(*whistleblowerIndex, whistleblowerReward-proposerReward); err != nil {
		return err
	}
	return nil
}

func (spec *Spec) ProcessEpochSlashings(ctx context.Context, epc *EpochsContext, process *EpochProcess, state *BeaconStateView) error {
	select {
	case <-ctx.Done():
		return TransitionCancelErr
	default: // Don't block.
		break
	}
	totalBalance := process.TotalActiveStake

	slashings, err := state.Slashings()
	if err != nil {
		return err
	}

	slashingsSum, err := slashings.Total()
	if err != nil {
		return err
	}
	slashingsWeight := slashingsSum * Gwei(spec.PROPORTIONAL_SLASHING_MULTIPLIER)
	var adjustedTotalSlashingBalance Gwei
	if totalBalance < slashingsWeight {
		adjustedTotalSlashingBalance = totalBalance
	} else {
		adjustedTotalSlashingBalance = slashingsWeight
	}

	bals, err := state.Balances()
	if err != nil {
		return err
	}
	for _, index := range process.IndicesToSlash {
		// Factored out from penalty numerator to avoid uint64 overflow
		slashedEffectiveBal := process.Statuses[index].Validator.EffectiveBalance
		penaltyNumerator := slashedEffectiveBal / spec.EFFECTIVE_BALANCE_INCREMENT
		penaltyNumerator *= adjustedTotalSlashingBalance
		penalty := penaltyNumerator / totalBalance * spec.EFFECTIVE_BALANCE_INCREMENT

		if err := bals.DecreaseBalance(index, penalty); err != nil {
			return err
		}
	}
	return nil
}
