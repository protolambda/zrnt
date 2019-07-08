package epoch

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
)

type EpochProcess func(state *BeaconState)

var epochProcessors = []EpochProcess{
	ProcessEpochJustification,
	ProcessEpochCrosslinks,
	ProcessEpochRewardsAndPenalties,
	ProcessEpochRegistryUpdates,
	ProcessEpochSlashings,
	ProcessEpochFinalUpdates,
}

func Transition(state *BeaconState) {
	for _, p := range epochProcessors {
		p(state)
	}
}
