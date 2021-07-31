package altair

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
)

type LightClientSnapshot struct {
	// Beacon block header
	Header common.BeaconBlockHeader `yaml:"header" json:"header"`
	// Sync committees corresponding to the header
	CurrentSyncCommittee common.SyncCommittee `yaml:"current_sync_committee" json:"current_sync_committee"`
	NextSyncCommittee    common.SyncCommittee `yaml:"next_sync_committee" json:"next_sync_committee"`
}

func (agg *LightClientSnapshot) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(
		&agg.Header,
		spec.Wrap(&agg.CurrentSyncCommittee),
		spec.Wrap(&agg.NextSyncCommittee),
	)
}

func (agg *LightClientSnapshot) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.FixedLenContainer(
		&agg.Header,
		spec.Wrap(&agg.CurrentSyncCommittee),
		spec.Wrap(&agg.NextSyncCommittee),
	)
}

func (agg *LightClientSnapshot) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&agg.Header,
		spec.Wrap(&agg.CurrentSyncCommittee),
		spec.Wrap(&agg.NextSyncCommittee),
	)
}

func (agg *LightClientSnapshot) FixedLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&agg.Header,
		spec.Wrap(&agg.CurrentSyncCommittee),
		spec.Wrap(&agg.NextSyncCommittee),
	)
}

func (agg *LightClientSnapshot) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		&agg.Header,
		spec.Wrap(&agg.CurrentSyncCommittee),
		spec.Wrap(&agg.NextSyncCommittee),
	)
}

// The BeaconState has 24 fields
// This is padded to 32, a depth of 5 bits
const syncCommitteeProofLen = 5

const NEXT_SYNC_COMMITTEE_INDEX = tree.Gindex64((1 << syncCommitteeProofLen) | _nextSyncCommittee)

type SyncCommitteeProofBranch [syncCommitteeProofLen]common.Root

func (a *SyncCommitteeProofBranch) Deserialize(dr *codec.DecodingReader) error {
	roots := a[:]
	return tree.ReadRoots(dr, &roots, syncCommitteeProofLen)
}

func (a SyncCommitteeProofBranch) Serialize(w *codec.EncodingWriter) error {
	return tree.WriteRoots(w, a[:])
}

func (a SyncCommitteeProofBranch) ByteLength() (out uint64) {
	return syncCommitteeProofLen * 32
}

func (a *SyncCommitteeProofBranch) FixedLength() uint64 {
	return syncCommitteeProofLen * 32
}

func (li SyncCommitteeProofBranch) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.ComplexVectorHTR(func(i uint64) tree.HTR {
		if i < syncCommitteeProofLen {
			return &li[i]
		}
		return nil
	}, syncCommitteeProofLen)
}

// Like the above, 5 bits deep, plus 1 for the checkpoint (it has two fields, we take the 2nd)
const finalizedRootProofLen = 5 + 1

const FINALIZED_ROOT_INDEX = tree.Gindex64((1 << finalizedRootProofLen) | (_stateFinalizedCheckpoint << 1) | 1)

type FinalizedRootProofBranch [finalizedRootProofLen]common.Root

func (a *FinalizedRootProofBranch) Deserialize(dr *codec.DecodingReader) error {
	roots := a[:]
	return tree.ReadRoots(dr, &roots, finalizedRootProofLen)
}

func (a FinalizedRootProofBranch) Serialize(w *codec.EncodingWriter) error {
	return tree.WriteRoots(w, a[:])
}

func (a FinalizedRootProofBranch) ByteLength() (out uint64) {
	return finalizedRootProofLen * 32
}

func (a *FinalizedRootProofBranch) FixedLength() uint64 {
	return finalizedRootProofLen * 32
}

func (li FinalizedRootProofBranch) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.ComplexVectorHTR(func(i uint64) tree.HTR {
		if i < finalizedRootProofLen {
			return &li[i]
		}
		return nil
	}, finalizedRootProofLen)
}

type LightClientUpdate struct {
	// Update beacon block header
	Header common.BeaconBlockHeader `yaml:"header" json:"header"`
	// Next sync committee corresponding to the header
	NextSyncCommittee       common.SyncCommittee     `yaml:"next_sync_committee" json:"next_sync_committee"`
	NextSyncCommitteeBranch SyncCommitteeProofBranch `yaml:"next_sync_committee_branch" json:"next_sync_committee_branch"`
	// Finality proof for the update header
	FinalityHeader common.BeaconBlockHeader `yaml:"finality_header" json:"finality_header"`
	FinalityBranch FinalizedRootProofBranch `yaml:"finality_branch" json:"finality_branch"`
	// Sync committee aggregate signature
	SyncCommitteeBits      SyncCommitteeBits   `yaml:"sync_committee_bits" json:"sync_committee_bits"`
	SyncCommitteeSignature common.BLSSignature `yaml:"sync_committee_signature" json:"sync_committee_signature"`
	// Fork version for the aggregate signature
	ForkVersion common.Version `yaml:"fork_version" json:"fork_version"`
}

func (agg *LightClientUpdate) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(
		&agg.Header,
		spec.Wrap(&agg.NextSyncCommittee),
		&agg.NextSyncCommitteeBranch,
		&agg.FinalityHeader,
		&agg.FinalityBranch,
		spec.Wrap(&agg.SyncCommitteeBits),
		&agg.SyncCommitteeSignature,
		&agg.ForkVersion,
	)
}

func (agg *LightClientUpdate) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.FixedLenContainer(
		&agg.Header,
		spec.Wrap(&agg.NextSyncCommittee),
		&agg.NextSyncCommitteeBranch,
		&agg.FinalityHeader,
		&agg.FinalityBranch,
		spec.Wrap(&agg.SyncCommitteeBits),
		&agg.SyncCommitteeSignature,
		&agg.ForkVersion,
	)
}

func (agg *LightClientUpdate) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&agg.Header,
		spec.Wrap(&agg.NextSyncCommittee),
		&agg.NextSyncCommitteeBranch,
		&agg.FinalityHeader,
		&agg.FinalityBranch,
		spec.Wrap(&agg.SyncCommitteeBits),
		&agg.SyncCommitteeSignature,
		&agg.ForkVersion,
	)
}

func (agg *LightClientUpdate) FixedLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&agg.Header,
		spec.Wrap(&agg.NextSyncCommittee),
		&agg.NextSyncCommitteeBranch,
		&agg.FinalityHeader,
		&agg.FinalityBranch,
		spec.Wrap(&agg.SyncCommitteeBits),
		&agg.SyncCommitteeSignature,
		&agg.ForkVersion,
	)
}

func (agg *LightClientUpdate) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		&agg.Header,
		spec.Wrap(&agg.NextSyncCommittee),
		&agg.NextSyncCommitteeBranch,
		&agg.FinalityHeader,
		&agg.FinalityBranch,
		spec.Wrap(&agg.SyncCommitteeBits),
		&agg.SyncCommitteeSignature,
		&agg.ForkVersion,
	)
}
