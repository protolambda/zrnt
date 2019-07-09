package components

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type Eth1Data struct {
	DepositRoot  Root // Hash-tree-root of DepositData tree.
	DepositCount DepositIndex
	BlockHash    Root
}

// Ethereum 1.0 chain data
type Eth1State struct {
	LatestEth1Data Eth1Data
	Eth1DataVotes  []Eth1Data
	DepositIndex   DepositIndex
}

type Eth1BlockData struct {
	Eth1Data Eth1Data
}

func (data *Eth1BlockData) Process(state *BeaconState) error {
	state.Eth1DataVotes = append(state.Eth1DataVotes, data.Eth1Data)
	count := Slot(0)
	for _, v := range state.Eth1DataVotes {
		if v == data.Eth1Data {
			count++
		}
	}
	if count*2 > SLOTS_PER_ETH1_VOTING_PERIOD {
		state.LatestEth1Data = data.Eth1Data
	}
	return nil
}
