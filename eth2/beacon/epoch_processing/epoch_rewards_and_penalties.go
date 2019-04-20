package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/deltas_computation"
)

var deltaCalculators = []DeltasCalculator{
	deltas_computation.DeltasJustificationAndFinalizationDeltas,
	deltas_computation.DeltasCrosslinks,
	// TODO: split up the above where possible, and add others where necessary
}

func ProcessEpochRewardsAndPenalties(state *BeaconState) {
	sum := NewDeltas(uint64(len(state.ValidatorRegistry)))
	for _, calc := range deltaCalculators {
		sum.Add(calc(state))
	}
	state.ApplyDeltas(sum)
}
