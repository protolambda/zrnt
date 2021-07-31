package altair

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type SyncAggregatorSelectionData struct {
	Slot              common.Slot `yaml:"slot" json:"slot"`
	SubcommitteeIndex Uint64View  `yaml:"subcommittee_index" json:"subcommittee_index"`
}

func (agg *SyncAggregatorSelectionData) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(
		&agg.Slot,
		&agg.SubcommitteeIndex,
	)
}

func (agg *SyncAggregatorSelectionData) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(
		&agg.Slot,
		&agg.SubcommitteeIndex,
	)
}

func (agg *SyncAggregatorSelectionData) ByteLength() uint64 {
	return 8 + 8
}

func (agg *SyncAggregatorSelectionData) FixedLength() uint64 {
	return 8 + 8
}

func (agg *SyncAggregatorSelectionData) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		&agg.Slot,
		&agg.SubcommitteeIndex,
	)
}
