package shuffling

import (
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/shuffle"
)

type ShufflingFeature struct {
	Meta interface {
		VersioningMeta
		SeedMeta
		ActiveIndicesMeta
		CommitteeCountMeta
		CrosslinkTimingMeta
	}
}

// TODO: may want to pool this to avoid large allocations in mainnet.
type ShufflingStatus struct {
	Previous *ShufflingEpoch
	Current  *ShufflingEpoch
}

func (shs *ShufflingStatus) GetCrosslinkCommittee(epoch Epoch, shard Shard) []ValidatorIndex {
	if shard >= SHARD_COUNT {
		// sanity check for development, method should only used for previous and current epoch.
		panic(fmt.Errorf("crosslink committee retrieval: out of range shard: %d", shard))
	}
	if epoch == shs.Current.Epoch {
		return shs.Current.Committees[shard]
	} else if epoch == shs.Previous.Epoch {
		return shs.Previous.Committees[shard]
	} else {
		panic(fmt.Errorf("crosslink committee retrieval: out of range epoch: %d", epoch))
	}
}

func (state *ShufflingFeature) LoadShufflingStatus() *ShufflingStatus {
	return &ShufflingStatus{
		Previous: state.LoadShufflingEpoch(state.Meta.PreviousEpoch()),
		Current: state.LoadShufflingEpoch(state.Meta.CurrentEpoch()),
	}
}

// With a high amount of shards, or low amount of validators,
// some shards may not have a committee this epoch.
type ShufflingEpoch struct {
	Epoch Epoch
	Shuffling  []ValidatorIndex              // the active validator indices, shuffled into their committee
	Committees [SHARD_COUNT][]ValidatorIndex // slices of Shuffling, 1 per slot. Committee can be nil slice.
}

func (state *ShufflingFeature) LoadShufflingEpoch(epoch Epoch) *ShufflingEpoch {
	shep := &ShufflingEpoch{
		Epoch: epoch,
	}
	currentEpoch := state.Meta.CurrentEpoch()
	previousEpoch := state.Meta.PreviousEpoch()
	nextEpoch := currentEpoch + 1

	if !(previousEpoch <= epoch && epoch <= nextEpoch) {
		panic("could not compute shuffling for out of range epoch")
	}

	seed := state.Meta.GetSeed(epoch)
	activeIndices := state.Meta.GetActiveValidatorIndices(epoch)
	shuffle.UnshuffleList(activeIndices, seed)
	shep.Shuffling = activeIndices

	validatorCount := uint64(len(activeIndices))
	committeeCount := state.Meta.GetCommitteeCount(epoch)
	if committeeCount > uint64(SHARD_COUNT) {
		panic("too many committees")
	}
	startShard := state.Meta.GetStartShard(epoch)
	for i := uint64(0); i < committeeCount; i++ {
		shard := startShard + Shard(i)
		startOffset := (validatorCount * i) / committeeCount
		endOffset := (validatorCount * (i + 1)) / committeeCount
		committee := shep.Shuffling[startOffset:endOffset]
		shep.Committees[shard] = committee
	}
	return shep
}
