package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type AttestationData struct {
	Slot  common.Slot           `json:"slot" yaml:"slot"`
	Index common.CommitteeIndex `json:"index" yaml:"index"`

	// LMD GHOST vote
	BeaconBlockRoot common.Root `json:"beacon_block_root" yaml:"beacon_block_root"`

	// FFG vote
	Source common.Checkpoint `json:"source" yaml:"source"`
	Target common.Checkpoint `json:"target" yaml:"target"`

	// Shard header root
	ShardHeaderRoot common.Root `json:"shard_header_root" yaml:"shard_header_root"`
}

func (a *AttestationData) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&a.Slot, &a.Index, &a.BeaconBlockRoot, &a.Source, &a.Target, &a.ShardHeaderRoot)
}

func (a *AttestationData) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(a.Slot, a.Index, &a.BeaconBlockRoot, &a.Source, &a.Target, &a.ShardHeaderRoot)
}

func (a *AttestationData) ByteLength() uint64 {
	return AttestationDataType.TypeByteLength()
}

func (*AttestationData) FixedLength() uint64 {
	return AttestationDataType.TypeByteLength()
}

func (p *AttestationData) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(p.Slot, p.Index, p.BeaconBlockRoot, &p.Source, &p.Target, &p.ShardHeaderRoot)
}

func (data *AttestationData) View() *AttestationDataView {
	brv := RootView(data.BeaconBlockRoot)
	srv := RootView(data.ShardHeaderRoot)
	c, _ := AttestationDataType.FromFields(
		Uint64View(data.Slot),
		Uint64View(data.Index),
		&brv,
		data.Source.View(),
		data.Target.View(),
		&srv,
	)
	return &AttestationDataView{c}
}

var AttestationDataType = ContainerType("AttestationData", []FieldDef{
	{"slot", common.SlotType},
	{"index", common.CommitteeIndexType},
	// LMD GHOST vote
	{"beacon_block_root", RootType},
	// FFG vote
	{"source", common.CheckpointType},
	{"target", common.CheckpointType},
	// Shard header root
	{"shard_header_root", RootType},
})

type AttestationDataView struct{ *ContainerView }

func (v *AttestationDataView) Raw() (*AttestationData, error) {
	fields, err := v.FieldValues()
	slot, err := common.AsSlot(fields[0], err)
	comm, err := common.AsCommitteeIndex(fields[1], err)
	root, err := AsRoot(fields[2], err)
	source, err := common.AsCheckPoint(fields[3], err)
	target, err := common.AsCheckPoint(fields[4], err)
	shardHeaderRoot, err := AsRoot(fields[5], err)
	if err != nil {
		return nil, err
	}
	rawSource, err := source.Raw()
	if err != nil {
		return nil, err
	}
	rawTarget, err := target.Raw()
	if err != nil {
		return nil, err
	}
	return &AttestationData{
		Slot:            slot,
		Index:           comm,
		BeaconBlockRoot: root,
		Source:          rawSource,
		Target:          rawTarget,
		ShardHeaderRoot: shardHeaderRoot,
	}, nil
}

func AsAttestationData(v View, err error) (*AttestationDataView, error) {
	c, err := AsContainer(v, err)
	return &AttestationDataView{c}, err
}
