package status

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/shuffling"
	"sort"
)

type ShufflingStatus struct {
	Current  ShufflingEpoch
	Previous ShufflingEpoch
}

func (status *ShufflingStatus) Load(state *BeaconState) {
	status.Current.Load(state, state.Epoch())
	status.Current.Load(state, state.PreviousEpoch())
}

type ShufflingEpoch struct {
	Shuffling       []ValidatorIndex              // the active validator indices, shuffled into their committee
	Committees      [SHARD_COUNT][]ValidatorIndex // slices of Shuffling, 1 per slot
	CommitteeSorted [SHARD_COUNT]ValidatorSet     // copy of Committees, with sorted indices for each committee
}

func (shep *ShufflingEpoch) Load(state *BeaconState, epoch Epoch) {
	currentEpoch := state.Epoch()
	previousEpoch := state.PreviousEpoch()
	nextEpoch := currentEpoch + 1

	if !(previousEpoch <= epoch && epoch <= nextEpoch) {
		panic("could not compute shuffling for out of range epoch")
	}

	seed := state.GenerateSeed(epoch)
	activeIndices := state.Validators.GetActiveValidatorIndices(epoch)
	shuffling.UnshuffleList(activeIndices, seed)
	shep.Shuffling = activeIndices

	validatorCount := uint64(len(activeIndices))
	committeeCount := state.Validators.GetEpochCommitteeCount(epoch)
	startShard := state.GetEpochStartShard(epoch)
	for shard := Shard(0); shard < SHARD_COUNT; shard++ {
		index := uint64((shard + SHARD_COUNT - startShard) % SHARD_COUNT)
		startOffset := (validatorCount * index) / committeeCount
		endOffset := (validatorCount * (index + 1)) / committeeCount
		committee := shep.Shuffling[startOffset:endOffset]
		shep.Committees[shard] = committee
		committeeSorted := make(ValidatorSet, len(committee), len(committee))
		copy(committeeSorted, committee)
		sort.Sort(committeeSorted)
		shep.Committees[shard] = committeeSorted
	}
}
