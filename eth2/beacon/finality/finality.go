package finality

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type FinalityState struct {
	JustificationBits           JustificationBits
	PreviousJustifiedCheckpoint Checkpoint
	CurrentJustifiedCheckpoint  Checkpoint
	FinalizedCheckpoint         Checkpoint
}

func (state *FinalityState) Finalized() Checkpoint {
	return state.FinalizedCheckpoint
}

func (state *FinalityState) PreviousJustified() Checkpoint {
	return state.PreviousJustifiedCheckpoint
}

func (state *FinalityState) CurrentJustified() Checkpoint {
	return state.CurrentJustifiedCheckpoint
}

type JustificationFeature struct {
	State *FinalityState
	Meta  interface {
		meta.Versioning
		meta.History
		meta.Staking
		meta.TargetStaking
	}
}

func (f *JustificationFeature) Justify(checkpoint Checkpoint) {
	currentEpoch := f.Meta.CurrentEpoch()
	if currentEpoch < checkpoint.Epoch {
		panic("cannot justify future epochs")
	}
	epochsAgo := currentEpoch - checkpoint.Epoch
	if epochsAgo >= Epoch(f.State.JustificationBits.BitLen()) {
		panic("cannot justify history past justification bitfield length")
	}

	f.State.CurrentJustifiedCheckpoint = checkpoint
	f.State.JustificationBits[0] |= 1 << epochsAgo
}

type JustificationBits [1]byte

func (jb *JustificationBits) BitLen() uint64 {
	return 4
}

// Prepare bitfield for next epoch by shifting previous bits (truncating to bitfield length)
func (jb *JustificationBits) NextEpoch() {
	// shift and mask
	jb[0] = (jb[0] << 1) & 0x0f
}

func (jb *JustificationBits) IsJustified(epochsAgo ...Epoch) bool {
	for _, t := range epochsAgo {
		if jb[0]&(1<<t) != 0 {
			return false
		}
	}
	return true
}
