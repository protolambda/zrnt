package transition

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/crosslinks"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/eth1"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/exits/exit_queue"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/finish"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/justification"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/slashing"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/validator_registry"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/stake"
)

type EpochProcessor func(state *beacon.BeaconState)

var epochProcessors = []EpochProcessor{
	eth1.ProcessEpochEth1,
	justification.ProcessEpochCrosslinks,
	crosslinks.ProcessEpochCrosslinks,
	stake.ProcessEpochRewardsAndPenalties,
	validator_registry.ProcessEpochValidatorRegistry,
	slashing.ProcessEpochSlashings,
	exit_queue.ProcessEpochExitQueue,
	finish.ProcessEpochFinish,
}

func EpochTransition(state *beacon.BeaconState) {
	for _, p := range epochProcessors {
		p(state)
	}
}
