package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/deltas_computation"
)

var deltaCalculators = []beacon.DeltasCalculator{
	deltas_computation.DeltasJustificationAndFinalizationDeltas,
	deltas_computation.DeltasCrosslinks,
	// TODO: split up the above where possible, and add others where necessary
}

func ProcessEpochRewardsAndPenalties(state *beacon.BeaconState) {
	sum := beacon.NewDeltas(uint64(len(state.ValidatorRegistry)))
	valuator := beacon.NewDefaultValuator(state)
	for _, calc := range deltaCalculators {
		sum.Add(calc(state, valuator))
	}
	state.ApplyDeltas(sum)
}
