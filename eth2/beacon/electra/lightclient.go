package electra

import (
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
)

const currentSyncCommitteeBranchLen = 6

type CurrentSyncCommitteeBranch [currentSyncCommitteeBranchLen]common.Root

func (fb *CurrentSyncCommitteeBranch) Deserialize(dr *codec.DecodingReader) error {
	roots := fb[:]
	return tree.ReadRoots(dr, &roots, currentSyncCommitteeBranchLen)
}

func (fb CurrentSyncCommitteeBranch) Serialize(w *codec.EncodingWriter) error {
	return tree.WriteRoots(w, fb[:])
}

func (fb CurrentSyncCommitteeBranch) ByteLength() (out uint64) {
	return currentSyncCommitteeBranchLen * 32
}

func (fb *CurrentSyncCommitteeBranch) FixedLength() uint64 {
	return currentSyncCommitteeBranchLen * 32
}

func (fb CurrentSyncCommitteeBranch) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.ComplexVectorHTR(func(i uint64) tree.HTR {
		if i < currentSyncCommitteeBranchLen {
			return &fb[i]
		}
		return nil
	}, currentSyncCommitteeBranchLen)
}

type LightClientBootstrap struct {
	Header                     deneb.LightClientHeader    `yaml:"header" json:"header"`
	CurrentSyncCommittee       common.SyncCommittee       `yaml:"current_sync_committee" json:"current_sync_committee"`
	CurrentSyncCommitteeBranch CurrentSyncCommitteeBranch `yaml:"current_sync_committee_branch" json:"current_sync_committee_branch"`
}

func (lcb *LightClientBootstrap) FixedLength(spec *common.Spec) uint64 {
	return 0
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

const finalizedRootProofLen = 7

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

type LightClientUpdate struct {
	// Update beacon block header
	AttestedHeader deneb.LightClientHeader `yaml:"attested_header" json:"attested_header"`
	// Next sync committee corresponding to the header
	NextSyncCommittee       common.SyncCommittee       `yaml:"next_sync_committee" json:"next_sync_committee"`
	NextSyncCommitteeBranch CurrentSyncCommitteeBranch `yaml:"next_sync_committee_branch" json:"next_sync_committee_branch"`
	// Finality proof for the update header
	FinalizedHeader deneb.LightClientHeader  `yaml:"finalized_header" json:"finalized_header"`
	FinalityBranch  FinalizedRootProofBranch `yaml:"finality_branch" json:"finality_branch"`
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
	return 0
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
	AttestedHeader  deneb.LightClientHeader  `yaml:"attested_header" json:"attested_header"`
	FinalizedHeader deneb.LightClientHeader  `yaml:"finalized_header" json:"finalized_header"`
	FinalityBranch  FinalizedRootProofBranch `yaml:"finality_branch" json:"finality_branch"`
	SyncAggregate   altair.SyncAggregate     `yaml:"sync_aggregate" json:"sync_aggregate"`
	SignatureSlot   common.Slot              `yaml:"signature_slot" json:"signature_slot"`
}

func (lcfu *LightClientFinalityUpdate) FixedLength(spec *common.Spec) uint64 {
	return 0
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
