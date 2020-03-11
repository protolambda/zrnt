package eth1

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/core"
)

type Eth1VoteProcessor interface {
	ProcessEth1Vote(data Eth1Data) error
}

type Eth1Data struct {
	DepositRoot  Root // Hash-tree-root of DepositData tree.
	DepositCount DepositIndex
	BlockHash    Root
}

type Eth1DataVotes []Eth1Data

func (_ *Eth1DataVotes) Limit() uint64 {
	return uint64(EPOCHS_PER_ETH1_VOTING_PERIOD * SLOTS_PER_EPOCH)
}

// Ethereum 1.0 chain data
type Eth1State struct {
	Eth1Data      Eth1Data
	Eth1DataVotes Eth1DataVotes
	DepositIndex  DepositIndex
}

func (state *Eth1State) DepIndex() DepositIndex {
	return state.DepositIndex
}

func (state *Eth1State) DepCount() DepositIndex {
	return state.Eth1Data.DepositCount
}

func (state *Eth1State) DepRoot() Root {
	return state.Eth1Data.DepositRoot
}

func (state *Eth1State) IncrementDepositIndex() {
	state.DepositIndex += 1
}

const SLOTS_PER_ETH1_VOTING_PERIOD = Slot(EPOCHS_PER_ETH1_VOTING_PERIOD) * SLOTS_PER_EPOCH

// Done at the end of every voting period
func (state *Eth1State) ResetEth1Votes() {
	if Slot(cap(state.Eth1DataVotes)) <= SLOTS_PER_ETH1_VOTING_PERIOD {
		state.Eth1DataVotes = make([]Eth1Data, 0, SLOTS_PER_ETH1_VOTING_PERIOD)
	} else {
		state.Eth1DataVotes = state.Eth1DataVotes[:0]
	}
}

func (state *Eth1State) ProcessEth1Vote(data Eth1Data) error {
	if Slot(len(state.Eth1DataVotes)) >= SLOTS_PER_ETH1_VOTING_PERIOD {
		return errors.New("cannot process Eth1 vote, already voted maximum times")
	}
	state.Eth1DataVotes = append(state.Eth1DataVotes, data)
	// only do costly counting if we have enough votes yet.
	if (Slot(len(state.Eth1DataVotes)) << 1) > SLOTS_PER_ETH1_VOTING_PERIOD {
		count := Slot(0)
		for i := range state.Eth1DataVotes {
			if state.Eth1DataVotes[i] == data {
				count++
			}
		}
		if (count << 1) > SLOTS_PER_ETH1_VOTING_PERIOD {
			state.Eth1Data = data
		}
	}
	return nil
}
