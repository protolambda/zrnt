package beacon

import (
	"context"
	"errors"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/ztyp/tree"
)

var TransitionCancelErr = errors.New("state transition was cancelled")

func (spec *Spec) ProcessSlot(ctx context.Context, state *BeaconStateView) error {
	select {
	case <-ctx.Done():
		return TransitionCancelErr
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
	if stateRoot == (Root{}) {
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

func (spec *Spec) ProcessEpoch(ctx context.Context, epc *EpochsContext, state *BeaconStateView) error {
	process, err := spec.PrepareEpochProcess(ctx, epc, state)
	if err != nil {
		return err
	}
	if err := spec.ProcessEpochJustification(ctx, epc, process, state); err != nil {
		return err
	}
	if err := spec.ProcessEpochRewardsAndPenalties(ctx, epc, process, state); err != nil {
		return err
	}
	if err := spec.ProcessEpochRegistryUpdates(ctx, epc, process, state); err != nil {
		return err
	}
	if err := spec.ProcessEpochSlashings(ctx, epc, process, state); err != nil {
		return err
	}
	if err := spec.ProcessEpochFinalUpdates(ctx, epc, process, state); err != nil {
		return err
	}
	return nil
}

// Process the state to the given slot.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
func (spec *Spec) ProcessSlots(ctx context.Context, epc *EpochsContext, state *BeaconStateView, slot Slot) error {
	// happens at the start of every CurrentSlot
	currentSlot, err := state.Slot()
	if err != nil {
		return err
	}
	if currentSlot > slot {
		return errors.New("cannot transition from pre-state with higher or equal slot than transition target")
	}
	for currentSlot < slot {
		select {
		case <-ctx.Done():
			return TransitionCancelErr
		default:
			break // Continue slot processing, don't block.
		}
		if err := spec.ProcessSlot(ctx, state); err != nil {
			return err
		}
		// Per-epoch transition happens at the start of the first slot of every epoch.
		// (with the slot still at the end of the last epoch)
		isEpochEnd := spec.SlotToEpoch(currentSlot + 1) != spec.SlotToEpoch(currentSlot)
		if isEpochEnd {
			if err := spec.ProcessEpoch(ctx, epc, state); err != nil {
				return err
			}
		}
		currentSlot += 1
		if err := state.SetSlot(currentSlot); err != nil {
			return err
		}
		if isEpochEnd {
			if err := epc.RotateEpochs(state); err != nil {
				return err
			}
		}
	}
	return nil
}

func (spec *Spec) ProcessBlock(ctx context.Context, epc *EpochsContext, state *BeaconStateView, block *BeaconBlock) error {
	if err := spec.ProcessHeader(ctx, epc, state, block); err != nil {
		return err
	}
	body := &block.Body
	if err := spec.ProcessRandaoReveal(ctx, epc, state, body.RandaoReveal); err != nil {
		return err
	}
	if err := spec.ProcessEth1Vote(ctx, epc, state, body.Eth1Data); err != nil {
		return err
	}
	if err := spec.ProcessProposerSlashings(ctx, epc, state, &body.ProposerSlashings); err != nil {
		return err
	}
	if err := spec.ProcessAttesterSlashings(ctx, epc, state, &body.AttesterSlashings); err != nil {
		return err
	}
	if err := spec.ProcessAttestations(ctx, epc, state, &body.Attestations); err != nil {
		return err
	}
	if err := spec.ProcessDeposits(ctx, epc, state, &body.Deposits); err != nil {
		return err
	}
	if err := spec.ProcessVoluntaryExits(ctx, epc, state, &body.VoluntaryExits); err != nil {
		return err
	}
	return nil
}

// Transition the state to the slot of the given block, then processes the block.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
//
func (spec *Spec) StateTransition(ctx context.Context, epc *EpochsContext, state *BeaconStateView, block *SignedBeaconBlock, validateResult bool) error {
	if err := spec.ProcessSlots(ctx, epc, state, block.Message.Slot); err != nil {
		return err
	}
	if validateResult {
		// Safe to ignore proposer index, it will be checked as part of the ProcessHeader call.
		if !spec.VerifyBlockSignature(epc, state, block, false) {
			return errors.New("block has invalid signature")
		}
	}

	if err := spec.ProcessBlock(ctx, epc, state, &block.Message); err != nil {
		return err
	}

	// State root verification
	if validateResult && block.Message.StateRoot != state.HashTreeRoot(tree.GetHashFn()) {
		return errors.New("block has invalid state root")
	}
	return nil
}

// Assuming the slot is valid, and optionally assume the proposer index is valid, check if the signature is valid
func (spec *Spec) VerifyBlockSignature(epc *EpochsContext, state *BeaconStateView, block *SignedBeaconBlock, validateProposerIndex bool) bool {
	if validateProposerIndex {
		proposerIndex, err := epc.GetBeaconProposer(block.Message.Slot)
		if err != nil {
			return false
		}
		if proposerIndex != block.Message.ProposerIndex {
			return false
		}
	}
	pub, ok := epc.PubkeyCache.Pubkey(block.Message.ProposerIndex)
	if !ok {
		return false
	}
	domain, err := state.GetDomain(spec.DOMAIN_BEACON_PROPOSER, spec.SlotToEpoch(block.Message.Slot))
	if err != nil {
		return false
	}
	return bls.Verify(pub, ComputeSigningRoot(block.Message.HashTreeRoot(), domain), block.Signature)
}
