package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessEpochFinish(state *BeaconState) {
	currentEpoch := state.Epoch()
	nextEpoch := currentEpoch + 1

	// Reset eth1 data votes
	if state.Slot % SLOTS_PER_ETH1_VOTING_PERIOD == 0 {
		state.Eth1DataVotes = make([]Eth1Data, 0)
	}

	// Set active index root
	indexRootPosition := (nextEpoch + ACTIVATION_EXIT_DELAY) % LATEST_ACTIVE_INDEX_ROOTS_LENGTH
	state.LatestActiveIndexRoots[indexRootPosition] =
		ssz.HashTreeRoot(state.ValidatorRegistry.GetActiveValidatorIndices(nextEpoch + ACTIVATION_EXIT_DELAY))
	state.LatestSlashedBalances[nextEpoch%LATEST_SLASHED_EXIT_LENGTH] =
		state.LatestSlashedBalances[currentEpoch%LATEST_SLASHED_EXIT_LENGTH]

	// Set randao mix
	state.LatestRandaoMixes[nextEpoch%LATEST_RANDAO_MIXES_LENGTH] = state.GetRandaoMix(currentEpoch)
	// Set historical root accumulator
	if nextEpoch%SLOTS_PER_HISTORICAL_ROOT.ToEpoch() == 0 {
		historicalBatch := HistoricalBatch{
			BlockRoots: state.LatestBlockRoots,
			StateRoots: state.LatestStateRoots,
		}

		state.HistoricalRoots = append(state.HistoricalRoots, ssz.HashTreeRoot(historicalBatch))
	}
	// Rotate current/previous epoch attestations
	state.PreviousEpochAttestations = state.CurrentEpochAttestations
	state.CurrentEpochAttestations = make([]*PendingAttestation, 0)
}
