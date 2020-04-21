package beacon

import (
	"errors"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Eth1Data struct {
	DepositRoot  Root // Hash-tree-root of DepositData tree.
	DepositCount DepositIndex
	BlockHash    Root
}

func (dat *Eth1Data) View() *Eth1DataView {
	depRv := RootView(dat.DepositRoot)
	blockRv := RootView(dat.BlockHash)
	c, _ := Eth1DataType.FromFields(&depRv, Uint64View(dat.DepositCount), &blockRv)
	return &Eth1DataView{c}
}

const SLOTS_PER_ETH1_VOTING_PERIOD = Slot(EPOCHS_PER_ETH1_VOTING_PERIOD) * SLOTS_PER_EPOCH

var Eth1DataType = ContainerType("Eth1Data", []FieldDef{
	{"deposit_root", RootType},
	{"deposit_count", Uint64Type},
	{"block_hash", Bytes32Type},
})

type Eth1DataView struct{ *ContainerView }

func AsEth1Data(v View, err error) (*Eth1DataView, error) {
	c, err := AsContainer(v, err)
	return &Eth1DataView{c}, err
}

func (v *Eth1DataView) DepositRoot() (Root, error) {
	return AsRoot(v.Get(0))
}

func (v *Eth1DataView) SetDepositRoot(r Root) error {
	rv := RootView(r)
	return v.Set(0, &rv)
}

func (v *Eth1DataView) DepositCount() (DepositIndex, error) {
	return AsDepositIndex(v.Get(1))
}

func (v *Eth1DataView) DepositIndex() (DepositIndex, error) {
	return AsDepositIndex(v.Get(2))
}

type Eth1DataVotes []Eth1Data

func (_ *Eth1DataVotes) Limit() uint64 {
	return uint64(SLOTS_PER_ETH1_VOTING_PERIOD)
}

var Eth1DataVotesType = ListType(Eth1DataType, uint64(SLOTS_PER_ETH1_VOTING_PERIOD))

type Eth1DataVotesView struct{ *ComplexListView }

func AsEth1DataVotes(v View, err error) (*Eth1DataVotesView, error) {
	c, err := AsComplexList(v, err)
	return &Eth1DataVotesView{c}, err
}

// Done at the end of every voting period
func (state *BeaconStateView) ResetEth1Votes() error {
	votes, err := state.Eth1DataVotes()
	if err != nil {
		return err
	}
	return votes.SetBacking(Eth1DataVotesType.DefaultNode())
}

func (state *BeaconStateView) ProcessEth1Vote(epc *EpochsContext, data Eth1Data) error {
	votes, err := state.Eth1DataVotes()
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
	vote := data.View()
	if err := votes.Append(vote); err != nil {
		return err
	}
	voteCount += 1
	// only do costly counting if we have enough votes yet.
	if Slot(voteCount<<1) > SLOTS_PER_ETH1_VOTING_PERIOD {
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
			return state.SetEth1Data(vote)
		}
	}
	return nil
}
