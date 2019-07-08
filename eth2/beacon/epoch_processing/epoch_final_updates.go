package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessEpochFinalUpdates(state *BeaconState) {
	currentEpoch := state.Epoch()
	nextEpoch := currentEpoch + 1

	// Reset eth1 data votes
	if (state.Slot+1)%SLOTS_PER_ETH1_VOTING_PERIOD == 0 {
		state.Eth1DataVotes = make([]Eth1Data, 0)
	}
	// Update effective balances with hysteresis
	for i, v := range state.Validators {
		balance := state.Balances[i]
		if balance < v.EffectiveBalance ||
			v.EffectiveBalance+3*HALF_INCREMENT < balance {
			v.EffectiveBalance = balance - (balance % EFFECTIVE_BALANCE_INCREMENT)
			if MAX_EFFECTIVE_BALANCE < v.EffectiveBalance {
				v.EffectiveBalance = MAX_EFFECTIVE_BALANCE
			}
		}
	}
	// Update start shard
	state.LatestStartShard = (state.LatestStartShard + state.GetShardDelta(currentEpoch)) % SHARD_COUNT

	// Set active index root
	indexRootPosition := (nextEpoch + ACTIVATION_EXIT_DELAY) % EPOCHS_PER_HISTORICAL_VECTOR
	state.LatestActiveIndexRoots[indexRootPosition] =
		ssz.HashTreeRoot(state.Validators.GetActiveValidatorIndices(nextEpoch+ACTIVATION_EXIT_DELAY), ValidatorIndexListSSZ)
	state.Slashings[nextEpoch%EPOCHS_PER_SLASHINGS_VECTOR] = 0 // TODO

	// Set randao mix
	state.LatestRandaoMixes[nextEpoch%EPOCHS_PER_HISTORICAL_VECTOR] = state.GetRandaoMix(currentEpoch)

	// Set historical root accumulator
	if nextEpoch%SLOTS_PER_HISTORICAL_ROOT.ToEpoch() == 0 {
		historicalBatch := HistoricalBatch{
			BlockRoots: state.LatestBlockRoots,
			StateRoots: state.LatestStateRoots,
		}

		state.HistoricalRoots = append(state.HistoricalRoots, ssz.HashTreeRoot(historicalBatch, HistoricalBatchSSZ))
	}
	// Rotate current/previous epoch attestations
	state.PreviousEpochAttestations = state.CurrentEpochAttestations
	state.CurrentEpochAttestations = nil
}
