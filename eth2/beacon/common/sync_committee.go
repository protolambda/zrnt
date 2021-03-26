package common

import (
	"encoding/binary"
	"fmt"
	"github.com/protolambda/zrnt/eth2/util/hashing"
)

type SyncCommittee struct {
	Spec *Spec
	// Period is the start of the current sync committee period
	Period uint64
	// CurrentSyncCommittee is a slice of SYNC_COMMITTEE_SIZE validator indices for the period
	// It may contain duplicates.
	CurrentSyncCommittee []ValidatorIndex
	// NextSyncCommittee is the sync commitee for the next period
	NextSyncCommittee []ValidatorIndex
}

func (epc *SyncCommittee) GetSyncCommittee(epoch Epoch) ([]ValidatorIndex, error) {
	period := uint64(epoch / epc.Spec.SHARD_COMMITTEE_PERIOD)
	if epc.Period == period {
		return epc.CurrentSyncCommittee, nil
	} else if epc.Period+1 == period {
		return epc.NextSyncCommittee, nil
	} else {
		return nil, fmt.Errorf("epoch %d is in period %d, but only periods %d and %d are available",
			epoch, period, epc.Period, epc.Period+1)
	}
}

// Computes a sync committee. The baseEpoch must align with the sync committee period,
// and active indices must match that of the time of the baseEpoch.
func ComputeSyncCommittee(spec *Spec, state BeaconState, baseEpoch Epoch, active []ValidatorIndex) (*SyncCommittee, error) {
	indices, err := ComputeSyncCommitteeIndices(spec, state, baseEpoch, active)
	if err != nil {
		return nil, err
	}
	// TODO
	return &SyncCommittee{
		Spec:                 spec,
		Period:               uint64(baseEpoch / spec.SHARD_COMMITTEE_PERIOD),
		CurrentSyncCommittee: indices,
		NextSyncCommittee:    nil, // TODO
	}, nil
}

func ComputeSyncCommitteeIndices(spec *Spec, state BeaconState, baseEpoch Epoch, active []ValidatorIndex) ([]ValidatorIndex, error) {
	// no validators? No sync committee to compute. Keep it nil so we can detect and compute it later.
	if len(active) == 0 {
		return nil, nil
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
