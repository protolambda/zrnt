package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	. "github.com/protolambda/zrnt/eth2/beacon/deltas_computation"
	. "github.com/protolambda/zrnt/eth2/core"
)

type DeltasCalculator func(state *BeaconState) *Deltas

var deltaCalculators = []DeltasCalculator{
	AttestationDeltas,
	CrosslinksDeltas,
}

func ProcessEpochRewardsAndPenalties(state *BeaconState) {
	if state.Slot.ToEpoch() == GENESIS_EPOCH {
		return
	}
	sum := NewDeltas(uint64(len(state.Validators)))
	for _, calc := range deltaCalculators {
		sum.Add(calc(state))
	}
	state.Balances.ApplyDeltas(sum)
}
