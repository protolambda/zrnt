package common

import (
	"encoding/binary"
	"errors"
	"github.com/protolambda/zrnt/eth2/util/hashing"
)

func ComputeSyncCommitteeIndices(spec *Spec, state BeaconState, baseEpoch Epoch, active []ValidatorIndex) ([]ValidatorIndex, error) {
	if len(active) == 0 {
		return nil, errors.New("no active validators to compute sync committee from")
	}
	syncCommittee := make([]ValidatorIndex, 0, spec.SYNC_SUBCOMMITTEE_SIZE)
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
	for uint64(len(syncCommittee)) < spec.SYNC_SUBCOMMITTEE_SIZE {
		shuffledI := PermuteIndex(spec.SHUFFLE_ROUND_COUNT, i%ValidatorIndex(len(active)),
			uint64(len(active)), periodSeed)
		candidateIndex := active[shuffledI]
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
			syncCommittee = append(syncCommittee, candidateIndex)
		}
		i += 1
	}
	return syncCommittee, nil
}
