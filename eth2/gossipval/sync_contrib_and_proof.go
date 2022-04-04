package gossipval

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type SyncContribAndProofValBackend interface {
	Spec
	Chain
	SlotAfter

	SeenContribution(aggregator common.ValidatorIndex, slot common.Slot, commIndex uint64) bool
	MarkContribution(aggregator common.ValidatorIndex, slot common.Slot, commIndex uint64)
}

func ValidateSyncContribAndProof(ctx context.Context, signedContribAndProof *altair.SignedContributionAndProof,
	scpVal SyncContribAndProofValBackend) GossipValidatorResult {
	//spec := scpVal.Spec()

	contribAndProof := &signedContribAndProof.Message
	contrib := &contribAndProof.Contribution

	// [IGNORE] The contribution's slot is for the current slot (with a MAXIMUM_GOSSIP_CLOCK_DISPARITY allowance), i.e. contribution.slot == current_slot.
	// TODO: unify below with other range checks of gossip validators
	slot := contrib.Slot
	slotRange := common.Slot(1)
	if slot+ATTESTATION_PROPAGATION_SLOT_RANGE < slot {
		return GossipValidatorResult{REJECT, fmt.Errorf("slot overflow: %d", slot)}
	}
	// check minimum, with account for clock disparity
	if minSlot := scpVal.SlotAfter(-MAXIMUM_GOSSIP_CLOCK_DISPARITY); slot+slotRange < minSlot {
		return GossipValidatorResult{IGNORE, fmt.Errorf("slot %d is too old, minimum slot is %d", slot, minSlot)}
	}
	// check maximum, with account for clock disparity
	if maxSlot := scpVal.SlotAfter(MAXIMUM_GOSSIP_CLOCK_DISPARITY); slot > maxSlot {
		return GossipValidatorResult{IGNORE, fmt.Errorf("slot %d is too new, maximum slot is %d", slot, maxSlot)}
	}

	// [REJECT] The subcommittee index is in the allowed range, i.e. contribution.subcommittee_index < SYNC_COMMITTEE_SUBNET_COUNT.
	if contrib.SubcommitteeIndex >= common.SYNC_COMMITTEE_SUBNET_COUNT {
		return GossipValidatorResult{REJECT, fmt.Errorf("sync sub committee index not within allowed range: %d", contrib.SubcommitteeIndex)}
	}

	// TODO: more validation
	return GossipValidatorResult{ACCEPT, nil}
}
