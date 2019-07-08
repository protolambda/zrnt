package epoch

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	. "github.com/protolambda/zrnt/eth2/core"
)

func ProcessEpochRewardsAndPenalties(state *BeaconState) {
	if state.Slot.ToEpoch() == GENESIS_EPOCH {
		return
	}
	sum := NewDeltas(uint64(len(state.Validators)))
	sum.Add(state.AttestationDeltas())
	sum.Add(state.CrosslinksDeltas())
	state.Balances.ApplyDeltas(sum)
}
