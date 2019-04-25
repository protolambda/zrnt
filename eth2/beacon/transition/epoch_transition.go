package transition

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/epoch_processing"
)

type EpochProcessor func(state *BeaconState)

var epochProcessors = []EpochProcessor{
	epoch_processing.ProcessEpochJustification,
	epoch_processing.ProcessEpochCrosslinks,
	epoch_processing.ProcessEpochRewardsAndPenalties,
	epoch_processing.ProcessRegistryUpdates,
	epoch_processing.ProcessEpochSlashings,
	epoch_processing.ProcessEpochFinalUpdates,
}

func EpochTransition(state *BeaconState) {
	for _, p := range epochProcessors {
		p(state)
	}
}
