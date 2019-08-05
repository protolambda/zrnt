package phase0

import (
	"github.com/protolambda/zrnt/eth2/beacon/crosslinks"
	"github.com/protolambda/zrnt/eth2/beacon/finality"
	"github.com/protolambda/zrnt/eth2/beacon/finalupdates"
	"github.com/protolambda/zrnt/eth2/beacon/registry"
	"github.com/protolambda/zrnt/eth2/beacon/rewardpenalty"
	"github.com/protolambda/zrnt/eth2/beacon/slashings"
)

type EpochProcessFeature struct {
	Meta interface {
		finality.JustificationEpochProcess
		crosslinks.CrosslinksEpochProcess
		rewardpenalty.RewardsAndPenaltiesEpochProcess
		registry.RegistryUpdateEpochProcess
		slashings.SlashingsEpochProcess
		finalupdates.FinalUpdatesEpochProcess
	}
}

func (f *EpochProcessFeature) ProcessEpoch() {
	f.Meta.ProcessEpochJustification()
	f.Meta.ProcessEpochCrosslinks()
	f.Meta.ProcessEpochRewardsAndPenalties()
	f.Meta.ProcessEpochRegistryUpdates()
	f.Meta.ProcessEpochSlashings()
	f.Meta.ProcessEpochFinalUpdates()
}
