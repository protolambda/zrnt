package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/processes/crosslinks"
	"github.com/protolambda/zrnt/eth2/beacon/processes/justification"
)

var deltaCalculators = []beacon.DeltasCalculator{
	justification.DeltasJustification,
	crosslinks.DeltasCrosslinks,
	// TODO: split up the above where possible, and add others where necessary
}

func ProcessEpochRewardsAndPenalties(state *beacon.BeaconState) {
	sum := beacon.NewDeltas(uint64(len(state.ValidatorRegistry)))
	for _, calc := range deltaCalculators {
		sum.Add(calc(state))
	}
	state.ValidatorBalances.ApplyStakeDeltas(sum)
}
