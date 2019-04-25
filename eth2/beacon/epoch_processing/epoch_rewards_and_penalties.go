package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/deltas_computation"
)

var deltaCalculators = []DeltasCalculator{
	deltas_computation.DeltasJustificationAndFinalizationDeltas,
	deltas_computation.DeltasCrosslinks,
}

func ProcessEpochRewardsAndPenalties(state *BeaconState) {
	sum := NewDeltas(uint64(len(state.ValidatorRegistry)))
	for _, calc := range deltaCalculators {
		sum.Add(calc(state))
	}
	state.ApplyDeltas(sum)
}
