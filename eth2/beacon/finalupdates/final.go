package finalupdates

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type FinalUpdatesEpochProcess interface {
	ProcessEpochFinalUpdates()
}

type FinalUpdateFeature struct {
	Meta interface {
		meta.Versioning
		meta.Eth1Voting
		meta.EffectiveBalancesUpdate
		meta.SlashingHistory
		meta.Randao
		meta.HistoryUpdate
		meta.EpochAttestations
	}
}

func (f *FinalUpdateFeature) ProcessEpochFinalUpdates() error {
	currentSlot, err := f.Meta.CurrentSlot()
	if err != nil {
		return err
	}
	nextSlot := currentSlot + 1
	nextEpoch := currentSlot.ToEpoch() + 1
	if nextEpoch != nextSlot.ToEpoch() {
		panic("final epoch updates may only be executed at the end of an epoch")
	}

	// Reset eth1 data votes if it is the end of the voting period.
	if nextSlot%SLOTS_PER_ETH1_VOTING_PERIOD == 0 {
		if err := f.Meta.ResetEth1Votes(); err != nil {
			return err
		}
	}

	if err := f.Meta.UpdateEffectiveBalances(); err != nil {
		return err
	}
	if err := f.Meta.ResetSlashings(nextEpoch); err != nil {
		return err
	}
	if err := f.Meta.PrepareRandao(nextEpoch); err != nil {
		return err
	}

	// Set historical root accumulator
	if nextEpoch%SLOTS_PER_HISTORICAL_ROOT.ToEpoch() == 0 {
		if err := f.Meta.UpdateHistoricalRoots(); err != nil {
			return err
		}
	}

	if err := f.Meta.RotateEpochAttestations(); err != nil {
		return err
	}

	return nil
}
