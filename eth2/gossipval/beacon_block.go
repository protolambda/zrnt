package gossipval

import (
	"context"
	"errors"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type BeaconBlockValBackend interface {
	Spec
	SlotAfter
	Chain
	GenesisValidatorsRoot

	// Checks if the (slot, proposer) pair was seen, does not do any tracking.
	SeenBlock(slot common.Slot, proposer common.ValidatorIndex) bool

	// When the block is fully validated (except proposer index check, but incl. signature check),
	// the combination can be marked as seen to avoid future duplicate blocks from being propagated.
	MarkBlock(slot common.Slot, proposer common.ValidatorIndex)
}

func ValidateBeaconBlock(ctx context.Context, block *common.BeaconBlockEnvelope,
	blockVal BeaconBlockValBackend) GossipValidatorResult {
	spec := blockVal.Spec()
	// [IGNORE] The block is not from a future slot (with a MAXIMUM_GOSSIP_CLOCK_DISPARITY allowance) --
	// i.e. validate that signed_beacon_block.message.slot <= current_slot
	if maxSlot := blockVal.SlotAfter(MAXIMUM_GOSSIP_CLOCK_DISPARITY); maxSlot < block.Slot {
		return GossipValidatorResult{IGNORE, fmt.Errorf("block slot %d is later than max slot %d", block.Slot, maxSlot)}
	}

	// [IGNORE] The block is the first block with valid signature received for the proposer for the slot, signed_beacon_block.message.slot.
	if blockVal.SeenBlock(block.Slot, block.ProposerIndex) {
		return GossipValidatorResult{IGNORE, fmt.Errorf("already seen a block for slot %d proposer %d", block.Slot, block.ProposerIndex)}
	}

	ch := blockVal.Chain()
	// [IGNORE] The block's parent (defined by block.parent_root) has been seen
	// (via both gossip and non-gossip sources)
	parentRef, ok := ch.ByBlock(block.ParentRoot)
	if !ok {
		return GossipValidatorResult{IGNORE, fmt.Errorf("block has unavailable parent block %s", block.ParentRoot)}
	}
	// Sanity check, implied condition
	if refSlot := parentRef.Step().Slot(); refSlot >= block.Slot {
		// It's OK to propagate, others do so and attack scope is limited, but it will not be processed later on. So just ignore it.
		return GossipValidatorResult{IGNORE, fmt.Errorf("block slot %d not after parent %d (%s)", block.Slot, refSlot, block.ParentRoot)}
	}

	// [IGNORE] The block is from a slot greater than the latest finalized slot --
	// i.e. validate that signed_beacon_block.message.slot > compute_start_slot_at_epoch(state.finalized_checkpoint.epoch)
	fin := ch.FinalizedCheckpoint()
	if finSlot, _ := spec.EpochStartSlot(fin.Epoch); block.Slot <= finSlot {
		return GossipValidatorResult{IGNORE, fmt.Errorf("block slot %d is not after finalized slot %d", block.Slot, finSlot)}
	}
	// [REJECT] The current finalized_checkpoint is an ancestor of block -- i.e. get_ancestor(store, block.parent_root, compute_start_slot_at_epoch(store.finalized_checkpoint.epoch)) == store.finalized_checkpoint.root
	if unknown, inSubtree := ch.InSubtree(fin.Root, block.ParentRoot); unknown {
		return GossipValidatorResult{IGNORE, fmt.Errorf("failed to determine if parent block %s is in subtree of finalized block %s", block.ParentRoot, fin.Root)}
	} else if !inSubtree {
		return GossipValidatorResult{REJECT, fmt.Errorf("parent block %s is not in subtree of finalized root %s", block.ParentRoot, fin.Root)}
	}

	// [REJECT] The block's parent (defined by block.parent_root) passes validation.
	// *implicit*: parent was already processed and put into forkchoice view, so it passes validation.

	parentEpc, err := parentRef.EpochsContext(ctx)
	if err != nil {
		return GossipValidatorResult{IGNORE, fmt.Errorf("cannot find context for parent block %s", block.ParentRoot)}
	}
	// [REJECT] The proposer signature, signed_beacon_block.signature, is valid with respect to the proposer_index pubkey.
	pub, ok := parentEpc.ValidatorPubkeyCache.Pubkey(block.ProposerIndex)
	if !ok {
		return GossipValidatorResult{IGNORE, fmt.Errorf("cannot find pubkey for proposer index %d", block.ProposerIndex)}
	}
	// Use untrusted proposer index, we validate this later, after signature check.
	if !block.VerifySignature(spec, blockVal.GenesisValidatorsRoot(), block.ProposerIndex, pub) {
		return GossipValidatorResult{REJECT, errors.New("invalid block signature")}
	}

	blockVal.MarkBlock(block.Slot, block.ProposerIndex)

	// [REJECT] The block is proposed by the expected proposer_index for the block's slot in the context of
	// the current shuffling (defined by parent_root/slot).

	targetEpoch := spec.SlotToEpoch(block.Slot)
	parentEpoch := spec.SlotToEpoch(parentRef.Step().Slot())
	var proposer common.ValidatorIndex
	if parentEpoch == targetEpoch {
		proposer, err = parentEpc.GetBeaconProposer(block.Slot)
		if err != nil {
			return GossipValidatorResult{IGNORE, fmt.Errorf("could not get proposer index for slot %d, from same epoch as parent block", block.Slot)}
		}
	} else if parentEpoch > targetEpoch {
		return GossipValidatorResult{REJECT, fmt.Errorf("expected parent epoch %d to not be after target %d", parentEpoch, targetEpoch)}
	} else {
		towardsCtx, cancel := context.WithTimeout(ctx, catchupTimeout)
		defer cancel()
		// the block slot was valid, so this must be valid.
		targetSlot, _ := spec.EpochStartSlot(targetEpoch)
		slotRef, err := ch.Towards(towardsCtx, block.ParentRoot, targetSlot)
		if err != nil {
			return GossipValidatorResult{IGNORE, fmt.Errorf("could not transition towards target: %v", err)}
		}
		slotEpc, err := slotRef.EpochsContext(ctx)
		if err != nil {
			return GossipValidatorResult{IGNORE, fmt.Errorf("could not fetch epochs context for slot reference: %v", err)}
		}
		proposer, err = slotEpc.GetBeaconProposer(block.Slot)
		if err != nil {
			return GossipValidatorResult{IGNORE, fmt.Errorf("could not fetch block proposer slot reference: %v", err)}
		}
	}

	if proposer != block.ProposerIndex {
		return GossipValidatorResult{REJECT, fmt.Errorf("expected proposer %d, but block was proposed by %d", proposer, block.ProposerIndex)}
	}

	return GossipValidatorResult{ACCEPT, nil}
}
