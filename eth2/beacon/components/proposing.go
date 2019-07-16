package components

import (
	"encoding/binary"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
)

// Return the beacon proposer index for the current slot
func (state *BeaconState) GetBeaconProposerIndex() ValidatorIndex {
	epoch := state.Epoch()
	committeesPerSlot := state.Validators.GetCommitteeCount(epoch) / uint64(SLOTS_PER_EPOCH)
	offset := Shard(committeesPerSlot) * Shard(state.Slot%SLOTS_PER_EPOCH)
	shard := (state.GetStartShard(epoch) + offset) % SHARD_COUNT
	firstCommittee := state.PrecomputedData.GetCrosslinkCommittee(epoch, shard)
	seed := state.GetSeed(epoch)
	buf := make([]byte, 32+8, 32+8)
	copy(buf[0:32], seed[:])
	for i := uint64(0); true; i++ {
		binary.LittleEndian.PutUint64(buf[32:], i)
		h := Hash(buf)
		for j := uint64(0); j < 32; j++ {
			randomByte := h[j]
			candidateIndex := firstCommittee[(uint64(epoch)+((i<<5)|j))%uint64(len(firstCommittee))]
			effectiveBalance := state.Validators[candidateIndex].EffectiveBalance
			if effectiveBalance*0xff >= MAX_EFFECTIVE_BALANCE*Gwei(randomByte) {
				return candidateIndex
			}
		}
	}
	return 0
}
