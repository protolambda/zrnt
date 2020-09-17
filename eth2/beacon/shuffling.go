package beacon

// With a high amount of shards, or low amount of validators,
// some shards may not have a committee this epoch.
type ShufflingEpoch struct {
	Epoch         Epoch
	ActiveIndices []ValidatorIndex
	Shuffling     []ValidatorIndex // the active validator indices, shuffled into their committee
	// slot (vector SLOTS_PER_EPOCH) -> index of committee (< MAX_COMMITTEES_PER_SLOT) -> index of validator within committee -> validator
	Committees [][][]ValidatorIndex // slices of Shuffling, 1 per slot. Committee can be nil slice.
}

func (spec *Spec) CommitteeCount(activeValidators uint64) uint64 {
	validatorsPerSlot := activeValidators / uint64(spec.SLOTS_PER_EPOCH)
	committeesPerSlot := validatorsPerSlot / spec.TARGET_COMMITTEE_SIZE
	if spec.MAX_COMMITTEES_PER_SLOT < committeesPerSlot {
		committeesPerSlot = spec.MAX_COMMITTEES_PER_SLOT
	}
	if committeesPerSlot == 0 {
		committeesPerSlot = 1
	}
	return committeesPerSlot
}

func (spec *Spec) ShufflingEpoch(state *BeaconStateView, indicesBounded []BoundedIndex, epoch Epoch) (*ShufflingEpoch, error) {
	mixes, err := state.RandaoMixes()
	if err != nil {
		return nil, err
	}
	seed, err := spec.GetSeed(mixes, epoch, spec.DOMAIN_BEACON_ATTESTER)
	if err != nil {
		return nil, err
	}
	return spec.NewShufflingEpoch(indicesBounded, seed, epoch), nil
}

func (spec *Spec) NewShufflingEpoch(indicesBounded []BoundedIndex, seed Root, epoch Epoch) *ShufflingEpoch {
	shep := &ShufflingEpoch{
		Epoch: epoch,
	}

	shep.ActiveIndices = make([]ValidatorIndex, 0, len(indicesBounded))
	for _, v := range indicesBounded {
		if v.Activation <= epoch && epoch < v.Exit {
			shep.ActiveIndices = append(shep.ActiveIndices, v.Index)
		}
	}

	// Copy over the active indices, then get the shuffling of them
	shep.Shuffling = make([]ValidatorIndex, len(shep.ActiveIndices), len(shep.ActiveIndices))
	for i, v := range shep.ActiveIndices {
		shep.Shuffling[i] = v
	}
	// shuffles the active indices into the shuffling
	// (name is misleading, unshuffle as a list results in original indices to be traced back to their functional committee position)
	UnshuffleList(spec.SHUFFLE_ROUND_COUNT, shep.Shuffling, seed)

	validatorCount := uint64(len(shep.Shuffling))
	committeesPerSlot := spec.CommitteeCount(validatorCount)
	committeeCount := committeesPerSlot * uint64(spec.SLOTS_PER_EPOCH)
	shep.Committees = make([][][]ValidatorIndex, spec.SLOTS_PER_EPOCH, spec.SLOTS_PER_EPOCH)
	for slot := uint64(0); slot < uint64(spec.SLOTS_PER_EPOCH); slot++ {
		shep.Committees[slot] = make([][]ValidatorIndex, 0, committeesPerSlot)
		for slotIndex := uint64(0); slotIndex < committeesPerSlot; slotIndex++ {
			index := (slot * committeesPerSlot) + slotIndex
			startOffset := (validatorCount * index) / committeeCount
			endOffset := (validatorCount * (index + 1)) / committeeCount
			committee := shep.Shuffling[startOffset:endOffset]
			shep.Committees[slot] = append(shep.Committees[slot], committee)
		}
	}
	return shep
}
