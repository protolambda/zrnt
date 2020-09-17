package beacon

import "context"

func (spec *Spec) ProcessEpochFinalUpdates(ctx context.Context, epc *EpochsContext, process *EpochProcess, state *BeaconStateView) error {
	select {
	case <-ctx.Done():
		return TransitionCancelErr
	default: // Don't block.
		break
	}
	nextEpoch := epc.NextEpoch.Epoch

	// Reset eth1 data votes if it is the end of the voting period.
	if nextEpoch%spec.EPOCHS_PER_ETH1_VOTING_PERIOD == 0 {
		if err := spec.ResetEth1Votes(state); err != nil {
			return err
		}
	}

	// update effective balances
	{
		HYSTERESIS_INCREMENT := spec.EFFECTIVE_BALANCE_INCREMENT / Gwei(spec.HYSTERESIS_QUOTIENT)
		DOWNWARD_THRESHOLD := HYSTERESIS_INCREMENT * Gwei(spec.HYSTERESIS_DOWNWARD_MULTIPLIER)
		UPWARD_THRESHOLD := HYSTERESIS_INCREMENT * Gwei(spec.HYSTERESIS_UPWARD_MULTIPLIER)

		vals, err := state.Validators()
		if err != nil {
			return err
		}
		bals, err := state.Balances()
		if err != nil {
			return err
		}
		balIter := bals.ReadonlyIter()
		for i := ValidatorIndex(0); true; i++ {
			el, ok, err := balIter.Next()
			if err != nil {
				return err
			}
			if !ok {
				break
			}
			balance, err := AsGwei(el, nil)
			if err != nil {
				return err
			}
			effBalance := process.Statuses[i].Validator.EffectiveBalance
			if balance+DOWNWARD_THRESHOLD < effBalance || effBalance+UPWARD_THRESHOLD < balance {
				effBalance = balance - (balance % spec.EFFECTIVE_BALANCE_INCREMENT)
				if spec.MAX_EFFECTIVE_BALANCE < effBalance {
					effBalance = spec.MAX_EFFECTIVE_BALANCE
				}
				val, err := vals.Validator(i)
				if err != nil {
					return err
				}
				if err := val.SetEffectiveBalance(effBalance); err != nil {
					return err
				}
			}
		}
	}

	slashings, err := state.Slashings()
	if err != nil {
		return err
	}
	if err := slashings.ResetSlashings(nextEpoch); err != nil {
		return err
	}

	mixes, err := state.RandaoMixes()
	if err != nil {
		return err
	}
	if err := mixes.PrepareRandao(nextEpoch); err != nil {
		return err
	}

	// Set historical root accumulator
	if nextEpoch%spec.SlotToEpoch(spec.SLOTS_PER_HISTORICAL_ROOT) == 0 {
		if err := state.UpdateHistoricalRoots(); err != nil {
			return err
		}
	}

	// Rotate current/previous epoch attestations
	prevAtts, err := state.PreviousEpochAttestations()
	if err != nil {
		return err
	}
	currAtts, err := state.CurrentEpochAttestations()
	if err != nil {
		return err
	}
	if err := prevAtts.SetBacking(currAtts.Backing()); err != nil {
		return err
	}
	if err := currAtts.SetBacking(spec.PendingAttestations().DefaultNode()); err != nil {
		return err
	}

	return nil
}
