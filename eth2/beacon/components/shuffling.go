package components

import (
	"encoding/binary"
	"errors"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/shuffling"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

// Randomness and committees
type ShufflingState struct {
	LatestRandaoMixes      [EPOCHS_PER_HISTORICAL_VECTOR]Root
	LatestStartShard       Shard
	LatestActiveIndexRoots [EPOCHS_PER_HISTORICAL_VECTOR]Root
	CompactCommitteesRoots [EPOCHS_PER_HISTORICAL_VECTOR]Root
}

func (state *ShufflingState) GetRandaoMix(epoch Epoch) Root {
	// Epoch is expected to be between (current_epoch - LATEST_RANDAO_MIXES_LENGTH, current_epoch].
	// TODO: spec has expectations on input, but doesn't enforce them, and purposefully ignores them in some calls.
	return state.LatestRandaoMixes[epoch%EPOCHS_PER_HISTORICAL_VECTOR]
}

func (state *ShufflingState) GetActiveIndexRoot(epoch Epoch) Root {
	// Epoch is expected to be between (current_epoch - LATEST_ACTIVE_INDEX_ROOTS_LENGTH + ACTIVATION_EXIT_DELAY, current_epoch + ACTIVATION_EXIT_DELAY].
	// TODO: spec has expectations on input, but doesn't enforce them, and purposefully ignores them in some calls.
	return state.LatestActiveIndexRoots[epoch%EPOCHS_PER_HISTORICAL_VECTOR]
}

// Generate a seed for the given epoch
func (state *ShufflingState) GenerateSeed(epoch Epoch) Root {
	buf := make([]byte, 32*3)
	mix := state.GetRandaoMix(epoch + EPOCHS_PER_HISTORICAL_VECTOR - MIN_SEED_LOOKAHEAD)
	copy(buf[0:32], mix[:])
	// get_active_index_root in spec, but only used once, and the assertion is unnecessary, since epoch input is always trusted
	activeIndexRoot := state.GetActiveIndexRoot(epoch)
	copy(buf[32:64], activeIndexRoot[:])
	binary.LittleEndian.PutUint64(buf[64:], uint64(epoch))
	return Hash(buf)
}

func (state *BeaconState) GetEpochStartShard(epoch Epoch) Shard {
	currentEpoch := state.Epoch()
	checkEpoch := currentEpoch + 1
	if epoch > checkEpoch {
		panic("cannot find start shard for epoch, epoch is too new")
	}
	shard := (state.LatestStartShard + state.Validators.GetShardDelta(currentEpoch)) % SHARD_COUNT
	for checkEpoch > epoch {
		checkEpoch--
		shard = (shard + SHARD_COUNT - state.Validators.GetShardDelta(checkEpoch)) % SHARD_COUNT
	}
	return shard
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

type RandaoRevealBlockData struct {
	RandaoReveal BLSSignature
}

var RandaoEpochSSZ = zssz.GetSSZ((*Epoch)(nil))

func (revealData *RandaoRevealBlockData) Process(state *BeaconState) error {
	propIndex := state.GetBeaconProposerIndex()
	proposer := state.Validators[propIndex]
	currentEpoch := state.Epoch()
	if !bls.BlsVerify(
		proposer.Pubkey,
		ssz.HashTreeRoot(state.Epoch(), RandaoEpochSSZ),
		revealData.RandaoReveal,
		state.GetDomain(DOMAIN_RANDAO, currentEpoch),
	) {
		return errors.New("randao invalid")
	}
	state.LatestRandaoMixes[state.Epoch()%EPOCHS_PER_HISTORICAL_VECTOR] = XorBytes32(
		state.GetRandaoMix(currentEpoch),
		Hash(revealData.RandaoReveal[:]))
	return nil
}
