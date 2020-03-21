package finalupdates

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type FinalUpdatesEpochProcess interface {
	ProcessEpochFinalUpdates(input FinalUpdateProcessInput) error
}

type FinalUpdateProcessInput interface {
	meta.Versioning
	meta.Eth1Voting
	meta.EffectiveBalancesUpdate
	meta.SlashingHistory
	meta.Randao
	meta.HistoryUpdate
	meta.EpochAttestations
}

func ProcessEpochFinalUpdates(input FinalUpdateProcessInput) error {
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
