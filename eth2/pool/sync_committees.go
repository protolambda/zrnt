package pool

import (
	"context"
	"fmt"
	"sync"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

// beacon root -> subnet -> contributions
type SyncCommitteeContributions map[common.Root]map[uint64][]*SubnetContrib

type SyncCommitteeMessages map[common.ValidatorIndex]*altair.SyncCommitteeMessage

func (msgs SyncCommitteeMessages) Select(root common.Root, members []common.ValidatorIndex) []*altair.SyncCommitteeMessage {
	out := make([]*altair.SyncCommitteeMessage, 0, len(members))
	for _, vi := range members {
		msg := msgs[vi]
		if msg.BeaconBlockRoot == root {
			out = append(out, msg)
		}
	}
	return out
}

type SubnetContrib struct {
	// A bit is set if a signature from the validator at the corresponding
	// index in the subcommittee is present in the aggregate `signature`.
	AggregationBits altair.SyncCommitteeSubnetBits
	// Signature by the validator(s) over the block root of `slot`
	Signature common.BLSSignature
}

// SyncCommitteePool is a very short lived buffer:
// - The sync committee messages of the previous, current and next slot are buffered
// - The sync committee contributions (subnet aggregates) of the previous, current and next slot are buffered
// - As soon as a slot is done, Reset(slot) should be called to transition to a new slot, rotating out buffers.
// - Nothing is aggregated ahead of time; packing work is avoided if we are not selected as aggregator
// - At any time the validator can run PackAggregate and PackContribution for the approximate current slot (previous/current/next slot accepted) and beacon block root.
type SyncCommitteePool struct {
	sync.Mutex

	spec *common.Spec

	currentSlot common.Slot

	prevContribs    SyncCommitteeContributions
	currentContribs SyncCommitteeContributions
	nextContribs    SyncCommitteeContributions

	prevMsgs    SyncCommitteeMessages
	currentMsgs SyncCommitteeMessages
	nextMsgs    SyncCommitteeMessages
}

func NewSyncCommitteePool(spec *common.Spec) *SyncCommitteePool {
	return &SyncCommitteePool{
		spec:        spec,
		currentSlot: ^common.Slot(0),
	}
}

func (sp *SyncCommitteePool) AddSyncCommitteeContribution(ctx context.Context, contrib *altair.SyncCommitteeContribution) error {
	sp.Lock()
	defer sp.Unlock()
	var subsByRoot map[common.Root]map[uint64][]*SubnetContrib
	if sp.currentSlot == contrib.Slot+1 {
		subsByRoot = sp.prevContribs
	} else if sp.currentSlot == contrib.Slot {
		subsByRoot = sp.currentContribs
	} else if sp.currentSlot+1 == contrib.Slot {
		subsByRoot = sp.nextContribs
	} else {
		return fmt.Errorf("current sync committee pool is at slot %d, cannot process contribution for slot %d", sp.currentSlot, contrib.Slot)
	}
	subs, ok := subsByRoot[contrib.BeaconBlockRoot]
	if !ok {
		subs = make(map[uint64][]*SubnetContrib)
		subsByRoot[contrib.BeaconBlockRoot] = subs
	}
	subs[uint64(contrib.SubcommitteeIndex)] = append(subs[uint64(contrib.SubcommitteeIndex)], &SubnetContrib{
		AggregationBits: contrib.AggregationBits,
		Signature:       contrib.Signature,
	})
	return nil
}

func (sp *SyncCommitteePool) AddSyncCommitteeMessage(ctx context.Context, msg *altair.SyncCommitteeMessage) error {
	sp.Lock()
	defer sp.Unlock()
	if sp.currentSlot == msg.Slot+1 {
		sp.prevMsgs[msg.ValidatorIndex] = msg
	} else if sp.currentSlot == msg.Slot {
		sp.currentMsgs[msg.ValidatorIndex] = msg
	} else if sp.currentSlot+1 == msg.Slot {
		sp.nextMsgs[msg.ValidatorIndex] = msg
	} else {
		return fmt.Errorf("current sync committee pool is at slot %d, cannot process message for slot %d", sp.currentSlot, msg.Slot)
	}
	return nil
}

func (sp *SyncCommitteePool) PackContribution(ctx context.Context, slot common.Slot, beaconBlockRoot common.Root, subnet uint64, subComm []common.ValidatorIndex) (*altair.SyncCommitteeContribution, error) {
	sp.Lock()
	defer sp.Unlock()
	// TODO: select messages & aggregate a contribution
	return nil, nil
}

func (sp *SyncCommitteePool) PackAggregate(ctx context.Context, slot common.Slot, beaconBlockRoot common.Root, syncCommittee []common.ValidatorIndex) (*altair.SyncAggregate, error) {
	sp.Lock()
	defer sp.Unlock()
	// TODO: select complimentary subnet contributions, fill gaps with any individual messages, and return it
	return nil, nil
}

func (sp *SyncCommitteePool) Reset(slot common.Slot) {
	if sp.currentSlot == slot+1 {
		sp.nextMsgs = sp.currentMsgs
		sp.currentMsgs = sp.prevMsgs
		sp.prevMsgs = make(SyncCommitteeMessages, sp.spec.SYNC_COMMITTEE_SIZE)

		sp.nextContribs = sp.currentContribs
		sp.currentContribs = sp.prevContribs
		sp.prevContribs = make(SyncCommitteeContributions)
	} else if sp.currentSlot == slot {
		return
	} else if sp.currentSlot+1 == slot {
		sp.prevMsgs = sp.currentMsgs
		sp.currentMsgs = sp.nextMsgs
		sp.nextMsgs = make(SyncCommitteeMessages, sp.spec.SYNC_COMMITTEE_SIZE)

		sp.prevContribs = sp.currentContribs
		sp.currentContribs = sp.nextContribs
		sp.nextContribs = make(SyncCommitteeContributions)
	} else {
		sp.prevMsgs = make(SyncCommitteeMessages, sp.spec.SYNC_COMMITTEE_SIZE)
		sp.currentMsgs = make(SyncCommitteeMessages, sp.spec.SYNC_COMMITTEE_SIZE)
		sp.nextMsgs = make(SyncCommitteeMessages, sp.spec.SYNC_COMMITTEE_SIZE)

		sp.prevContribs = make(SyncCommitteeContributions)
		sp.currentContribs = make(SyncCommitteeContributions)
		sp.nextContribs = make(SyncCommitteeContributions)
	}
	sp.currentSlot = slot
}
