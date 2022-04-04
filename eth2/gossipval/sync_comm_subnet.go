package gossipval

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type SyncCommitteeSubnetValBackend interface {
	Spec
	Chain
	SlotAfter

	SeenSyncCommMsg(validator common.ValidatorIndex, slot common.Slot, subnet uint64) bool
	MarkSyncCommMsg(validator common.ValidatorIndex, slot common.Slot, subnet uint64) bool
}

func ValidateSyncCommitteeSubnet(ctx context.Context, subnet uint64, syncCommMessage *altair.SyncCommitteeMessage,
	scpVal SyncCommitteeSubnetValBackend) GossipValidatorResult {
	//spec := scpVal.Spec()

	// TODO: validation
	return GossipValidatorResult{ACCEPT, nil}
}
