package fulu

import (
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"
	"github.com/protolambda/zrnt/eth2/beacon/electra"
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

type BeaconBlock struct {
	Slot          common.Slot                 `json:"slot" yaml:"slot"`
	ProposerIndex common.ValidatorIndex       `json:"proposer_index" yaml:"proposer_index"`
	ParentRoot    common.Root                 `json:"parent_root" yaml:"parent_root"`
	StateRoot     common.Root                 `json:"state_root" yaml:"state_root"`
	Body          BeaconBlockBody             `json:"body" yaml:"body"`
}

func (b *BeaconBlock) Header(spec *common.Spec) *common.BeaconBlockHeader {
	return &common.BeaconBlockHeader{
		Slot:          b.Slot,
		ProposerIndex: b.ProposerIndex,
		ParentRoot:    b.ParentRoot,
		StateRoot:     b.StateRoot,
		BodyRoot:      b.Body.HashTreeRoot(spec, tree.GetHashFn()),
	}
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

func (b *BeaconBlock) FixedLength(*common.Spec) uint64 {
	return 0
}

func (b *BeaconBlock) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&b.Slot, &b.ProposerIndex, &b.ParentRoot, &b.StateRoot, spec.Wrap(&b.Body))
}

type BeaconBlockBody struct {
	RandaoReveal common.BLSSignature `json:"randao_reveal" yaml:"randao_reveal"`
	Eth1Data     common.Eth1Data     `json:"eth1_data" yaml:"eth1_data"`
	Graffiti     common.Root         `json:"graffiti" yaml:"graffiti"`
	// Operations
	ProposerSlashings phase0.ProposerSlashings         `json:"proposer_slashings" yaml:"proposer_slashings"`
	AttesterSlashings electra.AttesterSlashings        `json:"attester_slashings" yaml:"attester_slashings"`
	Attestations      electra.Attestations             `json:"attestations" yaml:"attestations"`
	Deposits          phase0.Deposits                  `json:"deposits" yaml:"deposits"`
	VoluntaryExits    phase0.VoluntaryExits            `json:"voluntary_exits" yaml:"voluntary_exits"`
	SyncAggregate     altair.SyncAggregate             `json:"sync_aggregate" yaml:"sync_aggregate"`
	// Execution payload
	ExecutionPayload                 ExecutionPayload                 `json:"execution_payload" yaml:"execution_payload"`
	BlsToExecutionChanges            common.SignedBLSToExecutionChanges `json:"bls_to_execution_changes" yaml:"bls_to_execution_changes"`
	BlobKzgCommitments               deneb.KZGCommitments           `json:"blob_kzg_commitments" yaml:"blob_kzg_commitments"`
}

func (b *BeaconBlockBody) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&b.RandaoReveal, &b.Eth1Data, &b.Graffiti,
		spec.Wrap(&b.ProposerSlashings), spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations), spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
		spec.Wrap(&b.SyncAggregate), spec.Wrap(&b.ExecutionPayload), spec.Wrap(&b.BlsToExecutionChanges), spec.Wrap(&b.BlobKzgCommitments))
}

func (b *BeaconBlockBody) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&b.RandaoReveal, &b.Eth1Data, &b.Graffiti,
		spec.Wrap(&b.ProposerSlashings), spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations), spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
		spec.Wrap(&b.SyncAggregate), spec.Wrap(&b.ExecutionPayload), spec.Wrap(&b.BlsToExecutionChanges), spec.Wrap(&b.BlobKzgCommitments))
}

func (b *BeaconBlockBody) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&b.RandaoReveal, &b.Eth1Data, &b.Graffiti,
		spec.Wrap(&b.ProposerSlashings), spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations), spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
		spec.Wrap(&b.SyncAggregate), spec.Wrap(&b.ExecutionPayload), spec.Wrap(&b.BlsToExecutionChanges), spec.Wrap(&b.BlobKzgCommitments))
}

func (b *BeaconBlockBody) FixedLength(*common.Spec) uint64 {
	return 0
}

func (b *BeaconBlockBody) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&b.RandaoReveal, &b.Eth1Data, &b.Graffiti,
		spec.Wrap(&b.ProposerSlashings), spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations), spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
		spec.Wrap(&b.SyncAggregate), spec.Wrap(&b.ExecutionPayload), spec.Wrap(&b.BlsToExecutionChanges), spec.Wrap(&b.BlobKzgCommitments))
}

type ExecutionPayload struct {
	ParentHash     common.Hash32          `json:"parent_hash" yaml:"parent_hash"`
	FeeRecipient   common.Eth1Address     `json:"fee_recipient" yaml:"fee_recipient"`
	StateRoot      common.Root            `json:"state_root" yaml:"state_root"`
	ReceiptsRoot   common.Root            `json:"receipts_root" yaml:"receipts_root"`
	LogsBloom      common.LogsBloom       `json:"logs_bloom" yaml:"logs_bloom"`
	PrevRandao     common.Hash32          `json:"prev_randao" yaml:"prev_randao"`
	BlockNumber    Uint64View             `json:"block_number" yaml:"block_number"`
	GasLimit       Uint64View             `json:"gas_limit" yaml:"gas_limit"`
	GasUsed        Uint64View             `json:"gas_used" yaml:"gas_used"`
	Timestamp      common.Timestamp       `json:"timestamp" yaml:"timestamp"`
	ExtraData      common.ExtraData       `json:"extra_data" yaml:"extra_data"`
	BaseFeePerGas  Uint256View            `json:"base_fee_per_gas" yaml:"base_fee_per_gas"`
	BlockHash      common.Hash32          `json:"block_hash" yaml:"block_hash"`
	Transactions   common.PayloadTransactions `json:"transactions" yaml:"transactions"`
	Withdrawals    common.Withdrawals         `json:"withdrawals" yaml:"withdrawals"`
	BlobGasUsed    Uint64View                 `json:"blob_gas_used" yaml:"blob_gas_used"`
	ExcessBlobGas  Uint64View                 `json:"excess_blob_gas" yaml:"excess_blob_gas"`
	DepositRequests    common.DepositRequests       `json:"deposit_requests" yaml:"deposit_requests"`
	WithdrawalRequests common.WithdrawalRequests    `json:"withdrawal_requests" yaml:"withdrawal_requests"`
	ConsolidationRequests common.ConsolidationRequests `json:"consolidation_requests" yaml:"consolidation_requests"`
}

func (b *ExecutionPayload) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&b.ParentHash, &b.FeeRecipient, &b.StateRoot, &b.ReceiptsRoot, &b.LogsBloom, &b.PrevRandao,
		&b.BlockNumber, &b.GasLimit, &b.GasUsed, &b.Timestamp, &b.ExtraData, &b.BaseFeePerGas, &b.BlockHash, 
		spec.Wrap(&b.Transactions), spec.Wrap(&b.Withdrawals), &b.BlobGasUsed, &b.ExcessBlobGas,
		spec.Wrap(&b.DepositRequests), spec.Wrap(&b.WithdrawalRequests), spec.Wrap(&b.ConsolidationRequests))
}

func (b *ExecutionPayload) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&b.ParentHash, &b.FeeRecipient, &b.StateRoot, &b.ReceiptsRoot, &b.LogsBloom, &b.PrevRandao,
		&b.BlockNumber, &b.GasLimit, &b.GasUsed, &b.Timestamp, &b.ExtraData, &b.BaseFeePerGas, &b.BlockHash, 
		spec.Wrap(&b.Transactions), spec.Wrap(&b.Withdrawals), &b.BlobGasUsed, &b.ExcessBlobGas,
		spec.Wrap(&b.DepositRequests), spec.Wrap(&b.WithdrawalRequests), spec.Wrap(&b.ConsolidationRequests))
}

func (b *ExecutionPayload) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&b.ParentHash, &b.FeeRecipient, &b.StateRoot, &b.ReceiptsRoot, &b.LogsBloom, &b.PrevRandao,
		&b.BlockNumber, &b.GasLimit, &b.GasUsed, &b.Timestamp, &b.ExtraData, &b.BaseFeePerGas, &b.BlockHash, 
		spec.Wrap(&b.Transactions), spec.Wrap(&b.Withdrawals), &b.BlobGasUsed, &b.ExcessBlobGas,
		spec.Wrap(&b.DepositRequests), spec.Wrap(&b.WithdrawalRequests), spec.Wrap(&b.ConsolidationRequests))
}

func (b *ExecutionPayload) FixedLength(*common.Spec) uint64 {
	return 0
}

func (b *ExecutionPayload) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&b.ParentHash, &b.FeeRecipient, &b.StateRoot, &b.ReceiptsRoot, &b.LogsBloom, &b.PrevRandao,
		&b.BlockNumber, &b.GasLimit, &b.GasUsed, &b.Timestamp, &b.ExtraData, &b.BaseFeePerGas, &b.BlockHash, 
		spec.Wrap(&b.Transactions), spec.Wrap(&b.Withdrawals), &b.BlobGasUsed, &b.ExcessBlobGas,
		spec.Wrap(&b.DepositRequests), spec.Wrap(&b.WithdrawalRequests), spec.Wrap(&b.ConsolidationRequests))
}

type ExecutionPayloadHeader struct {
	ParentHash       common.Hash32      `json:"parent_hash" yaml:"parent_hash"`
	FeeRecipient     common.Eth1Address `json:"fee_recipient" yaml:"fee_recipient"`
	StateRoot        common.Root        `json:"state_root" yaml:"state_root"`
	ReceiptsRoot     common.Root        `json:"receipts_root" yaml:"receipts_root"`
	LogsBloom        common.LogsBloom   `json:"logs_bloom" yaml:"logs_bloom"`
	PrevRandao       common.Hash32      `json:"prev_randao" yaml:"prev_randao"`
	BlockNumber      Uint64View         `json:"block_number" yaml:"block_number"`
	GasLimit         Uint64View         `json:"gas_limit" yaml:"gas_limit"`
	GasUsed          Uint64View         `json:"gas_used" yaml:"gas_used"`
	Timestamp        common.Timestamp   `json:"timestamp" yaml:"timestamp"`
	ExtraDataRoot    common.Root        `json:"extra_data_root" yaml:"extra_data_root"`
	BaseFeePerGas    Uint256View        `json:"base_fee_per_gas" yaml:"base_fee_per_gas"`
	BlockHash        common.Hash32      `json:"block_hash" yaml:"block_hash"`
	TransactionsRoot common.Root        `json:"transactions_root" yaml:"transactions_root"`
	WithdrawalsRoot  common.Root        `json:"withdrawals_root" yaml:"withdrawals_root"`
	BlobGasUsed      Uint64View         `json:"blob_gas_used" yaml:"blob_gas_used"`
	ExcessBlobGas    Uint64View         `json:"excess_blob_gas" yaml:"excess_blob_gas"`
	DepositRequestsRoot     common.Root `json:"deposit_requests_root" yaml:"deposit_requests_root"`
	WithdrawalRequestsRoot  common.Root `json:"withdrawal_requests_root" yaml:"withdrawal_requests_root"`
	ConsolidationRequestsRoot common.Root `json:"consolidation_requests_root" yaml:"consolidation_requests_root"`
}

func (h *ExecutionPayloadHeader) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&h.ParentHash, &h.FeeRecipient, &h.StateRoot, &h.ReceiptsRoot, &h.LogsBloom, &h.PrevRandao,
		&h.BlockNumber, &h.GasLimit, &h.GasUsed, &h.Timestamp, &h.ExtraDataRoot, &h.BaseFeePerGas, &h.BlockHash,
		&h.TransactionsRoot, &h.WithdrawalsRoot, &h.BlobGasUsed, &h.ExcessBlobGas,
		&h.DepositRequestsRoot, &h.WithdrawalRequestsRoot, &h.ConsolidationRequestsRoot)
}

func (h *ExecutionPayloadHeader) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&h.ParentHash, &h.FeeRecipient, &h.StateRoot, &h.ReceiptsRoot, &h.LogsBloom, &h.PrevRandao,
		&h.BlockNumber, &h.GasLimit, &h.GasUsed, &h.Timestamp, &h.ExtraDataRoot, &h.BaseFeePerGas, &h.BlockHash,
		&h.TransactionsRoot, &h.WithdrawalsRoot, &h.BlobGasUsed, &h.ExcessBlobGas,
		&h.DepositRequestsRoot, &h.WithdrawalRequestsRoot, &h.ConsolidationRequestsRoot)
}

func (h *ExecutionPayloadHeader) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&h.ParentHash, &h.FeeRecipient, &h.StateRoot, &h.ReceiptsRoot, &h.LogsBloom, &h.PrevRandao,
		&h.BlockNumber, &h.GasLimit, &h.GasUsed, &h.Timestamp, &h.ExtraDataRoot, &h.BaseFeePerGas, &h.BlockHash,
		&h.TransactionsRoot, &h.WithdrawalsRoot, &h.BlobGasUsed, &h.ExcessBlobGas,
		&h.DepositRequestsRoot, &h.WithdrawalRequestsRoot, &h.ConsolidationRequestsRoot)
}

func (h *ExecutionPayloadHeader) FixedLength(*common.Spec) uint64 {
	return 0
}

func (h *ExecutionPayloadHeader) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&h.ParentHash, &h.FeeRecipient, &h.StateRoot, &h.ReceiptsRoot, &h.LogsBloom, &h.PrevRandao,
		&h.BlockNumber, &h.GasLimit, &h.GasUsed, &h.Timestamp, &h.ExtraDataRoot, &h.BaseFeePerGas, &h.BlockHash,
		&h.TransactionsRoot, &h.WithdrawalsRoot, &h.BlobGasUsed, &h.ExcessBlobGas,
		&h.DepositRequestsRoot, &h.WithdrawalRequestsRoot, &h.ConsolidationRequestsRoot)
}

type ExecutionPayloadHeaderView struct{ *ContainerView }

func AsExecutionPayloadHeader(v View, err error) (*ExecutionPayloadHeaderView, error) {
	c, err := AsContainer(v, err)
	return &ExecutionPayloadHeaderView{c}, err
}

var ExecutionPayloadHeaderType = ContainerType("ExecutionPayloadHeader", []FieldDef{
	{"parent_hash", common.Hash32Type},
	{"fee_recipient", common.Eth1AddressType},
	{"state_root", RootType},
	{"receipts_root", RootType},
	{"logs_bloom", common.LogsBloomType},
	{"prev_randao", common.Hash32Type},
	{"block_number", Uint64Type},
	{"gas_limit", Uint64Type},
	{"gas_used", Uint64Type},
	{"timestamp", common.TimestampType},
	{"extra_data_root", RootType},
	{"base_fee_per_gas", Uint256Type},
	{"block_hash", common.Hash32Type},
	{"transactions_root", RootType},
	{"withdrawals_root", RootType},
	{"blob_gas_used", Uint64Type},
	{"excess_blob_gas", Uint64Type},
	{"deposit_requests_root", RootType},
	{"withdrawal_requests_root", RootType},
	{"consolidation_requests_root", RootType},
})