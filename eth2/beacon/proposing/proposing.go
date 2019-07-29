package proposing

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
)

type EpochProposerIndices [SLOTS_PER_EPOCH]ValidatorIndex

type ProposersData struct {
	Epoch   Epoch
	Current *EpochProposerIndices
	Next    *EpochProposerIndices
}

func (state *ProposersData) GetBeaconProposerIndex(slot Slot) ValidatorIndex {
	epoch := slot.ToEpoch()
	if epoch == state.Epoch {
		return state.Current[slot-state.Epoch.GetStartSlot()]
	} else if epoch == state.Epoch+1 {
		return state.Next[slot-(state.Epoch + 1).GetStartSlot()]
	} else {
		panic(fmt.Errorf("slot %d not within range", slot))
	}
}

type ProposingFeature struct {
	Meta interface {
		meta.Versioning
		meta.CrosslinkCommittees
		meta.EffectiveBalances
		meta.CommitteeCount
		meta.CrosslinkTiming
		meta.EpochSeed
	}
}

func (f *ProposingFeature) LoadBeaconProposersData() (out *ProposersData) {
	currentEpoch := f.Meta.CurrentEpoch()
	return &ProposersData{
		Epoch:   currentEpoch,
		Current: f.LoadBeaconProposerIndices(currentEpoch),
		Next:    f.LoadBeaconProposerIndices(currentEpoch + 1),
	}
}

// Return the beacon proposer index for the current slot
func (f *ProposingFeature) LoadBeaconProposerIndices(epoch Epoch) (out *EpochProposerIndices) {
	seed := f.Meta.GetSeed(epoch)
	buf := make([]byte, 32+8, 32+8)
	copy(buf[0:32], seed[:])
	balanceWeightedProposer := func(committee []ValidatorIndex) ValidatorIndex {
		for i := uint64(0); true; i++ {
			binary.LittleEndian.PutUint64(buf[32:], i)
			h := Hash(buf)
			for j := uint64(0); j < 32; j++ {
				randomByte := h[j]
				candidateIndex := committee[(uint64(epoch)+((i<<5)|j))%uint64(len(committee))]
				effectiveBalance := f.Meta.EffectiveBalance(candidateIndex)
				if effectiveBalance*0xff >= MAX_EFFECTIVE_BALANCE*Gwei(randomByte) {
					return candidateIndex
				}
			}
		}
		panic("random (but balance-biased) infinite scrolling through a committee should always find a proposer")
	}

	out = new(EpochProposerIndices)

	// A.k.a. committeesPerSlot
	shardsPerSlot := Shard(f.Meta.GetCommitteeCount(epoch) / uint64(SLOTS_PER_EPOCH))

	startShard := f.Meta.GetStartShard(epoch)
	offset := Shard(0)
	for i := Slot(0); i < SLOTS_PER_EPOCH; i++ {
		offset += shardsPerSlot
		shard := (startShard + offset) % SHARD_COUNT
		firstCommittee := f.Meta.GetCrosslinkCommittee(epoch, shard)
		out[i] = balanceWeightedProposer(firstCommittee)
	}
	return
}
