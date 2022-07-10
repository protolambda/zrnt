package gossipval

import (
	"context"
	"errors"
	"fmt"

	blsu "github.com/protolambda/bls12-381-util"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"

	"github.com/protolambda/ztyp/tree"
)

type AggregatesValBackend interface {
	Spec
	Chain
	SlotAfter
	BadBlockValidator

	// Checks if the aggregate attestation defined by aggRoot = hash_tree_root(aggregate) has been seen
	// (via aggregate gossip, within a verified block, or through the creation of an equivalent aggregate locally).
	SeenAggregate(aggRoot common.Root) bool
	MarkAggregate(aggRoot common.Root)

	// Checks if an aggregate by the given aggregator for the given epoch has been seen before.
	SeenAggregator(targetEpoch common.Epoch, aggregator common.ValidatorIndex) bool
	MarkAggregator(targetEpoch common.Epoch, aggregator common.ValidatorIndex)
}

func ValidateAggregateAndProof(ctx context.Context, signedAgg *phase0.SignedAggregateAndProof,
	aggVal AggregatesValBackend) ([]common.ValidatorIndex, GossipValidatorResult) {
	spec := aggVal.Spec()
	// [IGNORE] aggregate.data.slot is within the last ATTESTATION_PROPAGATION_SLOT_RANGE
	// slots (with a MAXIMUM_GOSSIP_CLOCK_DISPARITY allowance) --
	// i.e. aggregate.data.slot + ATTESTATION_PROPAGATION_SLOT_RANGE >= current_slot >= aggregate.data.slot
	// overflow check
	att := &signedAgg.Message.Aggregate
	if err := CheckSlotSpan(aggVal.SlotAfter, att.Data.Slot, ATTESTATION_PROPAGATION_SLOT_RANGE); err != nil {
		return nil, GossipValidatorResult{IGNORE, fmt.Errorf("aggregate attestation not within slot range: %v", err)}
	}

	// [REJECT] The aggregate attestation's epoch matches its target --
	// i.e. aggregate.data.target.epoch == compute_epoch_at_slot(aggregate.data.slot)
	attEpoch := spec.SlotToEpoch(att.Data.Slot)
	if att.Data.Target.Epoch != attEpoch {
		return nil, GossipValidatorResult{REJECT, fmt.Errorf("attestation slot %d is epoch %d and does not match target %d", att.Data.Slot, attEpoch, att.Data.Target.Epoch)}
	}

	// [IGNORE] The aggregate is the first valid aggregate received for the aggregator with index
	// aggregate_and_proof.aggregator_index for the epoch aggregate.data.target.epoch.
	if epoch, index := att.Data.Target.Epoch, signedAgg.Message.AggregatorIndex; aggVal.SeenAggregator(epoch, index) {
		return nil, GossipValidatorResult{IGNORE, fmt.Errorf("already seen aggregate by %d for epoch %d", index, epoch)}
	}

	// [IGNORE] The valid aggregate attestation defined by hash_tree_root(aggregate) has not already been seen
	// (via aggregate gossip, within a verified block, or through the creation of an equivalent aggregate locally).
	aggRoot := att.HashTreeRoot(spec, tree.GetHashFn())
	if aggVal.SeenAggregate(aggRoot) {
		return nil, GossipValidatorResult{IGNORE, fmt.Errorf("attestation aggregate %s has already been seen", aggRoot)}
	}

	// [REJECT] The attestation has participants --
	// i.e., len(get_attesting_indices(state, aggregate.data, aggregate.aggregation_bits)) >= 1.
	if att.AggregationBits.OnesCount() < 1 {
		return nil, GossipValidatorResult{REJECT, fmt.Errorf("attestation has no participants")}
	}

	// [IGNORE] The block being voted for (aggregate.data.beacon_block_root) has been seen (via both gossip and non-gossip sources)
	// (a client MAY queue aggregates for processing once block is retrieved).
	// TODO

	// [REJECT] The block being voted for (aggregate.data.beacon_block_root) passes validation.
	if aggVal.IsBadBlock(att.Data.BeaconBlockRoot) {
		return nil, GossipValidatorResult{REJECT, errors.New("aggregate voted for invalid block")}
	}

	ch := aggVal.Chain()

	// [REJECT] The current finalized_checkpoint is an ancestor of the block defined
	// by aggregate.data.beacon_block_root --
	// i.e. get_ancestor(store, attestation.data.beacon_block_root, compute_start_slot_at_epoch(store.finalized_checkpoint.epoch))
	//        == store.finalized_checkpoint.root
	fin := ch.FinalizedCheckpoint()
	if att.Data.BeaconBlockRoot != fin.Root {
		if unknown, inSubtree := ch.InSubtree(fin.Root, att.Data.BeaconBlockRoot); unknown {
			return nil, GossipValidatorResult{IGNORE, errors.New("unknown block, cannot check if in subtree")}
		} else if !inSubtree {
			return nil, GossipValidatorResult{IGNORE, errors.New("block not in subtree of finalized root")}
		}
	} else if fin.Epoch > att.Data.Target.Epoch {
		return nil, GossipValidatorResult{REJECT, errors.New("cannot vote for finalized root as target")}
	}

	// 3 combined steps:
	// [REJECT] aggregate_and_proof.selection_proof selects the validator as an aggregator for the slot --
	// i.e. is_aggregator(state, aggregate.data.slot, aggregate.data.index, aggregate_and_proof.selection_proof) returns True.
	// [REJECT] The aggregator's validator index is within the committee --
	// i.e. aggregate_and_proof.aggregator_index in get_beacon_committee(state, aggregate.data.slot, aggregate.data.index).
	// [REJECT] The aggregate_and_proof.selection_proof is a valid signature of the aggregate.data.slot
	// by the validator with index aggregate_and_proof.aggregator_index.

	// target epoch was already validated to match the slot, which was validated to be within normal range. No overflows.
	startSlot, _ := spec.EpochStartSlot(att.Data.Target.Epoch)

	towardsCtx, cancel := context.WithTimeout(ctx, catchupTimeout)
	defer cancel()

	entry, err := ch.Towards(towardsCtx, att.Data.Target.Root, startSlot)
	if err != nil {
		return nil, GossipValidatorResult{IGNORE, err}
	}
	epc, err := entry.EpochsContext(ctx)
	if err != nil {
		return nil, GossipValidatorResult{IGNORE, err}
	}
	state, err := entry.State(ctx)
	if err != nil {
		return nil, GossipValidatorResult{IGNORE, err}
	}
	if valid, err := phase0.ValidateAggregateSelectionProof(spec, epc, state, att.Data.Slot, att.Data.Index, signedAgg.Message.AggregatorIndex, signedAgg.Message.SelectionProof); err != nil {
		return nil, GossipValidatorResult{IGNORE, err}
	} else if !valid {
		return nil, GossipValidatorResult{REJECT, errors.New("invalid aggregate")}
	}

	// [REJECT] The aggregator signature, signed_aggregate_and_proof.signature, is valid.
	dom, err := common.GetDomain(state, common.DOMAIN_AGGREGATE_AND_PROOF, att.Data.Target.Epoch)
	if err != nil {
		return nil, GossipValidatorResult{IGNORE, err}
	}
	sigRoot := common.ComputeSigningRoot(signedAgg.Message.HashTreeRoot(spec, tree.GetHashFn()), dom)
	pub, ok := epc.ValidatorPubkeyCache.Pubkey(signedAgg.Message.AggregatorIndex)
	if !ok {
		return nil, GossipValidatorResult{IGNORE, fmt.Errorf("missing pubkey: %d", signedAgg.Message.AggregatorIndex)}
	}
	blsPub, err := pub.Pubkey()
	if err != nil {
		return nil, GossipValidatorResult{IGNORE, fmt.Errorf("failed to deserialize cached pubkey: %v", err)}
	}
	sig, err := signedAgg.Signature.Signature()
	if err != nil {
		return nil, GossipValidatorResult{REJECT, fmt.Errorf("failed to deserialize aggregate signature: %v", err)}
	}
	if !blsu.Verify(blsPub, sigRoot[:2], sig) {
		return nil, GossipValidatorResult{REJECT, errors.New("invalid aggregate signature")}
	}

	// [REJECT] The signature of aggregate is valid.
	// Check signature and bitfields
	committee, err := epc.GetBeaconCommittee(att.Data.Slot, att.Data.Index)
	if err != nil {
		return nil, GossipValidatorResult{IGNORE, err}
	}
	if indexedAtt, err := att.ConvertToIndexed(spec, committee); err != nil {
		// it should always convert.
		// Something is very wrong if not, e.g. bad bitfield length.
		return nil, GossipValidatorResult{REJECT, err}
	} else if err := phase0.ValidateIndexedAttestation(spec, epc, state, indexedAtt); err != nil {
		return nil, GossipValidatorResult{REJECT, err}
	}

	aggVal.MarkAggregate(aggRoot)
	aggVal.MarkAggregator(att.Data.Target.Epoch, signedAgg.Message.AggregatorIndex)

	return committee, GossipValidatorResult{ACCEPT, nil}
}
