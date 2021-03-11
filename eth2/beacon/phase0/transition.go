package phase0

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/tree"
)

func ProcessSlot(ctx context.Context, spec *common.Spec, state *BeaconStateView) error {
	select {
	case <-ctx.Done():
		return common.TransitionCancelErr
	default:
		break // Continue slot processing, don't block.
	}
	// The state root could take long, but absolute worst case is around a 1.5 seconds.
	// With any caching, this is more like < 50 ms. So no context use.
	// Cache state root
	previousStateRoot := state.HashTreeRoot(tree.GetHashFn())

	stateRootsBatch, err := state.StateRoots()
	if err != nil {
		return err
	}
	slot, err := state.Slot()
	if err != nil {
		return err
	}
	if err := stateRootsBatch.SetRoot(slot, previousStateRoot); err != nil {
		return err
	}

	latestHeader, err := state.LatestBlockHeader()
	if err != nil {
		return err
	}
	stateRoot, err := latestHeader.StateRoot()
	if err != nil {
		return err
	}
	if stateRoot == (common.Root{}) {
		if err := latestHeader.SetStateRoot(previousStateRoot); err != nil {
			return err
		}
	}
	previousBlockRoot := latestHeader.HashTreeRoot(tree.GetHashFn())

	// Cache latest known block and state root
	blockRootsBatch, err := state.BlockRoots()
	if err != nil {
		return err
	}
	if err := blockRootsBatch.SetRoot(slot, previousBlockRoot); err != nil {
		return err
	}

	return nil
}

func ProcessEpoch(ctx context.Context, spec *common.Spec, epc *EpochsContext, state *BeaconStateView) error {
	process, err := PrepareEpochProcess(ctx, spec, epc, state)
	if err != nil {
		return err
	}
	if err := ProcessEpochJustification(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := ProcessEpochRewardsAndPenalties(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := ProcessEpochRegistryUpdates(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := ProcessEpochSlashings(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := ProcessEpochFinalUpdates(ctx, spec, epc, process, state); err != nil {
		return err
	}
	return nil
}

func ProcessBlock(ctx context.Context, spec *common.Spec, epc *EpochsContext, state *BeaconStateView, block *BeaconBlock) error {
	if err := ProcessHeader(ctx, spec, epc, state, block); err != nil {
		return err
	}
	body := &block.Body
	if err := ProcessRandaoReveal(ctx, spec, epc, state, body.RandaoReveal); err != nil {
		return err
	}
	if err := ProcessEth1Vote(ctx, spec, epc, state, body.Eth1Data); err != nil {
		return err
	}
	// Safety checks, in case the user of the function provided too many operations
	if err := body.CheckLimits(spec); err != nil {
		return err
	}

	if err := ProcessProposerSlashings(ctx, spec, epc, state, body.ProposerSlashings); err != nil {
		return err
	}
	if err := ProcessAttesterSlashings(ctx, spec, epc, state, body.AttesterSlashings); err != nil {
		return err
	}
	if err := ProcessAttestations(ctx, spec, epc, state, body.Attestations); err != nil {
		return err
	}
	if err := ProcessDeposits(ctx, spec, epc, state, body.Deposits); err != nil {
		return err
	}
	if err := ProcessVoluntaryExits(ctx, spec, epc, state, body.VoluntaryExits); err != nil {
		return err
	}
	return nil
}
