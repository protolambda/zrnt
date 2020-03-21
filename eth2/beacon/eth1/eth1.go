package eth1

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/ztyp/props"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Eth1VoteProcessor interface {
	ProcessEth1Vote(data Eth1Data) error
}

type Eth1Data struct {
	DepositRoot  Root // Hash-tree-root of DepositData tree.
	DepositCount DepositIndex
	BlockHash    Root
}

const SLOTS_PER_ETH1_VOTING_PERIOD = Slot(EPOCHS_PER_ETH1_VOTING_PERIOD) * SLOTS_PER_EPOCH

var Eth1DataType = &ContainerType{
	{"deposit_root", RootType},
	{"deposit_count", Uint64Type},
	{"block_hash", Bytes32Type},
}

type Eth1DataNode struct { *ContainerView }

func NewEth1DataNode() *Eth1DataNode {
	return &Eth1DataNode{ContainerView: Eth1DataType.New(nil)}
}

type Eth1DataProp ContainerReadProp

func (p Eth1DataProp) Eth1Data() (*Eth1DataNode, error) {
	if c, err := (ContainerReadProp)(p).Container(); err != nil {
		return nil, err
	} else {
		return &Eth1DataNode{ContainerView: c}, nil
	}
}

func (v *Eth1DataNode) DepositRoot() (Root, error) {
	return RootReadProp(PropReader(v, 0)).Root()
}

func (v *Eth1DataNode) DepositCount() (DepositIndex, error) {
	return DepositIndexReadProp(PropReader(v, 1)).DepositIndex()
}

func (v *Eth1DataNode) DepositIndex() (DepositIndex, error) {
	return DepositIndexReadProp(PropReader(v, 2)).DepositIndex()
}

type StateDepositIndexProps struct {
	DepositIndexReadProp
	DepositIndexWriteProp
}

func (p *StateDepositIndexProps) IncrementDepositIndex() error {
	d, err := p.DepositIndexReadProp.DepositIndex()
	if err != nil {
		return err
	}
	return p.DepositIndexWriteProp.SetDepositIndex(d + 1)
}

// Ethereum 1.0 chain data
type Eth1Props struct {
	Eth1Data      Eth1DataProp
	Eth1DataVotes Eth1DataVotes
	DepositIndex  StateDepositIndexProps
}

func (p *Eth1DataProp) DepIndex() (DepositIndex, error) {
	data, err := p.Eth1Data()
	if err != nil {
		return 0, err
	}
	return data.DepositIndex()
}

func (p *Eth1DataProp) DepCount() (DepositIndex, error) {
	data, err := p.Eth1Data()
	if err != nil {
		return 0, err
	}
	return data.DepositCount()
}

func (p *Eth1DataProp) DepRoot() (Root, error) {
	data, err := p.Eth1Data()
	if err != nil {
		return Root{}, err
	}
	return data.DepositRoot()
}

type Eth1DataVotes struct{ *ListView }

var Eth1DataVotesType = ListType(Eth1DataType, uint64(SLOTS_PER_ETH1_VOTING_PERIOD))

type StateEth1DepositDataVotesProp ListReadProp

func (p StateEth1DepositDataVotesProp) Eth1DataVotes() (*Eth1DataVotes, error) {
	v, err := ListReadProp(p).List()
	if v != nil {
		return nil, err
	}
	return &Eth1DataVotes{ListView: v}, nil
}

// Done at the end of every voting period
func (p *StateEth1DepositDataVotesProp) ResetEth1Votes() error {
	votes, err := p.Eth1DataVotes()
	if err != nil {
		return err
	}
	// TODO; viewhooks
	return votes.ViewHook.PropagateChange(Eth1DataVotesType.New(nil))
}

func (p *StateEth1DepositDataVotesProp) ProcessEth1Vote(data Eth1Data) error {
	votes, err := p.Eth1DataVotes()
	if err != nil {
		return err
	}
	voteCount, err := votes.Length()
	if err != nil {
		return err
	}
	if Slot(voteCount) >= SLOTS_PER_ETH1_VOTING_PERIOD {
		return errors.New("cannot process Eth1 vote, already voted maximum times")
	}
	vote := NewEth1DataNode()
	if err := vote.Set(0, &data.DepositRoot); err != nil {
		return err
	}
	if err := vote.Set(1, Uint64View(data.DepositCount)); err != nil {
		return err
	}
	if err := vote.Set(2, &data.BlockHash); err != nil {
		return err
	}

	if err := votes.Append(vote); err != nil {
		return err
	}
	voteCount += 1
	// only do costly counting if we have enough votes yet.
	if Slot(voteCount << 1) > SLOTS_PER_ETH1_VOTING_PERIOD {
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
