package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/util/hashing"
)

// Return the sequence of sync committee indices (which may include duplicate indices)
// for the next sync committee, given a state at a sync committee period boundary.
//
// Note: Committee can contain duplicate indices for small validator sets (< SYNC_COMMITTEE_SIZE + 128)
func ComputeSyncCommitteeIndices(spec *Spec, state BeaconState, baseEpoch Epoch, active []ValidatorIndex) ([]ValidatorIndex, error) {
	if len(active) == 0 {
		return nil, errors.New("no active validators to compute sync committee from")
	}
	slot, err := state.Slot()
	if err != nil {
		return nil, err
	}
	if epoch := spec.SlotToEpoch(slot); baseEpoch > epoch+1 {
		return nil, fmt.Errorf("stat at slot %d (epoch %d) is not far along enough to compute sync committee data for epoch %d", slot, epoch, baseEpoch)
	}
	syncCommitteeIndices := make([]ValidatorIndex, 0, spec.SYNC_COMMITTEE_SIZE)
	mixes, err := state.RandaoMixes()
	if err != nil {
		return nil, err
	}
	periodSeed, err := GetSeed(spec, mixes, baseEpoch, spec.DOMAIN_SYNC_COMMITTEE)
	if err != nil {
		return nil, err
	}
	vals, err := state.Validators()
	if err != nil {
		return nil, err
	}
	hFn := hashing.GetHashFn()
	var buf [32 + 8]byte
	copy(buf[0:32], periodSeed[:])
	var h [32]byte
	i := ValidatorIndex(0)
	for uint64(len(syncCommitteeIndices)) < spec.SYNC_COMMITTEE_SIZE {
		shuffledIndex := PermuteIndex(spec.SHUFFLE_ROUND_COUNT, i%ValidatorIndex(len(active)),
			uint64(len(active)), periodSeed)
		candidateIndex := active[shuffledIndex]
		validator, err := vals.Validator(candidateIndex)
		if err != nil {
			return nil, err
		}
		effectiveBalance, err := validator.EffectiveBalance()
		if err != nil {
			return nil, err
		}
		// every 32 rounds, create a new source for randomByte
		if i%32 == 0 {
			binary.LittleEndian.PutUint64(buf[32:32+8], uint64(i/32))
			h = hFn(buf[:])
		}
		randomByte := h[i%32]
		if effectiveBalance*0xff >= spec.MAX_EFFECTIVE_BALANCE*Gwei(randomByte) {
			syncCommitteeIndices = append(syncCommitteeIndices, candidateIndex)
		}
		i += 1
	}
	return syncCommitteeIndices, nil
}
