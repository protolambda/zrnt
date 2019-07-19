package shuffling

import (
	"encoding/binary"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
)

// Randomness and committees
type ShufflingState struct {
	LatestActiveIndexRoots [EPOCHS_PER_HISTORICAL_VECTOR]Root
}

// Epoch is expected to be between (current_epoch - EPOCHS_PER_HISTORICAL_VECTOR + ACTIVATION_EXIT_DELAY, current_epoch + ACTIVATION_EXIT_DELAY].
func (state *ShufflingState) GetActiveIndexRoot(epoch Epoch) Root {
	return state.LatestActiveIndexRoots[epoch%EPOCHS_PER_HISTORICAL_VECTOR]
}

func (state *ShufflingState) UpdateActiveIndexRoot(meta ActiveIndexRootMeta, epoch Epoch) {
	position := epoch % EPOCHS_PER_HISTORICAL_VECTOR
	state.LatestActiveIndexRoots[position] = meta.GetActiveIndexRoot(epoch)
}

// Generate a seed for the given epoch
func (state *ShufflingState) GetSeed(meta RandomnessMeta, epoch Epoch) Root {
	buf := make([]byte, 32*3)
	mix := meta.GetRandomMix(epoch + EPOCHS_PER_HISTORICAL_VECTOR - MIN_SEED_LOOKAHEAD) // Avoid underflow
	copy(buf[0:32], mix[:])
	activeIndexRoot := state.GetActiveIndexRoot(epoch)
	copy(buf[32:64], activeIndexRoot[:])
	binary.LittleEndian.PutUint64(buf[64:], uint64(epoch))
	return Hash(buf)
}

type ProposingReq interface {
	VersioningMeta
	CrosslinkCommitteeMeta
	CompactValidatorMeta
	RandomnessMeta
}

// Return the beacon proposer index for the current slot
func (state *ShufflingState) GetBeaconProposerIndex(meta ProposingReq) ValidatorIndex {
	epoch := meta.Epoch()
	committeesPerSlot := meta.GetCommitteeCount(epoch) / uint64(SLOTS_PER_EPOCH)
	offset := Shard(committeesPerSlot) * Shard(meta.Slot()%SLOTS_PER_EPOCH)
	shard := (meta.GetStartShard(epoch) + offset) % SHARD_COUNT
	firstCommittee := meta.GetCrosslinkCommittee(epoch, shard)
	seed := state.GetSeed(meta, epoch)
	buf := make([]byte, 32+8, 32+8)
	copy(buf[0:32], seed[:])
	for i := uint64(0); true; i++ {
		binary.LittleEndian.PutUint64(buf[32:], i)
		h := Hash(buf)
		for j := uint64(0); j < 32; j++ {
			randomByte := h[j]
			candidateIndex := firstCommittee[(uint64(epoch)+((i<<5)|j))%uint64(len(firstCommittee))]
			effectiveBalance := meta.EffectiveBalance(candidateIndex)
			if effectiveBalance*0xff >= MAX_EFFECTIVE_BALANCE*Gwei(randomByte) {
				return candidateIndex
			}
		}
	}
	return 0
}
