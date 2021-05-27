package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	. "github.com/protolambda/ztyp/view"
)

var PendingShardHeaderType = ContainerType("PendingShardHeader", []FieldDef{
	// TODO
})

type PendingShardHeader struct {
}

type PendingShardHeaders []PendingShardHeader

func PendingShardHeadersType(spec *common.Spec) *ComplexListTypeDef {
	return ComplexListType(PendingShardHeaderType,
		spec.MAX_SHARDS*spec.MAX_SHARD_HEADERS_PER_SHARD*uint64(spec.SLOTS_PER_EPOCH))
}
