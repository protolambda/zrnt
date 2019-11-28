package core

import (
	. "github.com/protolambda/ztyp/props"
	. "github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

const Bytes32Type = RootType

type Bytes []byte

type Shard uint64
const ShardType = Uint64Type

type CommitteeIndex uint64
const CommitteeIndexType = Uint64Type

type Gwei uint64
const GweiType = Uint64Type

type GweiReadProp Uint64ReadProp

func (p GweiReadProp) Gwei() (Gwei, error) {
	v, err := (Uint64ReadProp)(p).Uint64()
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
