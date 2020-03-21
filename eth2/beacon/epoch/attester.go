package epoch

import (
	"github.com/protolambda/zrnt/eth2/beacon/finality"
	"github.com/protolambda/zrnt/eth2/beacon/finalupdates"
	"github.com/protolambda/zrnt/eth2/beacon/registry"
	"github.com/protolambda/zrnt/eth2/beacon/rewardpenalty"
	"github.com/protolambda/zrnt/eth2/beacon/slashings"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)


type EpochProcessState struct {
	Statuses []AttesterStatus
	TotalActiveStake Gwei
	PrevEpoch EpochStakeSummary
	CurrEpoch EpochStakeSummary
	ActiveValidators uint64
	IndicesToSlash []ValidatorIndex
	IndicesToActivate []ValidatorIndex
	ExitQueueEnd Epoch
	ChurnLimit uint64
}

type EpochProcessors interface {
	finality.JustificationEpochProcess
	rewardpenalty.RewardsAndPenaltiesEpochProcess
	registry.RegistryUpdateEpochProcess
	slashings.SlashingsEpochProcess
	finalupdates.FinalUpdatesEpochProcess
}

type EpochProcessInput interface {
	finality.JustificationEpochProcessInput
	meta.Versioning
}

func (state *EpochProcessState) ProcessEpoch(proc EpochProcessors) error {
	if err := proc.ProcessEpochJustification(state); err != nil {
		return err
	}
	if err := proc.ProcessEpochRewardsAndPenalties(state); err != nil {
		return err
	}
	if err := proc.ProcessEpochRegistryUpdates(state); err != nil {
		return err
	}
	if err := proc.ProcessEpochSlashings(state); err != nil {
		return err
	}
	if err := proc.ProcessEpochFinalUpdates(state); err != nil {
		return err
	}
	return nil
}

