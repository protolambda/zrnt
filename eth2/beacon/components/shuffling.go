package components

import (
	"encoding/binary"
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/shuffling"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

// Randomness and committees
type ShufflingState struct {
	LatestActiveIndexRoots [EPOCHS_PER_HISTORICAL_VECTOR]Root
	CompactCommitteesRoots [EPOCHS_PER_HISTORICAL_VECTOR]Root
}

// Epoch is expected to be between (current_epoch - EPOCHS_PER_HISTORICAL_VECTOR + ACTIVATION_EXIT_DELAY, current_epoch + ACTIVATION_EXIT_DELAY].
func (state *ShufflingState) GetActiveIndexRoot(epoch Epoch) Root {
	return state.LatestActiveIndexRoots[epoch%EPOCHS_PER_HISTORICAL_VECTOR]
}

func (state *ShufflingState) GetCompactCommittesRoot(epoch Epoch) Root {
	return Root{} // TODO GetCompactCommittesRoot
}

func (state *BeaconState) UpdateActiveIndexRoot(epoch Epoch) {
	position := epoch % EPOCHS_PER_HISTORICAL_VECTOR
	indices := state.Validators.GetActiveValidatorIndices(epoch)
	state.LatestActiveIndexRoots[position] = ssz.HashTreeRoot(indices, ValidatorIndexListSSZ)
}

func (state *BeaconState) UpdateCompactCommitteesRoot(epoch Epoch) {
	position := epoch % EPOCHS_PER_HISTORICAL_VECTOR
	state.CompactCommitteesRoots[position] = state.GetCompactCommittesRoot(epoch)
}

// Generate a seed for the given epoch
func (state *BeaconState) GenerateSeed(epoch Epoch) Root {
	buf := make([]byte, 32*3)
	mix := state.GetRandomMix(epoch + EPOCHS_PER_HISTORICAL_VECTOR - MIN_SEED_LOOKAHEAD)
	copy(buf[0:32], mix[:])
	// get_active_index_root in spec, but only used once, and the assertion is unnecessary, since epoch input is always trusted
	activeIndexRoot := state.GetActiveIndexRoot(epoch)
	copy(buf[32:64], activeIndexRoot[:])
	binary.LittleEndian.PutUint64(buf[64:], uint64(epoch))
	return Hash(buf)
}

// Optimized compared to spec: takes pre-shuffled active indices as input, to not shuffle per-committee.
func computeCommittee(shuffled []ValidatorIndex, index uint64, count uint64) []ValidatorIndex {
	// Return the index'th shuffled committee out of the total committees data (shuffled active indices)
	startOffset := (uint64(len(shuffled)) * index) / count
	endOffset := (uint64(len(shuffled)) * (index + 1)) / count
	return shuffled[startOffset:endOffset]
}

func (state *BeaconState) GetCrosslinkCommittee(epoch Epoch, shard Shard) []ValidatorIndex {
	currentEpoch := state.Epoch()
	previousEpoch := state.PreviousEpoch()
	nextEpoch := currentEpoch + 1

	if !(previousEpoch <= epoch && epoch <= nextEpoch) {
		panic("could not retrieve crosslink committee for out of range slot")
	}

	seed := state.GenerateSeed(epoch)
	activeIndices := state.Validators.GetActiveValidatorIndices(epoch)
	// Active validators, shuffled in-place.
	// TODO: cache shuffling
	shuffling.UnshuffleList(activeIndices, seed)
	index := uint64((shard + SHARD_COUNT - state.GetEpochStartShard(epoch)) % SHARD_COUNT)
	count := state.Validators.GetEpochCommitteeCount(epoch)
	return computeCommittee(activeIndices, index, count)
}
