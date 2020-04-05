package beacon

import (
	"errors"
	"fmt"


	"github.com/protolambda/zrnt/eth2/util/shuffle"
)

// TODO: may want to pool this to avoid large allocations in mainnet.
type ShufflingStatus struct {
	PreviousShuffling *ShufflingEpoch
	CurrentShuffling  *ShufflingEpoch
	NextShuffling     *ShufflingEpoch
}

// Return the beacon committee at slot for index.
func (shs *ShufflingStatus) GetBeaconCommittee(slot Slot, index CommitteeIndex) ([]ValidatorIndex, error) {
	if index >= MAX_COMMITTEES_PER_SLOT {
		return nil, fmt.Errorf("crosslink committee retrieval: out of range committee index: %d", index)
	}

	epoch := slot.ToEpoch()
	epochSlot := slot%SLOTS_PER_EPOCH
	var slotComms [][]ValidatorIndex
	if epoch == shs.PreviousShuffling.Epoch {
		slotComms = shs.PreviousShuffling.Committees[epochSlot]
	} else if epoch == shs.CurrentShuffling.Epoch {
		slotComms = shs.CurrentShuffling.Committees[epochSlot]
	} else if epoch == shs.NextShuffling.Epoch {
		slotComms = shs.NextShuffling.Committees[epochSlot]
	} else {
		return nil, fmt.Errorf("crosslink committee retrieval: out of range epoch: %d", epoch)
	}
	if index >= CommitteeIndex(len(slotComms)) {
		return nil, fmt.Errorf("crosslink committee retrieval: out of range committee index: %d", index)
	}
	return slotComms[index], nil
}

func (shs *ShufflingStatus) GetCommitteeCountAtSlot(slot Slot) (uint64, error) {
	epoch := slot.ToEpoch()
	epochSlot := slot%SLOTS_PER_EPOCH
	if epoch == shs.PreviousShuffling.Epoch {
		return uint64(len(shs.PreviousShuffling.Committees[epochSlot])), nil
	} else if epoch == shs.CurrentShuffling.Epoch {
		return uint64(len(shs.CurrentShuffling.Committees[epochSlot])), nil
	} else if epoch == shs.NextShuffling.Epoch {
		return uint64(len(shs.NextShuffling.Committees[epochSlot])), nil
	} else {
		return 0, fmt.Errorf("crosslink committee retrieval: out of range epoch: %d", epoch)
	}
}

func LoadShufflingStatus(input PrepareShufflingInput) (*ShufflingStatus, error) {
	currentEpoch, err := input.CurrentEpoch()
	if err != nil {
		return nil, err
	}
	previousEpoch, err := input.PreviousEpoch()
	if err != nil {
		return nil, err
	}
	nextEpoch := currentEpoch + 1

	prevSh, err := LoadShufflingEpoch(input, previousEpoch)
	if err != nil {
		return nil, err
	}
	currSh, err := LoadShufflingEpoch(input, currentEpoch)
	if err != nil {
		return nil, err
	}
	nextSh, err := LoadShufflingEpoch(input, nextEpoch)
	if err != nil {
		return nil, err
	}
	return &ShufflingStatus{
		PreviousShuffling: prevSh,
		CurrentShuffling:  currSh,
		NextShuffling:     nextSh,
	}, nil
}

// With a high amount of shards, or low amount of validators,
// some shards may not have a committee this epoch.
type ShufflingEpoch struct {
	Epoch      Epoch
	Shuffling  []ValidatorIndex                    // the active validator indices, shuffled into their committee
	// slot -> index of committee (< MAX_COMMITTEES_PER_SLOT) -> index of validator within committee -> validator
	Committees [SLOTS_PER_EPOCH][][]ValidatorIndex // slices of Shuffling, 1 per slot. Committee can be nil slice.
}

func CommitteeCount(activeValidators uint64) uint64 {
	validatorsPerSlot := activeValidators / uint64(SLOTS_PER_EPOCH)
	committeesPerSlot := validatorsPerSlot / TARGET_COMMITTEE_SIZE
	if MAX_COMMITTEES_PER_SLOT < committeesPerSlot {
		committeesPerSlot = MAX_COMMITTEES_PER_SLOT
	}
	if committeesPerSlot == 0 {
		committeesPerSlot = 1
	}
	return committeesPerSlot
}

func LoadShufflingEpoch(input PrepareShufflingInput, epoch Epoch) (*ShufflingEpoch, error) {
	shep := &ShufflingEpoch{
		Epoch: epoch,
	}
	currentEpoch, err := input.CurrentEpoch()
	if err != nil {
		return nil, err
	}
	previousEpoch, err := input.PreviousEpoch()
	if err != nil {
		return nil, err
	}
	nextEpoch := currentEpoch + 1

	if !(previousEpoch <= epoch && epoch <= nextEpoch) {
		return nil, errors.New("could not compute shuffling for out of range epoch")
	}

	seed, err := input.GetSeed(epoch, DOMAIN_BEACON_ATTESTER)
	if err != nil {
		return nil, err
	}
	activeIndices, err := input.GetActiveValidatorIndices(epoch)
	if err != nil {
		return nil, err
	}
	shuffle.UnshuffleList(activeIndices, seed)
	shep.Shuffling = activeIndices

	validatorCount := uint64(len(activeIndices))
	committeesPerSlot := CommitteeCount(validatorCount)
	if committeesPerSlot > uint64(MAX_COMMITTEES_PER_SLOT) {
		return nil, errors.New("too many committees per slot")
	}
	committeeCount := committeesPerSlot * uint64(SLOTS_PER_EPOCH)
	for slot := uint64(0); slot < uint64(SLOTS_PER_EPOCH); slot++ {
		shep.Committees[slot] = make([][]ValidatorIndex, 0, committeesPerSlot)
		for slotIndex := uint64(0); slotIndex < committeesPerSlot; slotIndex++ {
			index := (slot * committeesPerSlot) + slotIndex
			startOffset := (validatorCount * index) / committeeCount
			endOffset := (validatorCount * (index + 1)) / committeeCount
			if startOffset == endOffset {
				return nil, errors.New("cannot compute shuffling, empty committee")
			}
			committee := shep.Shuffling[startOffset:endOffset]
			shep.Committees[slot] = append(shep.Committees[slot], committee)
		}
	}
	return shep, nil
}
