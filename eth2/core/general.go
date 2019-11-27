package core

import . "github.com/protolambda/ztyp/view"

type Root [32]byte
const Bytes32Type = RootType

type Bytes []byte

// 32 bits, not strictly an integer, hence represented as 4 bytes
// (bytes not necessarily corresponding to versions)
type Version [4]byte
var VersionType = VectorType(ByteType, 4)

func (v Version) ToUint32() uint32 {
	return uint32(v[0])<<24 | uint32(v[1])<<16 | uint32(v[2])<<8 | uint32(v[3])
}

type Shard uint64
const ShardType = Uint64Type

type CommitteeIndex uint64
const CommitteeIndexType = Uint64Type

type Gwei uint64
const GweiType = Uint64Type

type Checkpoint struct {
	Epoch Epoch
	Root  Root
}
var CheckpointType = &ContainerType{
	{"epoch", EpochType},
	{"root", RootType},
}

const SlotType = Uint64Type
