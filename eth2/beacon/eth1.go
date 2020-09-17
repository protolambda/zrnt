package beacon

import (
	"context"
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

func (c *Phase0Config) Eth1Data() *ContainerTypeDef {
	return ContainerType("Eth1Data", []FieldDef{
		{"deposit_root", RootType},
		{"deposit_count", Uint64Type},
		{"block_hash", Bytes32Type},
	})
}

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


func (c *Phase0Config) Eth1DataVotes() ListTypeDef {
	return ListType(c.Eth1Data(), uint64(c.EPOCHS_PER_ETH1_VOTING_PERIOD) * uint64(c.SLOTS_PER_EPOCH))
}

type Eth1DataVotesView struct{ *ComplexListView }

func AsEth1DataVotes(v View, err error) (*Eth1DataVotesView, error) {
	c, err := AsComplexList(v, err)
	return &Eth1DataVotesView{c}, err
}

// Done at the end of every voting period
func (c *Phase0Config) ResetEth1Votes(state *BeaconStateView) error {
	votes, err := state.Eth1DataVotes()
	if err != nil {
		return err
	}
	return votes.SetBacking(c.Eth1DataVotes().DefaultNode())
}

func (spec *Spec) ProcessEth1Vote(ctx context.Context, epc *EpochsContext, state *BeaconStateView, data Eth1Data) error {
	select {
	case <-ctx.Done():
		return TransitionCancelErr
	default: // Don't block.
		break
	}
	votes, err := state.Eth1DataVotes()
	if err != nil {
		return err
	}
	voteCount, err := votes.Length()
	if err != nil {
		return err
	}
	period := uint64(spec.EPOCHS_PER_ETH1_VOTING_PERIOD) * uint64(spec.SLOTS_PER_EPOCH)
	if voteCount >= period {
		return errors.New("cannot process Eth1 vote, already voted maximum times")
	}
	vote := data.View()
	if err := votes.Append(vote); err != nil {
		return err
	}
	voteCount += 1
	// only do costly counting if we have enough votes yet.
	if voteCount<<1 > period {
		count := uint64(0)
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
		if (count << 1) > period {
			return state.SetEth1Data(vote)
		}
	}
	return nil
}
