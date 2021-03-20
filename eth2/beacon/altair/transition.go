package altair

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/tree"
)

func ProcessEpoch(ctx context.Context, spec *common.Spec, epc *phase0.EpochsContext, state *BeaconStateView) error {
	process, err := phase0.PrepareEpochProcess(ctx, spec, epc, state)
	if err != nil {
		return err
	}
	if err := phase0.ProcessEpochJustification(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := phase0.ProcessEpochRewardsAndPenalties(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := phase0.ProcessEpochRegistryUpdates(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := phase0.ProcessEpochSlashings(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := phase0.ProcessEffectiveBalanceUpdates(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := phase0.ProcessEth1DataReset(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := phase0.ProcessSlashingsReset(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := phase0.ProcessRandaoMixesReset(ctx, spec, epc, process, state); err != nil {
		return err
	}
	return nil
}

func ProcessBlock(ctx context.Context, spec *common.Spec, state *BeaconStateView, block *BeaconBlock) error {
	if err := common.ProcessHeader(ctx, spec, epc, state, block); err != nil {
		return err
	}
	body := &block.Body
	if err := phase0.ProcessRandaoReveal(ctx, spec, epc, state, body.RandaoReveal); err != nil {
		return err
	}
	if err := phase0.ProcessEth1Vote(ctx, spec, epc, state, body.Eth1Data); err != nil {
		return err
	}
	// Safety checks, in case the user of the function provided too many operations
	if err := body.CheckLimits(spec); err != nil {
		return err
	}

	if err := phase0.ProcessProposerSlashings(ctx, spec, epc, state, body.ProposerSlashings); err != nil {
		return err
	}
	if err := phase0.ProcessAttesterSlashings(ctx, spec, epc, state, body.AttesterSlashings); err != nil {
		return err
	}
	if err := phase0.ProcessAttestations(ctx, spec, epc, state, body.Attestations); err != nil {
		return err
	}
	// TODO new deposit processing
	if err := phase0.ProcessDeposits(ctx, spec, epc, state, body.Deposits); err != nil {
		return err
	}
	if err := phase0.ProcessVoluntaryExits(ctx, spec, epc, state, body.VoluntaryExits); err != nil {
		return err
	}
	ProcessSyncCommittee(ctx, spec, body.Sync)
	return nil
}
