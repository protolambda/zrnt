package altair

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
)

func ProcessEpoch(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView) error {
	vals, err := state.Validators()
	if err != nil {
		return err
	}
	flats, err := common.FlattenValidators(vals)
	if err != nil {
		return err
	}
	just := phase0.JustificationStakeData{
		CurrentEpoch: epc.CurrentEpoch.Epoch,
		// TODO
		TotalActiveStake:              0,
		PrevEpochUnslashedTargetStake: 0, // Now with TIMELY_TARGET_FLAG_INDEX participation
		CurrEpochUnslashedTargetStake: 0, // ditto
	}
	if err := phase0.ProcessEpochJustification(ctx, spec, &just, state); err != nil {
		return err
	}
	if err := ProcessEpochRewardsAndPenalties(ctx, spec, epc, state); err != nil {
		return err
	}
	if err := phase0.ProcessEpochRegistryUpdates(ctx, spec, epc, flats, state); err != nil {
		return err
	}
	if err := phase0.ProcessEpochSlashings(ctx, spec, epc, flats, state); err != nil {
		return err
	}
	if err := phase0.ProcessEth1DataReset(ctx, spec, epc, state); err != nil {
		return err
	}
	if err := phase0.ProcessEffectiveBalanceUpdates(ctx, spec, epc, flats, state); err != nil {
		return err
	}
	if err := phase0.ProcessSlashingsReset(ctx, spec, epc, state); err != nil {
		return err
	}
	if err := phase0.ProcessRandaoMixesReset(ctx, spec, epc, state); err != nil {
		return err
	}
	if err := phase0.ProcessHistoricalRootsUpdate(ctx, spec, epc, state); err != nil {
		return err
	}
	if err := ProcessParticipationFlagUpdates(ctx, spec, state); err != nil {
		return err
	}
	if err := ProcessSyncCommitteeUpdates(ctx, spec, state); err != nil {
		return err
	}
	return nil
}

func ProcessBlock(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, block *BeaconBlock) error {
	header := block.Header(spec)
	expectedProposer, err := epc.GetBeaconProposer(block.Slot)
	if err != nil {
		return err
	}
	if err := common.ProcessHeader(ctx, spec, state, header, expectedProposer); err != nil {
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
	// Note: state.AddValidator changed in Altair, but the deposit processing itself stayed the same.
	if err := phase0.ProcessDeposits(ctx, spec, epc, state, body.Deposits); err != nil {
		return err
	}
	if err := phase0.ProcessVoluntaryExits(ctx, spec, epc, state, body.VoluntaryExits); err != nil {
		return err
	}
	if err := ProcessSyncCommittee(ctx, spec, state, &body.SyncAggregate); err != nil {
		return err
	}
	return nil
}
