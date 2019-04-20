package block_processing

import . "github.com/protolambda/zrnt/eth2/beacon"

func ProcessBlockEth1(state *BeaconState, block *BeaconBlock) error {
	state.Eth1DataVotes = append(state.Eth1DataVotes, block.Body.Eth1Data)
	count := Slot(0)
	for _, v := range state.Eth1DataVotes {
		if v == block.Body.Eth1Data {
			count++
		}
	}
	if count * 2 > SLOTS_PER_ETH1_VOTING_PERIOD {
		state.LatestEth1Data = block.Body.Eth1Data
	}
	return nil
}
