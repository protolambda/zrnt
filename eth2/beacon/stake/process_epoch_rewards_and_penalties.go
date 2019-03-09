package stake

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/crosslinks"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/justification"
)

var deltaCalculators = []DeltasCalculator{
	justification.DeltasJustification,
	crosslinks.DeltasCrosslinks,
	// TODO: split up the above where possible, and add others where necessary
}

func ProcessEpochRewardsAndPenalties(state *beacon.BeaconState) {
	sum := NewDeltas(uint64(len(state.Validator_registry)))
	for _, calc := range deltaCalculators {
		sum.Add(calc(state))
	}
	state.Validator_balances.ApplyStakeDeltas(sum)
}
