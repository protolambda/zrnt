package components

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type FinalityState struct {
	PreviousJustifiedEpoch Epoch
	CurrentJustifiedEpoch  Epoch
	PreviousJustifiedRoot  Root
	CurrentJustifiedRoot   Root
	JustificationBitfield  uint64
	FinalizedEpoch         Epoch
	FinalizedRoot          Root
}
