package components

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type Eth1Data struct {
	DepositRoot  Root // Hash-tree-root of DepositData tree.
	DepositCount DepositIndex
	BlockHash    Root
}

type Eth1DataVotes []Eth1Data

func (_ *Eth1DataVotes) Limit() uint64 {
	return uint64(SLOTS_PER_ETH1_VOTING_PERIOD)
}

// Ethereum 1.0 chain data
type Eth1State struct {
	Eth1Data      Eth1Data
	Eth1DataVotes Eth1DataVotes
	DepositIndex  DepositIndex
}

// Done at the end of every voting period
func (state *Eth1State) ResetEth1Votes() {
	state.Eth1DataVotes = make([]Eth1Data, 0)
}

func (state *Eth1State) ProcessEth1Vote(data Eth1Data) error {
	state.Eth1DataVotes = append(state.Eth1DataVotes, data)
	count := Slot(0)
	for _, v := range state.Eth1DataVotes {
		if v == data {
			count++
		}
	}
	if count*2 > SLOTS_PER_ETH1_VOTING_PERIOD {
		state.Eth1Data = data
	}
	return nil
}
