package beacon

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Eth1Address [20]byte

func (p Eth1Address) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(p[:])), nil
}

func (p Eth1Address) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *Eth1Address) UnmarshalText(text []byte) error {
	if p == nil {
		return errors.New("cannot decode into nil Eth1Address")
	}
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	if len(text) != 40 {
		return fmt.Errorf("unexpected length string '%s'", string(text))
	}
	_, err := hex.Decode(p[:], text)
	return err
}

type Eth1Data struct {
	DepositRoot  Root // Hash-tree-root of DepositData tree.
	DepositCount DepositIndex
	BlockHash    Root
}

func (b *Eth1Data) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&b.DepositRoot, &b.DepositCount, &b.BlockHash)
}

func (a *Eth1Data) Serialize(w *codec.EncodingWriter) error {
	return w.Container(a.DepositRoot, a.DepositCount, a.BlockHash)
}

func (a *Eth1Data) ByteLength() uint64 {
	return Eth1DataType.TypeByteLength()
}

func (a *Eth1Data) FixedLength() uint64 {
	return Eth1DataType.TypeByteLength()
}

func (b *Eth1Data) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(b.DepositRoot, b.DepositCount, b.BlockHash)
}

func (dat *Eth1Data) View() *Eth1DataView {
	depRv := RootView(dat.DepositRoot)
	blockRv := RootView(dat.BlockHash)
	c, _ := Eth1DataType.FromFields(&depRv, Uint64View(dat.DepositCount), &blockRv)
	return &Eth1DataView{c}
}

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

func (a *Eth1DataVotes) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, Eth1Data{})
		return &(*a)[i]
	}, Eth1DataType.TypeByteLength(), uint64(spec.EPOCHS_PER_ETH1_VOTING_PERIOD)*uint64(spec.SLOTS_PER_EPOCH))
}

func (a Eth1DataVotes) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, Eth1DataType.TypeByteLength(), uint64(len(a)))
}

func (a Eth1DataVotes) ByteLength(spec *Spec) (out uint64) {
	return uint64(len(a)) * Eth1DataType.TypeByteLength()
}

func (a *Eth1DataVotes) FixedLength(spec *Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li Eth1DataVotes) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, uint64(spec.EPOCHS_PER_ETH1_VOTING_PERIOD)*uint64(spec.SLOTS_PER_EPOCH))
}

func (c *Phase0Config) Eth1DataVotes() ListTypeDef {
	return ListType(Eth1DataType, uint64(c.EPOCHS_PER_ETH1_VOTING_PERIOD)*uint64(c.SLOTS_PER_EPOCH))
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
