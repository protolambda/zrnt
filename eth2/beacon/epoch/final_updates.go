package epoch

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
)

func ProcessEpochFinalUpdates(state *BeaconState) {
	nextEpoch := state.Epoch() + 1

	// Reset eth1 data votes if it is the end of the voting period.
	if (state.Slot+1)%SLOTS_PER_ETH1_VOTING_PERIOD == 0 {
		state.ResetEth1Votes()
	}

	state.UpdateEffectiveBalances()
	state.RotateStartShard()
	state.UpdateActiveIndexRoot(nextEpoch + ACTIVATION_EXIT_DELAY)
	state.UpdateCompactCommitteesRoot(nextEpoch)
	state.ResetSlashings(nextEpoch)
	state.PrepareRandao(nextEpoch)

	// Set historical root accumulator
	if nextEpoch%SLOTS_PER_HISTORICAL_ROOT.ToEpoch() == 0 {
		state.UpdateHistoricalRoots()
	}

	state.RotateEpochAttestations()
}
