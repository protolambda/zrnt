package transition

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/beacon/epoch_processing"
)

type EpochProcessor func(state *BeaconState)

var epochProcessors = []EpochProcessor{
	ProcessEpochJustification,
	ProcessEpochCrosslinks,
	ProcessEpochRewardsAndPenalties,
	ProcessRegistryUpdates,
	ProcessEpochSlashings,
	ProcessEpochFinalUpdates,
}

func EpochTransition(state *BeaconState) {
	for _, p := range epochProcessors {
		p(state)
	}
}
