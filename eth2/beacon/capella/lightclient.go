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
	Beacon          common.BeaconBlockHeader
	Execution       ExecutionPayloadHeader
	ExecutionBranch ExecutionBranch
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
	return LightClientHeaderType.TypeByteLength()
}

func (l *LightClientHeader) HashTreeRoot(h tree.HashFn) common.Root {
	return h.HashTreeRoot(&l.Beacon, &l.Execution, &l.ExecutionBranch)
}

type LightClientBootstrap struct {
	Header                     LightClientHeader
	CurrentSyncCommittee       common.SyncCommittee
	CurrentSyncCommitteeBranch altair.SyncCommitteeProofBranch
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
