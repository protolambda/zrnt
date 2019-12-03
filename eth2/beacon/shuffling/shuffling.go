package shuffling

import (
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/shuffle"
)

type ShufflingFeature struct {
	Meta interface {
		meta.Versioning
		meta.EpochSeed
		meta.ActiveIndices
		meta.CommitteeCount
	}
}

// TODO: may want to pool this to avoid large allocations in mainnet.
type ShufflingStatus struct {
	PreviousShuffling *ShufflingEpoch
	CurrentShuffling  *ShufflingEpoch
	NextShuffling     *ShufflingEpoch
}

// Return the beacon committee at slot for index.
func (shs *ShufflingStatus) GetBeaconCommittee(slot Slot, index CommitteeIndex) []ValidatorIndex {
	if index >= MAX_COMMITTEES_PER_SLOT {
		panic(fmt.Errorf("beacon committee retrieval: out of range committee index: %d", index))
	}

	epoch := slot.ToEpoch()
	if epoch == shs.PreviousShuffling.Epoch {
		return shs.PreviousShuffling.Committees[slot % SLOTS_PER_EPOCH][index]
	} else if epoch == shs.CurrentShuffling.Epoch {
		return shs.CurrentShuffling.Committees[slot % SLOTS_PER_EPOCH][index]
	} else if epoch == shs.NextShuffling.Epoch {
		return shs.NextShuffling.Committees[slot % SLOTS_PER_EPOCH][index]
	} else {
		panic(fmt.Errorf("beacon committee retrieval: out of range epoch: %d", epoch))
	}
}

func (f *ShufflingFeature) LoadShufflingStatus() *ShufflingStatus {
	currentEpoch := f.Meta.CurrentEpoch()
	previousEpoch := f.Meta.PreviousEpoch()
	nextEpoch := currentEpoch + 1

	return &ShufflingStatus{
		PreviousShuffling: f.LoadShufflingEpoch(currentEpoch),
		CurrentShuffling:  f.LoadShufflingEpoch(previousEpoch),
		NextShuffling:     f.LoadShufflingEpoch(nextEpoch),
	}
}

// With a high amount of shards, or low amount of validators,
// some shards may not have a committee this epoch.
type ShufflingEpoch struct {
	Epoch      Epoch
	Shuffling  []ValidatorIndex              // the active validator indices, shuffled into their committee
	Committees [SLOTS_PER_EPOCH][MAX_COMMITTEES_PER_SLOT][]ValidatorIndex // slices of Shuffling, 1 per slot. Committee can be nil slice.
}

func (f *ShufflingFeature) LoadShufflingEpoch(epoch Epoch) *ShufflingEpoch {
	shep := &ShufflingEpoch{
		Epoch: epoch,
	}
	currentEpoch := f.Meta.CurrentEpoch()
	previousEpoch := f.Meta.PreviousEpoch()
	nextEpoch := currentEpoch + 1

	if !(previousEpoch <= epoch && epoch <= nextEpoch) {
		panic("could not compute shuffling for out of range epoch")
	}

	seed := f.Meta.GetSeed(epoch, DOMAIN_BEACON_ATTESTER)
	activeIndices := f.Meta.GetActiveValidatorIndices(epoch)
	shuffle.UnshuffleList(activeIndices, seed)
	shep.Shuffling = activeIndices

	validatorCount := uint64(len(activeIndices))
	committeesPerSlot := f.Meta.GetCommitteeCountAtSlot(epoch.GetStartSlot())
	if committeesPerSlot > uint64(MAX_COMMITTEES_PER_SLOT) {
		panic("too many committees per slot")
	}
	committeeCount := committeesPerSlot * uint64(SLOTS_PER_EPOCH)
	for slot := uint64(0); slot < uint64(SLOTS_PER_EPOCH); slot++ {
		for slotIndex := uint64(0); slotIndex < committeesPerSlot; slotIndex++ {
			index := (slot * committeesPerSlot) + slotIndex
			startOffset := (validatorCount * index) / committeeCount
			endOffset := (validatorCount * (index + 1)) / committeeCount
			if startOffset == endOffset {
				panic("empty committee")
			}
			committee := shep.Shuffling[startOffset:endOffset]
			shep.Committees[slot][slotIndex] = committee
		}
	}
	return shep
}
