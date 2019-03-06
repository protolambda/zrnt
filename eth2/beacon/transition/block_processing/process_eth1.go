package block_processing

import "github.com/protolambda/go-beacon-transition/eth2/beacon"

func ProcessEth1(state *beacon.BeaconState, block *beacon.BeaconBlock) error 	{
	// If there exists an eth1_data_vote in state.Eth1_data_votes for which eth1_data_vote.eth1_data == block.Eth1_data (there will be at most one), set eth1_data_vote.vote_count += 1.
	// Otherwise, append to state.Eth1_data_votes a new Eth1DataVote(eth1_data=block.Eth1_data, vote_count=1).
	found := false
	for i, vote := range state.Eth1_data_votes {
		if vote.Eth1_data == block.Body.Eth1_data {
			state.Eth1_data_votes[i].Vote_count += 1
			found = true
			break
		}
	}
	if !found {
		state.Eth1_data_votes = append(state.Eth1_data_votes, beacon.Eth1DataVote{Eth1_data: block.Body.Eth1_data, Vote_count: 1})
	}
	return nil
}