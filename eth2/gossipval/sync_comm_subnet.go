package gossipval

import (
	"context"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type SyncCommitteeSubnetValBackend interface {
	Spec
	Chain
	SlotAfter
	DomainGetter

	SeenSyncCommMsg(validator common.ValidatorIndex, slot common.Slot, subnet uint64) bool
	MarkSyncCommMsg(validator common.ValidatorIndex, slot common.Slot, subnet uint64)
}

func ValidateSyncCommitteeSubnet(ctx context.Context, subnet uint64, syncCommMessage *altair.SyncCommitteeMessage,
	scpVal SyncCommitteeSubnetValBackend) ([]common.ValidatorIndex, GossipValidatorResult) {
	spec := scpVal.Spec()

	// [IGNORE] The message's slot is for the current slot (with a MAXIMUM_GOSSIP_CLOCK_DISPARITY allowance),
	// i.e. sync_committee_message.slot == current_slot.
	if err := CheckSlotSpan(scpVal.SlotAfter, syncCommMessage.Slot, 1); err != nil {
		return nil, GossipValidatorResult{IGNORE, fmt.Errorf("sync comm message not for current slot: %v", err)}
	}

	ch := scpVal.Chain()
	entry, ok := ch.ByBlockSlot(syncCommMessage.BeaconBlockRoot, syncCommMessage.Slot)
	if !ok {
		return nil, GossipValidatorResult{IGNORE, fmt.Errorf("cannot find beacon block that sync contribution contributes to")}
	}
	epc, err := entry.EpochsContext(ctx)
	if err != nil {
		return nil, GossipValidatorResult{IGNORE, err}
	}

	// [REJECT] The subnet_id is valid for the given validator,
	// i.e. subnet_id in compute_subnets_for_sync_committee(state, sync_committee_message.validator_index).
	// Note this validation implies the validator is part of the broader current sync committee along with the correct subcommittee.
	if !epc.CurrentSyncCommittee.InSubnet(spec, syncCommMessage.ValidatorIndex, subnet) {
		return nil, GossipValidatorResult{REJECT, fmt.Errorf("validator %d is not in sync committee subnet %d at slot %d",
			syncCommMessage.ValidatorIndex, subnet, syncCommMessage.Slot)}
	}

	// [IGNORE] There has been no other valid sync committee message for the declared slot for the validator referenced by sync_committee_message.validator_index
	// (this requires maintaining a cache of size SYNC_COMMITTEE_SIZE // SYNC_COMMITTEE_SUBNET_COUNT for each subnet that can be flushed after each slot).
	// Note this validation is per topic so that for a given slot, multiple messages could be forwarded with the same validator_index as long as the subnet_ids are distinct.
	if scpVal.SeenSyncCommMsg(syncCommMessage.ValidatorIndex, syncCommMessage.Slot, subnet) {
		return nil, GossipValidatorResult{IGNORE, fmt.Errorf("already seen validator %d contribute to sync subnet %d at slot %d",
			syncCommMessage.ValidatorIndex, subnet, syncCommMessage.Slot)}
	}

	// [REJECT] The signature is valid for the message beacon_block_root for the validator referenced by validator_index.
	if err := syncCommMessage.VerifySignature(spec, epc, scpVal.GetDomain); err != nil {
		return nil, GossipValidatorResult{REJECT, fmt.Errorf("invalid sync committee signature from validator %d subnet %d slot %d: %v",
			syncCommMessage.ValidatorIndex, subnet, syncCommMessage.Slot, err)}
	}

	scpVal.MarkSyncCommMsg(syncCommMessage.ValidatorIndex, syncCommMessage.Slot, subnet)

	return epc.CurrentSyncCommittee.Indices, GossipValidatorResult{ACCEPT, nil}
}
