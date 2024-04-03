package altair

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

func LightClientSnapshotType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("SyncCommittee", []FieldDef{
		{"header", common.BeaconBlockHeaderType},
		{"current_sync_committee", common.SyncCommitteeType(spec)},
		{"next_sync_committee", common.SyncCommitteeType(spec)},
	})
}

type LightClientSnapshot struct {
	// Beacon block header
	Header common.BeaconBlockHeader `yaml:"header" json:"header"`
	// Sync committees corresponding to the header
	CurrentSyncCommittee common.SyncCommittee `yaml:"current_sync_committee" json:"current_sync_committee"`
	NextSyncCommittee    common.SyncCommittee `yaml:"next_sync_committee" json:"next_sync_committee"`
}

func (lcs *LightClientSnapshot) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(
		&lcs.Header,
		spec.Wrap(&lcs.CurrentSyncCommittee),
		spec.Wrap(&lcs.NextSyncCommittee),
	)
}

func (lcs *LightClientSnapshot) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.FixedLenContainer(
		&lcs.Header,
		spec.Wrap(&lcs.CurrentSyncCommittee),
		spec.Wrap(&lcs.NextSyncCommittee),
	)
}

func (lcs *LightClientSnapshot) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&lcs.Header,
		spec.Wrap(&lcs.CurrentSyncCommittee),
		spec.Wrap(&lcs.NextSyncCommittee),
	)
}

func (lcs *LightClientSnapshot) FixedLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&lcs.Header,
		spec.Wrap(&lcs.CurrentSyncCommittee),
		spec.Wrap(&lcs.NextSyncCommittee),
	)
}

func (lcs *LightClientSnapshot) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		&lcs.Header,
		spec.Wrap(&lcs.CurrentSyncCommittee),
		spec.Wrap(&lcs.NextSyncCommittee),
	)
}

// The BeaconState has 24 fields
// This is padded to 32, a depth of 5 bits
const syncCommitteeProofLen = 5

const NEXT_SYNC_COMMITTEE_INDEX = tree.Gindex64((1 << syncCommitteeProofLen) | _nextSyncCommittee)

var SyncCommitteeProofBranchType = VectorType(RootType, syncCommitteeProofLen)

type SyncCommitteeProofBranch [syncCommitteeProofLen]common.Root

func (sb *SyncCommitteeProofBranch) Deserialize(dr *codec.DecodingReader) error {
	roots := sb[:]
	return tree.ReadRoots(dr, &roots, syncCommitteeProofLen)
}

func (sb SyncCommitteeProofBranch) Serialize(w *codec.EncodingWriter) error {
	return tree.WriteRoots(w, sb[:])
}

func (sb SyncCommitteeProofBranch) ByteLength() (out uint64) {
	return syncCommitteeProofLen * 32
}

func (sb *SyncCommitteeProofBranch) FixedLength() uint64 {
	return syncCommitteeProofLen * 32
}

func (sb SyncCommitteeProofBranch) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.ComplexVectorHTR(func(i uint64) tree.HTR {
		if i < syncCommitteeProofLen {
			return &sb[i]
		}
		return nil
	}, syncCommitteeProofLen)
}

// Like the above, 5 bits deep, plus 1 for the checkpoint (it has two fields, we take the 2nd)
const finalizedRootProofLen = 5 + 1

const FINALIZED_ROOT_INDEX = tree.Gindex64((1 << finalizedRootProofLen) | (_stateFinalizedCheckpoint << 1) | 1)

var FinalizedRootProofBranchType = VectorType(RootType, finalizedRootProofLen)

type FinalizedRootProofBranch [finalizedRootProofLen]common.Root

func (fb *FinalizedRootProofBranch) Deserialize(dr *codec.DecodingReader) error {
	roots := fb[:]
	return tree.ReadRoots(dr, &roots, finalizedRootProofLen)
}

func (fb FinalizedRootProofBranch) Serialize(w *codec.EncodingWriter) error {
	return tree.WriteRoots(w, fb[:])
}

func (fb FinalizedRootProofBranch) ByteLength() (out uint64) {
	return finalizedRootProofLen * 32
}

func (fb *FinalizedRootProofBranch) FixedLength() uint64 {
	return finalizedRootProofLen * 32
}

func (fb FinalizedRootProofBranch) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.ComplexVectorHTR(func(i uint64) tree.HTR {
		if i < finalizedRootProofLen {
			return &fb[i]
		}
		return nil
	}, finalizedRootProofLen)
}

func LightClientUpdateType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("SyncCommittee", []FieldDef{
		{"attested_header", LightClientHeaderType},
		{"next_sync_committee", common.SyncCommitteeType(spec)},
		{"next_sync_committee_branch", SyncCommitteeProofBranchType},
		{"finalized_header", LightClientHeaderType},
		{"finality_branch", FinalizedRootProofBranchType},
		{"sync_aggregate", SyncAggregateType(spec)},
		{"signature_slot", common.SlotType},
	})
}

type LightClientUpdate struct {
	// Update beacon block header
	AttestedHeader LightClientHeader `yaml:"attested_header" json:"attested_header"`
	// Next sync committee corresponding to the header
	NextSyncCommittee       common.SyncCommittee     `yaml:"next_sync_committee" json:"next_sync_committee"`
	NextSyncCommitteeBranch SyncCommitteeProofBranch `yaml:"next_sync_committee_branch" json:"next_sync_committee_branch"`
	// Finality proof for the update header
	FinalizedHeader LightClientHeader `yaml:"finalized_header" json:"finalized_header"`
	FinalityBranch  FinalizedRootProofBranch `yaml:"finality_branch" json:"finality_branch"`
	// Sync committee aggregate signature
	SyncAggregate SyncAggregate `yaml:"sync_aggregate" json:"sync_aggregate"`
	// Slot at which the aggregate signature was created (untrusted)
	SignatureSlot common.Slot `yaml:"signature_slot" json:"signature_slot"`
}

func (lcu *LightClientUpdate) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(
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
	return w.FixedLenContainer(
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

var LightClientHeaderType = ContainerType("LightClientHeaderType", []FieldDef{
	{Name: "beacon", Type: common.BeaconBlockHeaderType},
})

type LightClientHeader struct {
	Beacon common.BeaconBlockHeader `yaml:"beacon" json:"beacon"`
}

func (lch *LightClientHeader) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(
		&lch.Beacon,
	)
}

func (lch *LightClientHeader) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(
		&lch.Beacon,
	)
}

func (lch *LightClientHeader) ByteLength() uint64 {
	return codec.ContainerLength(
		&lch.Beacon,
	)
}

func (lch *LightClientHeader) FixedLength() uint64 {
	return codec.ContainerLength(
		&lch.Beacon,
	)
}

func (lch *LightClientHeader) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		&lch.Beacon,
	)
}

type LightClientBootstrap struct {
	Header                     LightClientHeader        `yaml:"header" json:"header"`
	CurrentSyncCommittee       common.SyncCommittee     `yaml:"current_sync_committee" json:"current_sync_committee"`
	CurrentSyncCommitteeBranch SyncCommitteeProofBranch `yaml:"current_sync_committee_branch" json:"current_sync_committee_branch"`
}

func NewLightClientBootstrapType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("LightClientHeader", []FieldDef{
		{Name: "header", Type: LightClientHeaderType},
		{Name: "current_sync_committee", Type: common.SyncCommitteeType(spec)},
		{Name: "current_sync_committee_branch", Type: SyncCommitteeProofBranchType},
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

type LightClientFinalityUpdate struct {
	AttestedHeader  LightClientHeader        `yaml:"attested_header" json:"attested_header"`
	FinalizedHeader common.BeaconBlockHeader `yaml:"finalized_header" json:"finalized_header"`
	FinalityBranch  FinalizedRootProofBranch `yaml:"finality_branch" json:"finality_branch"`
	SyncAggregate   SyncAggregate            `yaml:"sync_aggregate" json:"sync_aggregate"`
	SignatureSlot   common.Slot              `yaml:"signature_slot" json:"signature_slot"`
}

func LightClientFinalityUpdateType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("SyncCommittee", []FieldDef{
		{Name: "attested_header", Type: common.BeaconBlockHeaderType},
		{Name: "finalized_header", Type: common.BeaconBlockHeaderType},
		{Name: "finality_branch", Type: FinalizedRootProofBranchType},
		{Name: "sync_aggregate", Type: SyncAggregateType(spec)},
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
	AttestedHeader LightClientHeader `yaml:"attested_header" json:"attested_header"`
	SyncAggregate  SyncAggregate     `yaml:"sync_aggregate" json:"sync_aggregate"`
	SignatureSlot  common.Slot       `yaml:"signature_slot" json:"signature_slot"`
}

func LightClientOptimisticUpdateType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("SyncCommittee", []FieldDef{
		{Name: "attested_header", Type: common.BeaconBlockHeaderType},
		{Name: "finalized_header", Type: common.BeaconBlockHeaderType},
		{Name: "finality_branch", Type: FinalizedRootProofBranchType},
		{Name: "sync_aggregate", Type: SyncAggregateType(spec)},
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
