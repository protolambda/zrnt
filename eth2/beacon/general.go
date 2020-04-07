package beacon

import (
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Root = tree.Root
const Bytes32Type = RootType

type Bytes []byte

type Shard Uint64View
const ShardType = Uint64Type

type CommitteeIndex Uint64View
const CommitteeIndexType = Uint64Type

func AsCommitteeIndex(v View, err error) (CommitteeIndex, error) {
	i, err := AsUint64(v, err)
	return CommitteeIndex(i), err
}

type Gwei Uint64View
const GweiType = Uint64Type

func AsGwei(v View, err error) (Gwei, error) {
	i, err := AsUint64(v, err)
	return Gwei(i), err
}

type Checkpoint struct {
	Epoch Epoch
	Root  Root
}

func (c *Checkpoint) View() *CheckpointView {
	r := RootView(c.Root)
	res, _ := CheckpointType.FromFields(Uint64View(c.Epoch), &r)
	return &CheckpointView{res}
}

var CheckpointType = ContainerType("Checkpoint", []FieldDef{
	{"epoch", EpochType},
	{"root", RootType},
})

type CheckpointView struct {
	*ContainerView
}

func (v *CheckpointView) Set(ch *Checkpoint) error {
	return v.SetBacking(ch.View().Backing())
}

func (v *CheckpointView) Epoch() (Epoch, error) {
	return AsEpoch(v.Get(0))
}

func (v *CheckpointView) Root() (Root, error) {
	return AsRoot(v.Get(0))
}

func (v *CheckpointView) Raw() (Checkpoint, error) {
	epoch, err := AsEpoch(v.Get(0))
	if err != nil {
		return Checkpoint{}, err
	}
	root, err := AsRoot(v.Get(1))
	if err != nil {
		return Checkpoint{}, err
	}
	return Checkpoint{Epoch: epoch, Root: root}, nil
}

func AsCheckPoint(v View, err error) (*CheckpointView, error) {
	c, err := AsContainer(v, err)
	return &CheckpointView{c}, err
}