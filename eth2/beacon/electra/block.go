package electra

import (
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type SignedBeaconBlock struct {
	Message   BeaconBlock         `json:"message" yaml:"message"`
	Signature common.BLSSignature `json:"signature" yaml:"signature"`
}

var _ common.EnvelopeBuilder = (*SignedBeaconBlock)(nil)

func (b *SignedBeaconBlock) Envelope(spec *common.Spec, digest common.ForkDigest) *common.BeaconBlockEnvelope {
	header := b.Message.Header(spec)
	return &common.BeaconBlockEnvelope{
		ForkDigest:        digest,
		BeaconBlockHeader: *header,
		Body:              &b.Message.Body,
		BlockRoot:         header.HashTreeRoot(tree.GetHashFn()),
		Signature:         b.Signature,
	}
}

func (b *SignedBeaconBlock) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedBeaconBlock) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedBeaconBlock) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&b.Message), &b.Signature)
}

func (a *SignedBeaconBlock) FixedLength(*common.Spec) uint64 {
	return 0
}

func (b *SignedBeaconBlock) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&b.Message), b.Signature)
}

func (block *SignedBeaconBlock) SignedHeader(spec *common.Spec) *common.SignedBeaconBlockHeader {
	return &common.SignedBeaconBlockHeader{
		Message:   *block.Message.Header(spec),
		Signature: block.Signature,
	}
}

type BeaconBlock struct {
	Slot          common.Slot           `json:"slot" yaml:"slot"`
	ProposerIndex common.ValidatorIndex `json:"proposer_index" yaml:"proposer_index"`
	ParentRoot    common.Root           `json:"parent_root" yaml:"parent_root"`
	StateRoot     common.Root           `json:"state_root" yaml:"state_root"`
	Body          BeaconBlockBody       `json:"body" yaml:"body"`
}

func (b *BeaconBlock) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&b.Slot, &b.ProposerIndex, &b.ParentRoot, &b.StateRoot, spec.Wrap(&b.Body))
}

func (b *BeaconBlock) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&b.Slot, &b.ProposerIndex, &b.ParentRoot, &b.StateRoot, spec.Wrap(&b.Body))
}

func (b *BeaconBlock) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&b.Slot, &b.ProposerIndex, &b.ParentRoot, &b.StateRoot, spec.Wrap(&b.Body))
}

func (a *BeaconBlock) FixedLength(*common.Spec) uint64 {
	return 0
}

func (b *BeaconBlock) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(b.Slot, b.ProposerIndex, b.ParentRoot, b.StateRoot, spec.Wrap(&b.Body))
}

func BeaconBlockType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("BeaconBlock", []FieldDef{
		{"slot", common.SlotType},
		{"proposer_index", common.ValidatorIndexType},
		{"parent_root", RootType},
		{"state_root", RootType},
		{"body", BeaconBlockBodyType(spec)},
	})
}

func SignedBeaconBlockType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("SignedBeaconBlock", []FieldDef{
		{"message", BeaconBlockType(spec)},
		{"signature", common.BLSSignatureType},
	})
}

func (block *BeaconBlock) Header(spec *common.Spec) *common.BeaconBlockHeader {
	return &common.BeaconBlockHeader{
		Slot:          block.Slot,
		ProposerIndex: block.ProposerIndex,
		ParentRoot:    block.ParentRoot,
		StateRoot:     block.StateRoot,
		BodyRoot:      block.Body.HashTreeRoot(spec, tree.GetHashFn()),
	}
}

type BeaconBlockBody struct {
	RandaoReveal common.BLSSignature `json:"randao_reveal" yaml:"randao_reveal"`
	Eth1Data     common.Eth1Data     `json:"eth1_data" yaml:"eth1_data"`
	Graffiti     common.Root         `json:"graffiti" yaml:"graffiti"`

	// Operations
	ProposerSlashings phase0.ProposerSlashings `json:"proposer_slashings" yaml:"proposer_slashings"`
	// [Modified in Electra:EIP7549], MAX_ATTESTER_SLASHINGS_ELECTRA
	AttesterSlashings AttesterSlashings `json:"attester_slashings" yaml:"attester_slashings"`
	// [Modified in Electra:EIP7549], MAX_ATTESTATIONS_ELECTRA
	Attestations   Attestations          `json:"attestations" yaml:"attestations"`
	Deposits       phase0.Deposits       `json:"deposits" yaml:"deposits"`
	VoluntaryExits phase0.VoluntaryExits `json:"voluntary_exits" yaml:"voluntary_exits"`
	SyncAggregate  altair.SyncAggregate  `json:"sync_aggregate" yaml:"sync_aggregate"`
	// Execution
	ExecutionPayload      deneb.ExecutionPayload             `json:"execution_payload" yaml:"execution_payload"`
	BLSToExecutionChanges common.SignedBLSToExecutionChanges `json:"bls_to_execution_changes" yaml:"bls_to_execution_changes"`
	BlobKZGCommitments    deneb.KZGCommitments               `json:"blob_kzg_commitments" yaml:"blob_kzg_commitments"`
	// [New in Electra]
	ExecutionRequests ExecutionRequests `json:"execution_requests" yaml:"execution_requests"`
}

func (b *BeaconBlockBody) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(
		&b.RandaoReveal, &b.Eth1Data,
		&b.Graffiti, spec.Wrap(&b.ProposerSlashings),
		spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations),
		spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
		spec.Wrap(&b.SyncAggregate), spec.Wrap(&b.ExecutionPayload),
		spec.Wrap(&b.BLSToExecutionChanges),
		spec.Wrap(&b.BlobKZGCommitments),
		spec.Wrap(&b.ExecutionRequests),
	)
}

func (b *BeaconBlockBody) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(
		&b.RandaoReveal, &b.Eth1Data,
		&b.Graffiti, spec.Wrap(&b.ProposerSlashings),
		spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations),
		spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
		spec.Wrap(&b.SyncAggregate), spec.Wrap(&b.ExecutionPayload),
		spec.Wrap(&b.BLSToExecutionChanges),
		spec.Wrap(&b.BlobKZGCommitments),
		spec.Wrap(&b.ExecutionRequests),
	)
}

func (b *BeaconBlockBody) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&b.RandaoReveal, &b.Eth1Data,
		&b.Graffiti, spec.Wrap(&b.ProposerSlashings),
		spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations),
		spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
		spec.Wrap(&b.SyncAggregate), spec.Wrap(&b.ExecutionPayload),
		spec.Wrap(&b.BLSToExecutionChanges),
		spec.Wrap(&b.BlobKZGCommitments),
		spec.Wrap(&b.ExecutionRequests),
	)
}

func (a *BeaconBlockBody) FixedLength(*common.Spec) uint64 {
	return 0
}

func (b *BeaconBlockBody) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		b.RandaoReveal, &b.Eth1Data,
		b.Graffiti, spec.Wrap(&b.ProposerSlashings),
		spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations),
		spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
		spec.Wrap(&b.SyncAggregate), spec.Wrap(&b.ExecutionPayload),
		spec.Wrap(&b.BLSToExecutionChanges),
		spec.Wrap(&b.BlobKZGCommitments),
		spec.Wrap(&b.ExecutionRequests),
	)
}

func (b *BeaconBlockBody) CheckLimits(spec *common.Spec) error {
	if x := uint64(len(b.ProposerSlashings)); x > uint64(spec.MAX_PROPOSER_SLASHINGS) {
		return fmt.Errorf("too many proposer slashings: %d", x)
	}
	if x := uint64(len(b.AttesterSlashings)); x > uint64(spec.MAX_ATTESTER_SLASHINGS_ELECTRA) {
		return fmt.Errorf("too many attester slashings: %d", x)
	}
	if x := uint64(len(b.Attestations)); x > uint64(spec.MAX_ATTESTATIONS_ELECTRA) {
		return fmt.Errorf("too many attestations: %d", x)
	}
	if x := uint64(len(b.Deposits)); x > uint64(spec.MAX_DEPOSITS) {
		return fmt.Errorf("too many deposits: %d", x)
	}
	if x := uint64(len(b.VoluntaryExits)); x > uint64(spec.MAX_VOLUNTARY_EXITS) {
		return fmt.Errorf("too many voluntary exits: %d", x)
	}
	// TODO: also check sum of byte size, sanity check block size.
	if x := uint64(len(b.ExecutionPayload.Transactions)); x > uint64(spec.MAX_TRANSACTIONS_PER_PAYLOAD) {
		return fmt.Errorf("too many transactions: %d", x)
	}
	if x := uint64(len(b.BLSToExecutionChanges)); x > uint64(spec.MAX_BLS_TO_EXECUTION_CHANGES) {
		return fmt.Errorf("too many bls-to-execution changes: %d", x)
	}
	if x := uint64(len(b.BlobKZGCommitments)); x > uint64(spec.MAX_BLOBS_PER_BLOCK) {
		return fmt.Errorf("too many blob kzg commitments: %d", x)
	}
	if x := uint64(len(b.ExecutionRequests.Deposits)); x > uint64(spec.MAX_DEPOSIT_REQUESTS_PER_PAYLOAD) {
		return fmt.Errorf("too many execution request deposits: %d", x)
	}
	if x := uint64(len(b.ExecutionRequests.Withdrawals)); x > uint64(spec.MAX_WITHDRAWAL_REQUESTS_PER_PAYLOAD) {
		return fmt.Errorf("too many execution request withdrawals: %d", x)
	}
	if x := uint64(len(b.ExecutionRequests.Consolidations)); x > uint64(spec.MAX_CONSOLIDATION_REQUESTS_PER_PAYLOAD) {
		return fmt.Errorf("too many execution request consolidations: %d", x)
	}
	return nil
}

func (b *BeaconBlockBody) Shallow(spec *common.Spec) *BeaconBlockBodyShallow {
	return &BeaconBlockBodyShallow{
		RandaoReveal:          b.RandaoReveal,
		Eth1Data:              b.Eth1Data,
		Graffiti:              b.Graffiti,
		ProposerSlashings:     b.ProposerSlashings,
		AttesterSlashings:     b.AttesterSlashings,
		Attestations:          b.Attestations,
		Deposits:              b.Deposits,
		VoluntaryExits:        b.VoluntaryExits,
		SyncAggregate:         b.SyncAggregate,
		ExecutionPayloadRoot:  b.ExecutionPayload.HashTreeRoot(spec, tree.GetHashFn()),
		BLSToExecutionChanges: b.BLSToExecutionChanges,
		BlobKZGCommitments:    b.BlobKZGCommitments,
		ExecutionRequests:     b.ExecutionRequests,
	}
}

func (b *BeaconBlockBody) GetTransactions() []common.Transaction {
	return b.ExecutionPayload.Transactions
}

func (b *BeaconBlockBody) GetBlobKZGCommitments() []common.KZGCommitment {
	return b.BlobKZGCommitments
}

func BeaconBlockBodyType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("BeaconBlockBody", []FieldDef{
		{"randao_reveal", common.BLSSignatureType},
		{"eth1_data", common.Eth1DataType}, // Eth1 data vote
		{"graffiti", common.Bytes32Type},   // Arbitrary data
		// Operations
		{"proposer_slashings", phase0.BlockProposerSlashingsType(spec)},
		{"attester_slashings", BlockAttesterSlashingsType(spec)},
		{"attestations", BlockAttestationsType(spec)},
		{"deposits", phase0.BlockDepositsType(spec)},
		{"voluntary_exits", phase0.BlockVoluntaryExitsType(spec)},
		{"sync_aggregate", altair.SyncAggregateType(spec)},
		// Capella
		{"execution_payload", deneb.ExecutionPayloadType(spec)},
		{"bls_to_execution_changes", common.BlockSignedBLSToExecutionChangesType(spec)},
		// Deneb
		{"blob_kzg_commitments", deneb.KZGCommitmentsType(spec)},
		// Electra
		{"execution_requests", ExecutionRequestsType(spec)},
	})
}

type BeaconBlockBodyShallow struct {
	RandaoReveal common.BLSSignature `json:"randao_reveal" yaml:"randao_reveal"`
	Eth1Data     common.Eth1Data     `json:"eth1_data" yaml:"eth1_data"`
	Graffiti     common.Root         `json:"graffiti" yaml:"graffiti"`

	ProposerSlashings phase0.ProposerSlashings `json:"proposer_slashings" yaml:"proposer_slashings"`
	AttesterSlashings AttesterSlashings        `json:"attester_slashings" yaml:"attester_slashings"`
	Attestations      Attestations             `json:"attestations" yaml:"attestations"`
	Deposits          phase0.Deposits          `json:"deposits" yaml:"deposits"`
	VoluntaryExits    phase0.VoluntaryExits    `json:"voluntary_exits" yaml:"voluntary_exits"`

	SyncAggregate altair.SyncAggregate `json:"sync_aggregate" yaml:"sync_aggregate"`

	ExecutionPayloadRoot common.Root `json:"execution_payload_root" yaml:"execution_payload_root"`

	BLSToExecutionChanges common.SignedBLSToExecutionChanges `json:"bls_to_execution_changes" yaml:"bls_to_execution_changes"`

	BlobKZGCommitments deneb.KZGCommitments `json:"blob_kzg_commitments" yaml:"blob_kzg_commitments"` // new in EIP-4844
	// [New in Electra]
	ExecutionRequests ExecutionRequests `json:"execution_requests" yaml:"execution_requests"`
}

func (b *BeaconBlockBodyShallow) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(
		&b.RandaoReveal, &b.Eth1Data,
		&b.Graffiti, spec.Wrap(&b.ProposerSlashings),
		spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations),
		spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
		spec.Wrap(&b.SyncAggregate), &b.ExecutionPayloadRoot,
		spec.Wrap(&b.BLSToExecutionChanges),
		spec.Wrap(&b.BlobKZGCommitments),
		spec.Wrap(&b.ExecutionRequests),
	)
}

func (b *BeaconBlockBodyShallow) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(
		&b.RandaoReveal, &b.Eth1Data,
		&b.Graffiti, spec.Wrap(&b.ProposerSlashings),
		spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations),
		spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
		spec.Wrap(&b.SyncAggregate), &b.ExecutionPayloadRoot,
		spec.Wrap(&b.BLSToExecutionChanges),
		spec.Wrap(&b.BlobKZGCommitments),
		spec.Wrap(&b.ExecutionRequests),
	)
}

func (b *BeaconBlockBodyShallow) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		&b.RandaoReveal, &b.Eth1Data,
		&b.Graffiti, spec.Wrap(&b.ProposerSlashings),
		spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations),
		spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
		spec.Wrap(&b.SyncAggregate), &b.ExecutionPayloadRoot,
		spec.Wrap(&b.BLSToExecutionChanges),
		spec.Wrap(&b.BlobKZGCommitments),
		spec.Wrap(&b.ExecutionRequests),
	)
}

func (a *BeaconBlockBodyShallow) FixedLength(*common.Spec) uint64 {
	return 0
}

func (b *BeaconBlockBodyShallow) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		b.RandaoReveal, &b.Eth1Data,
		b.Graffiti, spec.Wrap(&b.ProposerSlashings),
		spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations),
		spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
		spec.Wrap(&b.SyncAggregate), &b.ExecutionPayloadRoot,
		spec.Wrap(&b.BLSToExecutionChanges),
		spec.Wrap(&b.BlobKZGCommitments),
		spec.Wrap(&b.ExecutionRequests),
	)
}

func (b *BeaconBlockBodyShallow) WithExecutionPayload(spec *common.Spec, payload deneb.ExecutionPayload) (*BeaconBlockBody, error) {
	payloadRoot := payload.HashTreeRoot(spec, tree.GetHashFn())
	if b.ExecutionPayloadRoot != payloadRoot {
		return nil, fmt.Errorf("payload does not match expected root: %s <> %s", b.ExecutionPayloadRoot, payloadRoot)
	}
	return &BeaconBlockBody{
		RandaoReveal:          b.RandaoReveal,
		Eth1Data:              b.Eth1Data,
		Graffiti:              b.Graffiti,
		ProposerSlashings:     b.ProposerSlashings,
		AttesterSlashings:     b.AttesterSlashings,
		Attestations:          b.Attestations,
		Deposits:              b.Deposits,
		VoluntaryExits:        b.VoluntaryExits,
		SyncAggregate:         b.SyncAggregate,
		ExecutionPayload:      payload,
		BLSToExecutionChanges: b.BLSToExecutionChanges,
		BlobKZGCommitments:    b.BlobKZGCommitments,
		ExecutionRequests:     b.ExecutionRequests,
	}, nil
}
