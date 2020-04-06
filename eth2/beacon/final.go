package beacon

import (

)

func (state *BeaconStateView) ProcessEpochFinalUpdates(epc *EpochsContext, process *EpochProcess) error {
	nextEpoch := epc.NextEpoch.Epoch

	// Reset eth1 data votes if it is the end of the voting period.
	if nextEpoch%EPOCHS_PER_ETH1_VOTING_PERIOD == 0 {
		if err := state.ResetEth1Votes(); err != nil {
			return err
		}
	}

	// update effective balances
	for i, v := range state.Validators {
		// TODO
		balance := state.Balances[i]
		if balance < v.EffectiveBalance ||
			v.EffectiveBalance+3*HALF_INCREMENT < balance {
			v.EffectiveBalance = balance - (balance % EFFECTIVE_BALANCE_INCREMENT)
			if MAX_EFFECTIVE_BALANCE < v.EffectiveBalance {
				v.EffectiveBalance = MAX_EFFECTIVE_BALANCE
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
	if nextEpoch%SLOTS_PER_HISTORICAL_ROOT.ToEpoch() == 0 {
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
	if err := currAtts.SetBacking(PendingAttestationsType.DefaultNode()); err != nil {
		return err
	}

	return nil
}
