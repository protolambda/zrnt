package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessEpochFinish(state *beacon.BeaconState) {
	currentEpoch := state.Epoch()
	nextEpoch := currentEpoch + 1
	// Set active index root
	indexRootPosition := (nextEpoch + beacon.ACTIVATION_EXIT_DELAY) % beacon.LATEST_ACTIVE_INDEX_ROOTS_LENGTH
	state.LatestActiveIndexRoots[indexRootPosition] =
		ssz.HashTreeRoot(state.ValidatorRegistry.GetActiveValidatorIndices(nextEpoch + beacon.ACTIVATION_EXIT_DELAY))
	state.LatestSlashedBalances[nextEpoch%beacon.LATEST_SLASHED_EXIT_LENGTH] =
		state.LatestSlashedBalances[currentEpoch%beacon.LATEST_SLASHED_EXIT_LENGTH]

	// Set randao mix
	state.LatestRandaoMixes[nextEpoch%beacon.LATEST_RANDAO_MIXES_LENGTH] = state.GetRandaoMix(currentEpoch)
	// Set historical root accumulator
	if nextEpoch%beacon.SLOTS_PER_HISTORICAL_ROOT.ToEpoch() == 0 {
		historicalBatch := beacon.HistoricalBatch{
			BlockRoots: state.LatestBlockRoots,
			StateRoots: state.LatestStateRoots,
		}

		state.HistoricalRoots = append(state.HistoricalRoots, ssz.HashTreeRoot(historicalBatch))
	}
	// Rotate current/previous epoch attestations
	state.PreviousEpochAttestations = state.CurrentEpochAttestations
	state.CurrentEpochAttestations = make([]beacon.PendingAttestation, 0)
}
