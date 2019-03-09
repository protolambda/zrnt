package epoch_processing

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
)

// "maybe_reset_eth1_period"
func ProcessEpochEth1(state *beacon.BeaconState) {
	if (state.Epoch()+1)%beacon.EPOCHS_PER_ETH1_VOTING_PERIOD == 0 {
		// look for a majority vote
		for _, dataVote := range state.Eth1DataVotes {
			if dataVote.VoteCount*2 > uint64(beacon.EPOCHS_PER_ETH1_VOTING_PERIOD)*uint64(beacon.SLOTS_PER_EPOCH) {
				// more than half the votes in this voting period were for this data_vote value
				state.LatestEth1Data = dataVote.Eth1Data
				break
			}
		}
		// reset votes
		state.Eth1DataVotes = make([]beacon.Eth1DataVote, 0)
	}
}
