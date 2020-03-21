package phase0

import (
	"github.com/protolambda/zrnt/eth2/beacon/finality"
	"github.com/protolambda/zrnt/eth2/beacon/finalupdates"
	"github.com/protolambda/zrnt/eth2/beacon/registry"
	"github.com/protolambda/zrnt/eth2/beacon/rewardpenalty"
	"github.com/protolambda/zrnt/eth2/beacon/slashings"
)

type EpochProcessFeature struct {
	Meta interface {
		finality.JustificationEpochProcess
		rewardpenalty.RewardsAndPenaltiesEpochProcess
		registry.RegistryUpdateEpochProcess
		slashings.SlashingsEpochProcess
		finalupdates.FinalUpdatesEpochProcess
	}
}

func (f *EpochProcessFeature) ProcessEpoch() error {
	if err := f.Meta.ProcessEpochJustification(); err != nil {
		return err
	}
	if err := f.Meta.ProcessEpochRewardsAndPenalties(); err != nil {
		return err
	}
	if err := f.Meta.ProcessEpochRegistryUpdates(); err != nil {
		return err
	}
	if err := f.Meta.ProcessEpochSlashings(); err != nil {
		return err
	}
	if err := f.Meta.ProcessEpochFinalUpdates(); err != nil {
		return err
	}
	return nil
}
