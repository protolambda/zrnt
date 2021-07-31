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

func (lcu *LightClientUpdate) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(
		&lcu.Header,
		spec.Wrap(&lcu.NextSyncCommittee),
		&lcu.NextSyncCommitteeBranch,
		&lcu.FinalityHeader,
		&lcu.FinalityBranch,
		spec.Wrap(&lcu.SyncCommitteeBits),
		&lcu.SyncCommitteeSignature,
		&lcu.ForkVersion,
	)
}

func (lcu *LightClientUpdate) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.FixedLenContainer(
		&lcu.Header,
		spec.Wrap(&lcu.NextSyncCommittee),
		&lcu.NextSyncCommitteeBranch,
		&lcu.FinalityHeader,
		&lcu.FinalityBranch,
		spec.Wrap(&lcu.SyncCommitteeBits),
		&lcu.SyncCommitteeSignature,
		&lcu.ForkVersion,
	)
}

func (lcu *LightClientUpdate) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&lcu.Header,
		spec.Wrap(&lcu.NextSyncCommittee),
		&lcu.NextSyncCommitteeBranch,
		&lcu.FinalityHeader,
		&lcu.FinalityBranch,
		spec.Wrap(&lcu.SyncCommitteeBits),
		&lcu.SyncCommitteeSignature,
		&lcu.ForkVersion,
	)
}

func (lcu *LightClientUpdate) FixedLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&lcu.Header,
		spec.Wrap(&lcu.NextSyncCommittee),
		&lcu.NextSyncCommitteeBranch,
		&lcu.FinalityHeader,
		&lcu.FinalityBranch,
		spec.Wrap(&lcu.SyncCommitteeBits),
		&lcu.SyncCommitteeSignature,
		&lcu.ForkVersion,
	)
}

func (lcu *LightClientUpdate) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		&lcu.Header,
		spec.Wrap(&lcu.NextSyncCommittee),
		&lcu.NextSyncCommitteeBranch,
		&lcu.FinalityHeader,
		&lcu.FinalityBranch,
		spec.Wrap(&lcu.SyncCommitteeBits),
		&lcu.SyncCommitteeSignature,
		&lcu.ForkVersion,
	)
}
