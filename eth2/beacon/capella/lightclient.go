package capella

import (
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"github.com/protolambda/ztyp/view"
)

// https://github.com/ethereum/consensus-specs/blob/dev/specs/capella/light-client/sync-protocol.md
const ExecutionBranchLength = 4

var ExecutionBranchType = view.VectorType(common.Bytes32Type, ExecutionBranchLength)

type ExecutionBranch [ExecutionBranchLength]common.Bytes32

func (eb *ExecutionBranch) Deserialize(dr *codec.DecodingReader) error {
	roots := eb[:]
	return tree.ReadRoots(dr, &roots, 4)
}

func (eb *ExecutionBranch) FixedLength() uint64 {
	return ExecutionBranchType.TypeByteLength()
}

func (eb *ExecutionBranch) Serialize(w *codec.EncodingWriter) error {
	return tree.WriteRoots(w, eb[:])
}

func (eb *ExecutionBranch) ByteLength() (out uint64) {
	return ExecutionBranchType.TypeByteLength()
}

func (eb *ExecutionBranch) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.ComplexVectorHTR(func(i uint64) tree.HTR {
		if i < ExecutionBranchLength {
			return &eb[i]
		}
		return nil
	}, ExecutionBranchLength)
}

type LightClientHeader struct {
	Beacon          common.BeaconBlockHeader `yaml:"beacon" json:"beacon"`
	Execution       ExecutionPayloadHeader `yaml:"execution" json:"execution"`
	ExecutionBranch ExecutionBranch `yaml:"execution_branch" json:"execution_branch"`
}

var LightClientHeaderType = view.ContainerType("LightClientHeader", []view.FieldDef{
	{Name: "beacon", Type: common.BeaconBlockHeaderType},
	{Name: "execution", Type: ExecutionPayloadHeaderType},
	{Name: "execution_branch", Type: ExecutionBranchType},
})

func (l *LightClientHeader) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&l.Beacon, &l.Execution, &l.ExecutionBranch)
}

func (l *LightClientHeader) FixedLength() uint64 {
	return LightClientHeaderType.TypeByteLength()
}

func (l *LightClientHeader) Serialize(w *codec.EncodingWriter) error {
	return w.Container(&l.Beacon, &l.Execution, &l.ExecutionBranch)
}

func (l *LightClientHeader) ByteLength() (out uint64) {
	return codec.ContainerLength(&l.Beacon, &l.Execution, &l.ExecutionBranch)
}

func (l *LightClientHeader) HashTreeRoot(h tree.HashFn) common.Root {
	return h.HashTreeRoot(&l.Beacon, &l.Execution, &l.ExecutionBranch)
}

type LightClientBootstrap struct {
	Header                     LightClientHeader `yaml:"header" json:"header"`
	CurrentSyncCommittee       common.SyncCommittee `yaml:"current_sync_committee" json:"current_sync_committee"`
	CurrentSyncCommitteeBranch altair.SyncCommitteeProofBranch `yaml:"current_sync_committee_branch" json:"current_sync_committee_branch"`
}

func NewLightClientBootstrapType(spec *common.Spec) *view.ContainerTypeDef {
	return view.ContainerType("LightClientHeader", []view.FieldDef{
		{Name: "header", Type: LightClientHeaderType},
		{Name: "next_sync_committee", Type: common.SyncCommitteeType(spec)},
		{Name: "next_sync_committee_branch", Type: altair.SyncCommitteeProofBranchType},
	})
}

func (lcb *LightClientBootstrap) FixedLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&lcb.Header, spec.Wrap(&lcb.CurrentSyncCommittee), &lcb.CurrentSyncCommitteeBranch)
}

func (lcb *LightClientBootstrap) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&lcb.Header, spec.Wrap(&lcb.CurrentSyncCommittee), &lcb.CurrentSyncCommitteeBranch)
}

func (lcb *LightClientBootstrap) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&lcb.Header, spec.Wrap(&lcb.CurrentSyncCommittee), &lcb.CurrentSyncCommitteeBranch)
}

func (lcb *LightClientBootstrap) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&lcb.Header, spec.Wrap(&lcb.CurrentSyncCommittee), &lcb.CurrentSyncCommitteeBranch)
}

func (lcb *LightClientBootstrap) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		&lcb.Header,
		spec.Wrap(&lcb.CurrentSyncCommittee),
		&lcb.CurrentSyncCommitteeBranch,
	)
}

func LightClientUpdateType(spec *common.Spec) *view.ContainerTypeDef {
	return view.ContainerType("SyncCommittee", []view.FieldDef{
		{Name: "attested_header", Type: common.BeaconBlockHeaderType},
		{Name: "next_sync_committee", Type: common.SyncCommitteeType(spec)},
		{Name: "next_sync_committee_branch", Type: altair.SyncCommitteeProofBranchType},
		{Name: "finalized_header", Type: LightClientHeaderType},
		{Name: "finality_branch", Type: altair.FinalizedRootProofBranchType},
		{Name: "sync_aggregate", Type: altair.SyncAggregateType(spec)},
		{Name: "signature_slot", Type: common.SlotType},
	})
}

type LightClientUpdate struct {
	// Update beacon block header
	AttestedHeader LightClientHeader `yaml:"attested_header" json:"attested_header"`
	// Next sync committee corresponding to the header
	NextSyncCommittee       common.SyncCommittee            `yaml:"next_sync_committee" json:"next_sync_committee"`
	NextSyncCommitteeBranch altair.SyncCommitteeProofBranch `yaml:"next_sync_committee_branch" json:"next_sync_committee_branch"`
	// Finality proof for the update header
	FinalizedHeader LightClientHeader        `yaml:"finalized_header" json:"finalized_header"`
	FinalityBranch  altair.FinalizedRootProofBranch `yaml:"finality_branch" json:"finality_branch"`
	// Sync committee aggregate signature
	SyncAggregate altair.SyncAggregate `yaml:"sync_aggregate" json:"sync_aggregate"`
	// Slot at which the aggregate signature was created (untrusted)
	SignatureSlot common.Slot `yaml:"signature_slot" json:"signature_slot"`
}

func (lcu *LightClientUpdate) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(
		&lcu.AttestedHeader,
		spec.Wrap(&lcu.NextSyncCommittee),
		&lcu.NextSyncCommitteeBranch,
		&lcu.FinalizedHeader,
		&lcu.FinalityBranch,
		spec.Wrap(&lcu.SyncAggregate),
		&lcu.SignatureSlot,
	)
}

func (lcu *LightClientUpdate) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(
		&lcu.AttestedHeader,
		spec.Wrap(&lcu.NextSyncCommittee),
		&lcu.NextSyncCommitteeBranch,
		&lcu.FinalizedHeader,
		&lcu.FinalityBranch,
		spec.Wrap(&lcu.SyncAggregate),
		&lcu.SignatureSlot,
	)
}

func (lcu *LightClientUpdate) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&lcu.AttestedHeader,
		spec.Wrap(&lcu.NextSyncCommittee),
		&lcu.NextSyncCommitteeBranch,
		&lcu.FinalizedHeader,
		&lcu.FinalityBranch,
		spec.Wrap(&lcu.SyncAggregate),
		&lcu.SignatureSlot,
	)
}

func (lcu *LightClientUpdate) FixedLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&lcu.AttestedHeader,
		spec.Wrap(&lcu.NextSyncCommittee),
		&lcu.NextSyncCommitteeBranch,
		&lcu.FinalizedHeader,
		&lcu.FinalityBranch,
		spec.Wrap(&lcu.SyncAggregate),
		&lcu.SignatureSlot,
	)
}

func (lcu *LightClientUpdate) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		&lcu.AttestedHeader,
		spec.Wrap(&lcu.NextSyncCommittee),
		&lcu.NextSyncCommitteeBranch,
		&lcu.FinalizedHeader,
		&lcu.FinalityBranch,
		spec.Wrap(&lcu.SyncAggregate),
		&lcu.SignatureSlot,
	)
}

type LightClientFinalityUpdate struct {
	AttestedHeader  LightClientHeader               `yaml:"attested_header" json:"attested_header"`
	FinalizedHeader LightClientHeader        `yaml:"finalized_header" json:"finalized_header"`
	FinalityBranch  altair.FinalizedRootProofBranch `yaml:"finality_branch" json:"finality_branch"`
	SyncAggregate   altair.SyncAggregate            `yaml:"sync_aggregate" json:"sync_aggregate"`
	SignatureSlot   common.Slot                     `yaml:"signature_slot" json:"signature_slot"`
}

func LightClientFinalityUpdateType(spec *common.Spec) *view.ContainerTypeDef {
	return view.ContainerType("SyncCommittee", []view.FieldDef{
		{Name: "attested_header", Type: LightClientHeaderType},
		{Name: "finalized_header", Type: LightClientHeaderType},
		{Name: "finality_branch", Type: altair.FinalizedRootProofBranchType},
		{Name: "sync_aggregate", Type: altair.SyncAggregateType(spec)},
		{Name: "signature_slot", Type: common.SlotType},
	})
}

func (lcfu *LightClientFinalityUpdate) FixedLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&lcfu.AttestedHeader, &lcfu.FinalizedHeader, &lcfu.FinalityBranch, spec.Wrap(&lcfu.SyncAggregate), &lcfu.SignatureSlot)
}

func (lcfu *LightClientFinalityUpdate) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&lcfu.AttestedHeader, &lcfu.FinalizedHeader, &lcfu.FinalityBranch, spec.Wrap(&lcfu.SyncAggregate), &lcfu.SignatureSlot)
}

func (lcfu *LightClientFinalityUpdate) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&lcfu.AttestedHeader, &lcfu.FinalizedHeader, &lcfu.FinalityBranch, spec.Wrap(&lcfu.SyncAggregate), &lcfu.SignatureSlot)
}

func (lcfu *LightClientFinalityUpdate) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&lcfu.AttestedHeader, &lcfu.FinalizedHeader, &lcfu.FinalityBranch, spec.Wrap(&lcfu.SyncAggregate), &lcfu.SignatureSlot)
}

func (lcfu *LightClientFinalityUpdate) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		&lcfu.AttestedHeader,
		&lcfu.FinalizedHeader,
		&lcfu.FinalityBranch,
		spec.Wrap(&lcfu.SyncAggregate),
		&lcfu.SignatureSlot,
	)
}

type LightClientOptimisticUpdate struct {
	AttestedHeader LightClientHeader    `yaml:"attested_header" json:"attested_header"`
	SyncAggregate  altair.SyncAggregate `yaml:"sync_aggregate" json:"sync_aggregate"`
	SignatureSlot  common.Slot          `yaml:"signature_slot" json:"signature_slot"`
}

func LightClientOptimisticUpdateType(spec *common.Spec) *view.ContainerTypeDef {
	return view.ContainerType("SyncCommittee", []view.FieldDef{
		{Name: "attested_header", Type: LightClientHeaderType},
		{Name: "finalized_header", Type: common.BeaconBlockHeaderType},
		{Name: "finality_branch", Type: altair.FinalizedRootProofBranchType},
		{Name: "sync_aggregate", Type: altair.SyncAggregateType(spec)},
		{Name: "signature_slot", Type: common.SlotType},
	})
}

func (lcou *LightClientOptimisticUpdate) FixedLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&lcou.AttestedHeader, spec.Wrap(&lcou.SyncAggregate), &lcou.SignatureSlot)
}

func (lcou *LightClientOptimisticUpdate) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&lcou.AttestedHeader, spec.Wrap(&lcou.SyncAggregate), &lcou.SignatureSlot)
}

func (lcou *LightClientOptimisticUpdate) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&lcou.AttestedHeader, spec.Wrap(&lcou.SyncAggregate), &lcou.SignatureSlot)
}

func (lcou *LightClientOptimisticUpdate) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&lcou.AttestedHeader, spec.Wrap(&lcou.SyncAggregate), &lcou.SignatureSlot)
}

func (lcou *LightClientOptimisticUpdate) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		&lcou.AttestedHeader,
		spec.Wrap(&lcou.SyncAggregate),
		&lcou.SignatureSlot,
	)
}
