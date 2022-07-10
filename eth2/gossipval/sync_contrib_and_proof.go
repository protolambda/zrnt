package gossipval

import (
	"context"
	"errors"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type SyncContribAndProofValBackend interface {
	Spec
	Chain
	SlotAfter
	DomainGetter

	SeenContribution(aggregator common.ValidatorIndex, slot common.Slot, subnet uint64) bool
	MarkContribution(aggregator common.ValidatorIndex, slot common.Slot, subnet uint64)
}

func ValidateSyncContribAndProof(ctx context.Context, signedContribAndProof *altair.SignedContributionAndProof,
	scpVal SyncContribAndProofValBackend) ([]common.ValidatorIndex, GossipValidatorResult) {
	spec := scpVal.Spec()

	contribAndProof := &signedContribAndProof.Message
	contrib := &contribAndProof.Contribution

	// [IGNORE] The contribution's slot is for the current slot (with a MAXIMUM_GOSSIP_CLOCK_DISPARITY allowance), i.e. contribution.slot == current_slot.
	if err := CheckSlotSpan(scpVal.SlotAfter, contrib.Slot, 1); err != nil {
		return nil, GossipValidatorResult{IGNORE, fmt.Errorf("contribution not for current slot: %v", err)}
	}

	// [REJECT] The subcommittee index is in the allowed range, i.e. contribution.subcommittee_index < SYNC_COMMITTEE_SUBNET_COUNT.
	if contrib.SubcommitteeIndex >= common.SYNC_COMMITTEE_SUBNET_COUNT {
		return nil, GossipValidatorResult{REJECT, fmt.Errorf("sync sub committee index not within allowed range: %d", contrib.SubcommitteeIndex)}
	}

	// [REJECT] The contribution has participants -- that is, any(contribution.aggregation_bits).
	if count := contrib.AggregationBits.OnesCount(); count == 0 {
		return nil, GossipValidatorResult{REJECT, fmt.Errorf("sync committee contribution needs to have at least 1 participant")}
	}
	// [REJECT] contribution_and_proof.selection_proof selects the validator as an aggregator for the slot --
	// i.e. is_sync_committee_aggregator(contribution_and_proof.selection_proof) returns True.
	if !altair.IsSyncCommitteeAggregator(spec, contribAndProof.SelectionProof) {
		return nil, GossipValidatorResult{REJECT, errors.New("invalid sync aggregator selection proof")}
	}

	ch := scpVal.Chain()
	entry, ok := ch.ByBlockSlot(contrib.BeaconBlockRoot, contrib.Slot)
	if !ok {
		return nil, GossipValidatorResult{IGNORE, fmt.Errorf("cannot find beacon block that sync contribution contributes to")}
	}
	epc, err := entry.EpochsContext(ctx)
	if err != nil {
		return nil, GossipValidatorResult{IGNORE, err}
	}

	// [REJECT] The aggregator's validator index is in the declared subcommittee of the current sync committee --
	// i.e. state.validators[contribution_and_proof.aggregator_index].pubkey in get_sync_subcommittee_pubkeys(state, contribution.subcommittee_index).
	pubs, indices, err := epc.CurrentSyncCommittee.Subcommittee(spec, uint64(contrib.SubcommitteeIndex))
	if err != nil {
		return nil, GossipValidatorResult{REJECT, err}
	} else {
		found := false
		for _, valIndex := range indices {
			if valIndex == contribAndProof.AggregatorIndex {
				found = true
				break
			}
		}
		if !found {
			return nil, GossipValidatorResult{REJECT, fmt.Errorf("could not find aggregator %d in sync sub-committee %d at block %s",
				contribAndProof.AggregatorIndex, contrib.SubcommitteeIndex, contrib.BeaconBlockRoot)}
		}
	}

	// [IGNORE] The sync committee contribution is the first valid contribution received for the aggregator with index contribution_and_proof.aggregator_index
	// for the slot contribution.slot and subcommittee index contribution.subcommittee_index
	// (this requires maintaining a cache of size SYNC_COMMITTEE_SIZE for this topic that can be flushed after each slot).
	if scpVal.SeenContribution(contribAndProof.AggregatorIndex, contrib.Slot, uint64(contrib.SubcommitteeIndex)) {
		return nil, GossipValidatorResult{IGNORE, fmt.Errorf("already seen conribution of aggregator %d at slot %d for sync subnet %d",
			contribAndProof.AggregatorIndex, contrib.Slot, uint64(contrib.SubcommitteeIndex))}
	}

	// [REJECT] The contribution_and_proof.selection_proof is a valid signature of the SyncAggregatorSelectionData
	// derived from the contribution by the validator with index contribution_and_proof.aggregator_index.
	if err := altair.ValidateSyncAggregatorSelectionProof(spec, epc, scpVal.GetDomain,
		contribAndProof.AggregatorIndex, contribAndProof.SelectionProof,
		contrib.Slot, uint64(contrib.SubcommitteeIndex)); err != nil {
		return nil, GossipValidatorResult{REJECT, fmt.Errorf("invalid sync agg selection proof: %v", err)}
	}

	// [REJECT] The aggregator signature, signed_contribution_and_proof.signature, is valid.
	if err := signedContribAndProof.VerifySignature(spec, epc, scpVal.GetDomain); err != nil {
		return nil, GossipValidatorResult{REJECT, fmt.Errorf("invalid sync contribution aggregator signature: %v", err)}
	}

	// [REJECT] The aggregate signature is valid for the message beacon_block_root and aggregate pubkey
	// derived from the participation info in aggregation_bits for the subcommittee specified by the contribution.subcommittee_index.
	if err := contribAndProof.Contribution.VerifySignature(spec, pubs, scpVal.GetDomain); err != nil {
		return nil, GossipValidatorResult{REJECT, fmt.Errorf("invalid sync contribution signature: %v", err)}
	}

	scpVal.MarkContribution(contribAndProof.AggregatorIndex, contrib.Slot, uint64(contrib.SubcommitteeIndex))

	return indices, GossipValidatorResult{ACCEPT, nil}
}
