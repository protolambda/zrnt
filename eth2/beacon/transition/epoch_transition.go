package transition

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/epoch_processing"
)

type EpochProcessor func(state *beacon.BeaconState)

var epochProcessors = []EpochProcessor{
	epoch_processing.ProcessEpochEth1,
	epoch_processing.ProcessEpochCrosslinks,
	epoch_processing.ProcessEpochCrosslinks,
	epoch_processing.ProcessEpochRewardsAndPenalties,
	epoch_processing.ProcessEpochValidatorRegistry,
	epoch_processing.ProcessEpochSlashings,
	epoch_processing.ProcessEpochExitQueue,
	epoch_processing.ProcessEpochFinish,
}

func EpochTransition(state *beacon.BeaconState) {
	for _, p := range epochProcessors {
		p(state)
	}
}
