package beacon

import (

)

func (state *BeaconStateView) ProcessEpochFinalUpdates() error {
	currentSlot, err := input.CurrentSlot()
	if err != nil {
		return err
	}
	nextSlot := currentSlot + 1
	nextEpoch := currentSlot.ToEpoch() + 1
	if nextEpoch != nextSlot.ToEpoch() {
		panic("final epoch updates may only be executed at the end of an epoch")
	}

	// Reset eth1 data votes if it is the end of the voting period.
	if nextEpoch%EPOCHS_PER_ETH1_VOTING_PERIOD == 0 {
		if err := input.ResetEth1Votes(); err != nil {
			return err
		}
	}

	if err := input.UpdateEffectiveBalances(); err != nil {
		return err
	}
	if err := input.ResetSlashings(nextEpoch); err != nil {
		return err
	}
	if err := input.PrepareRandao(nextEpoch); err != nil {
		return err
	}

	// Set historical root accumulator
	if nextEpoch%SLOTS_PER_HISTORICAL_ROOT.ToEpoch() == 0 {
		if err := input.UpdateHistoricalRoots(); err != nil {
			return err
		}
	}

	if err := input.RotateEpochAttestations(); err != nil {
		return err
	}

	return nil
}
