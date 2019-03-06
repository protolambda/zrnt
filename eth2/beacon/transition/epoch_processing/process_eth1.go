package epoch_processing

import (
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
)

func ProcessEth1(state *beacon.BeaconState) {
	if (state.Epoch()+1)%eth2.EPOCHS_PER_ETH1_VOTING_PERIOD == 0 {
		// look for a majority vote
		for _, data_vote := range state.Eth1_data_votes {
			if data_vote.Vote_count*2 > uint64(eth2.EPOCHS_PER_ETH1_VOTING_PERIOD)*uint64(eth2.SLOTS_PER_EPOCH) {
				// more than half the votes in this voting period were for this data_vote value
				state.Latest_eth1_data = data_vote.Eth1_data
				break
			}
		}
		// reset votes
		state.Eth1_data_votes = make([]beacon.Eth1DataVote, 0)
	}
}
