package components

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type FinalityState struct {
	JustificationBits           JustificationBits
	PreviousJustifiedCheckpoint Checkpoint
	CurrentJustifiedCheckpoint  Checkpoint
	FinalizedCheckpoint         Checkpoint
}

type JustificationBits [(JUSTIFICATION_BITS_LENGTH + 7) / 8]byte

func (jb *JustificationBits) BitLen() uint32 {
	return JUSTIFICATION_BITS_LENGTH
}

type Checkpoint struct {
	Epoch Epoch
	Root  Root
}
