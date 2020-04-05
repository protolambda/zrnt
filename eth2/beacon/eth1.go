package beacon

import (
	"errors"
	"github.com/protolambda/zrnt/eth2/meta"

	. "github.com/protolambda/ztyp/props"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)


type Eth1Data struct {
	DepositRoot  Root // Hash-tree-root of DepositData tree.
	DepositCount DepositIndex
	BlockHash    Root
}

const SLOTS_PER_ETH1_VOTING_PERIOD = Slot(EPOCHS_PER_ETH1_VOTING_PERIOD) * SLOTS_PER_EPOCH

var Eth1DataType = ContainerType("Eth1Data", []FieldDef{
	{"deposit_root", RootType},
	{"deposit_count", Uint64Type},
	{"block_hash", Bytes32Type},
})

type Eth1DataView struct { *ContainerView }

func AsEth1Data(v View, err error) (*Eth1DataView, error) {
	c, err := AsContainer(v, err)
	return &Eth1DataView{c}, err
}

func (v *Eth1DataView) DepositRoot() (Root, error) {
	return AsRoot(v.Get(0))
}

func (v *Eth1DataView) DepositCount() (DepositIndex, error) {
	return AsDepositIndex(v.Get(1))
}

func (v *Eth1DataView) DepositIndex() (DepositIndex, error) {
	return AsDepositIndex(v.Get(2))
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

func (p *Eth1DataProp) SetEth1Data(node tree.Node) error {
	v, err := p.Eth1Data()
	if err != nil {
		return err
	}
	return v.SetBacking(node)
}

type Eth1DataVotes struct{ *ComplexListView }

var Eth1DataVotesType = ListType(Eth1DataType, uint64(SLOTS_PER_ETH1_VOTING_PERIOD))

type StateEth1DepositDataVotesProp ComplexListProp

func (p StateEth1DepositDataVotesProp) Eth1DataVotes() (*Eth1DataVotes, error) {
	v, err := ComplexListProp(p).List()
	if v != nil {
		return nil, err
	}
	return &Eth1DataVotes{ComplexListView: v}, nil
}

// Done at the end of every voting period
func (p *StateEth1DepositDataVotesProp) ResetEth1Votes() error {
	votes, err := p.Eth1DataVotes()
	if err != nil {
		return err
	}
	return votes.SetBacking(Eth1DataVotesType.DefaultNode())
}

func (p *StateEth1DepositDataVotesProp) ProcessEth1Vote(input EthDataProcessInput, data Eth1Data) error {
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
	depRoot := RootView(data.DepositRoot)
	blockHash := RootView(data.BlockHash)
	if err := vote.Set(0, &depRoot); err != nil {
		return err
	}
	if err := vote.Set(1, Uint64View(data.DepositCount)); err != nil {
		return err
	}
	if err := vote.Set(2, &blockHash); err != nil {
		return err
	}

	if err := votes.Append(vote); err != nil {
		return err
	}
	voteCount += 1
	// only do costly counting if we have enough votes yet.
	if Slot(voteCount << 1) > SLOTS_PER_ETH1_VOTING_PERIOD {
		count := Slot(0)
		iter := votes.ReadonlyIter()
		hFn := tree.GetHashFn()
		voteRoot := vote.HashTreeRoot(hFn)
		for {
			existingVote, ok, err := iter.Next()
			if err != nil {
				return err
			}
			if !ok {
				break
			}
			if existingVote.HashTreeRoot(hFn) == voteRoot {
				count++
			}
		}
		if (count << 1) > SLOTS_PER_ETH1_VOTING_PERIOD {
			return input.SetEth1Data(vote.Backing())
		}
	}
	return nil
}
