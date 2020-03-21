package core

import (
	. "github.com/protolambda/ztyp/props"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Root = tree.Root
const Bytes32Type = RootType

type Bytes []byte

type Shard uint64
const ShardType = Uint64Type

type CommitteeIndex uint64
const CommitteeIndexType = Uint64Type

type CommitteeIndexReadProp Uint64ReadProp

func (p CommitteeIndexReadProp) CommitteeIndex() (CommitteeIndex, error) {
	v, err := Uint64ReadProp(p).Uint64()
	return CommitteeIndex(v), err
}

type Gwei uint64
const GweiType = Uint64Type

type GweiReadProp Uint64ReadProp

func (p GweiReadProp) Gwei() (Gwei, error) {
	v, err := Uint64ReadProp(p).Uint64()
	return Gwei(v), err
}

type Checkpoint struct {
	Epoch Epoch
	Root  Root
}
var CheckpointType = &ContainerType{
	{"epoch", EpochType},
	{"root", RootType},
}

type CheckpointProp ContainerReadProp

func (p CheckpointProp) CheckPoint() (Checkpoint, error) {
	v, err := ContainerReadProp(p).Container()
	if err != nil {
		return Checkpoint{}, err
	}
	epoch, err := EpochReadProp(PropReader(v, 0)).Epoch()
	if err != nil {
		return Checkpoint{}, err
	}
	root, err := RootReadProp(PropReader(v, 1)).Root()
	if err != nil {
		return Checkpoint{}, err
	}
	return Checkpoint{Epoch: epoch, Root: root}, nil
}

func (p CheckpointProp) SetCheckPoint(ch Checkpoint) error {
	v, err := ContainerReadProp(p).Container()
	if err != nil {
		return err
	}
	if err := EpochWriteProp(PropWriter(v, 0)).SetEpoch(ch.Epoch); err != nil {
		return err
	}
	if err := RootWriteProp(PropWriter(v, 1)).SetRoot(ch.Root); err != nil {
		return err
	}
	return nil
}

type EpochStakeSummary struct {
	SourceStake Gwei
	TargetStake Gwei
	HeadStake Gwei
}
