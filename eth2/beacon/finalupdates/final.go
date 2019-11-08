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

func (f *FinalUpdateFeature) ProcessEpochFinalUpdates() {
	nextSlot := f.Meta.CurrentSlot() + 1
	nextEpoch := f.Meta.CurrentEpoch() + 1
	if nextEpoch != nextSlot.ToEpoch() {
		panic("final epoch updates may only be executed at the end of an epoch")
	}

	// Reset eth1 data votes if it is the end of the voting period.
	if nextSlot%SLOTS_PER_ETH1_VOTING_PERIOD == 0 {
		f.Meta.ResetEth1Votes()
	}

	f.Meta.UpdateEffectiveBalances()
	f.Meta.ResetSlashings(nextEpoch)
	f.Meta.PrepareRandao(nextEpoch)

	// Set historical root accumulator
	if nextEpoch%SLOTS_PER_HISTORICAL_ROOT.ToEpoch() == 0 {
		f.Meta.UpdateHistoricalRoots()
	}

	f.Meta.RotateEpochAttestations()
}
