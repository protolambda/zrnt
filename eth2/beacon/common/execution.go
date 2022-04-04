package common

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/conv"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Hash32 = Root

const Hash32Type = RootType

var ExecutionPayloadHeaderType = ContainerType("ExecutionPayloadHeader", []FieldDef{
	{"parent_hash", Hash32Type},
	{"fee_recipient", Eth1AddressType},
	{"state_root", Bytes32Type},
	{"receipts_root", Bytes32Type},
	{"logs_bloom", LogsBloomType},
	{"prev_randao", Bytes32Type},
	{"block_number", Uint64Type},
	{"gas_limit", Uint64Type},
	{"gas_used", Uint64Type},
	{"timestamp", TimestampType},
	{"extra_data", ExtraDataType},
	{"base_fee_per_gas", Uint256Type},
	{"block_hash", Hash32Type},
	{"transactions_root", RootType},
})

type ExecutionPayloadHeaderView struct {
	*ContainerView
}

func (v *ExecutionPayloadHeaderView) Raw() (*ExecutionPayloadHeader, error) {
	values, err := v.FieldValues()
	if err != nil {
		return nil, err
	}
	if len(values) != 14 {
		return nil, fmt.Errorf("unexpected number of execution payload header fields: %d", len(values))
	}
	parentHash, err := AsRoot(values[0], err)
	feeRecipient, err := AsEth1Address(values[1], err)
	stateRoot, err := AsRoot(values[2], err)
	receiptsRoot, err := AsRoot(values[3], err)
	logsBloomView, err := AsLogsBloom(values[4], err)
	prevRandao, err := AsRoot(values[5], err)
	blockNumber, err := AsUint64(values[6], err)
	gasLimit, err := AsUint64(values[7], err)
	gasUsed, err := AsUint64(values[8], err)
	timestamp, err := AsTimestamp(values[9], err)
	extraDataView, err := AsExtraData(values[10], err)
	baseFeePerGas, err := AsUint256(values[11], err)
	blockHash, err := AsRoot(values[12], err)
	transactionsRoot, err := AsRoot(values[13], err)
	if err != nil {
		return nil, err
	}
	logsBloom, err := logsBloomView.Raw()
	if err != nil {
		return nil, err
	}
	extraData, err := extraDataView.Raw()
	if err != nil {
		return nil, err
	}
	return &ExecutionPayloadHeader{
		ParentHash:       parentHash,
		FeeRecipient:     feeRecipient,
		StateRoot:        stateRoot,
		ReceiptsRoot:     receiptsRoot,
		LogsBloom:        *logsBloom,
		PrevRandao:       prevRandao,
		BlockNumber:      blockNumber,
		GasLimit:         gasLimit,
		GasUsed:          gasUsed,
		Timestamp:        timestamp,
		ExtraData:        extraData,
		BaseFeePerGas:    baseFeePerGas,
		BlockHash:        blockHash,
		TransactionsRoot: transactionsRoot,
	}, nil
}

func (v *ExecutionPayloadHeaderView) ParentHash() (Hash32, error) {
	return AsRoot(v.Get(0))
}

func (v *ExecutionPayloadHeaderView) FeeRecipient() (Eth1Address, error) {
	return AsEth1Address(v.Get(1))
}

func (v *ExecutionPayloadHeaderView) StateRoot() (Bytes32, error) {
	return AsRoot(v.Get(2))
}

func (v *ExecutionPayloadHeaderView) ReceiptRoot() (Bytes32, error) {
	return AsRoot(v.Get(3))
}

func (v *ExecutionPayloadHeaderView) LogsBloom() (*LogsBloom, error) {
	logV, err := AsLogsBloom(v.Get(4))
	if err != nil {
		return nil, err
	}
	return logV.Raw()
}

func (v *ExecutionPayloadHeaderView) Random() (Bytes32, error) {
	return AsRoot(v.Get(5))
}

func (v *ExecutionPayloadHeaderView) BlockNumber() (Uint64View, error) {
	return AsUint64(v.Get(6))
}

func (v *ExecutionPayloadHeaderView) GasLimit() (Uint64View, error) {
	return AsUint64(v.Get(7))
}

func (v *ExecutionPayloadHeaderView) GasUsed() (Uint64View, error) {
	return AsUint64(v.Get(8))
}

func (v *ExecutionPayloadHeaderView) Timestamp() (Timestamp, error) {
	return AsTimestamp(v.Get(9))
}

func (v *ExecutionPayloadHeaderView) BaseFeePerGas() (Uint256View, error) {
	return AsUint256(v.Get(10))
}

func (v *ExecutionPayloadHeaderView) BlockHash() (Hash32, error) {
	return AsRoot(v.Get(11))
}

func (v *ExecutionPayloadHeaderView) TransactionsRoot() (Root, error) {
	return AsRoot(v.Get(12))
}

func AsExecutionPayloadHeader(v View, err error) (*ExecutionPayloadHeaderView, error) {
	c, err := AsContainer(v, err)
	return &ExecutionPayloadHeaderView{c}, err
}

type ExecutionPayloadHeader struct {
	ParentHash       Hash32      `json:"parent_hash" yaml:"parent_hash"`
	FeeRecipient     Eth1Address `json:"fee_recipient" yaml:"fee_recipient"`
	StateRoot        Bytes32     `json:"state_root" yaml:"state_root"`
	ReceiptsRoot     Bytes32     `json:"receipts_root" yaml:"receipts_root"`
	LogsBloom        LogsBloom   `json:"logs_bloom" yaml:"logs_bloom"`
	PrevRandao       Bytes32     `json:"prev_randao" yaml:"prev_randao"`
	BlockNumber      Uint64View  `json:"block_number" yaml:"block_number"`
	GasLimit         Uint64View  `json:"gas_limit" yaml:"gas_limit"`
	GasUsed          Uint64View  `json:"gas_used" yaml:"gas_used"`
	Timestamp        Timestamp   `json:"timestamp" yaml:"timestamp"`
	ExtraData        ExtraData   `json:"extra_data" yaml:"extra_data"`
	BaseFeePerGas    Uint256View `json:"base_fee_per_gas" yaml:"base_fee_per_gas"`
	BlockHash        Hash32      `json:"block_hash" yaml:"block_hash"`
	TransactionsRoot Root        `json:"transactions_root" yaml:"transactions_root"`
}

func (s *ExecutionPayloadHeader) View() *ExecutionPayloadHeaderView {
	ed, err := s.ExtraData.View()
	if err != nil {
		panic(err)
	}
	pr, cb, sr, rr := (*RootView)(&s.ParentHash), s.FeeRecipient.View(), (*RootView)(&s.StateRoot), (*RootView)(&s.ReceiptsRoot)
	lb, rng, nr, gl, gu := s.LogsBloom.View(), (*RootView)(&s.PrevRandao), s.BlockNumber, s.GasLimit, s.GasUsed
	ts, bf, bh, tr := Uint64View(s.Timestamp), &s.BaseFeePerGas, (*RootView)(&s.BlockHash), (*RootView)(&s.TransactionsRoot)

	v, err := AsExecutionPayloadHeader(ExecutionPayloadHeaderType.FromFields(pr, cb, sr, rr, lb, rng, nr, gl, gu, ts, ed, bf, bh, tr))
	if err != nil {
		panic(err)
	}
	return v
}

func (s *ExecutionPayloadHeader) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&s.ParentHash, &s.FeeRecipient, &s.StateRoot,
		&s.ReceiptsRoot, &s.LogsBloom, &s.PrevRandao, &s.BlockNumber, &s.GasLimit,
		&s.GasUsed, &s.Timestamp, &s.ExtraData, &s.BaseFeePerGas, &s.BlockHash, &s.TransactionsRoot)
}

func (s *ExecutionPayloadHeader) Serialize(w *codec.EncodingWriter) error {
	return w.Container(&s.ParentHash, &s.FeeRecipient, &s.StateRoot,
		&s.ReceiptsRoot, &s.LogsBloom, &s.PrevRandao, &s.BlockNumber, &s.GasLimit,
		&s.GasUsed, &s.Timestamp, &s.ExtraData, &s.BaseFeePerGas, &s.BlockHash, &s.TransactionsRoot)
}

func (s *ExecutionPayloadHeader) ByteLength() uint64 {
	return codec.ContainerLength(&s.ParentHash, &s.FeeRecipient, &s.StateRoot,
		&s.ReceiptsRoot, &s.LogsBloom, &s.PrevRandao, &s.BlockNumber, &s.GasLimit,
		&s.GasUsed, &s.Timestamp, &s.ExtraData, &s.BaseFeePerGas, &s.BlockHash, &s.TransactionsRoot)
}

func (b *ExecutionPayloadHeader) FixedLength() uint64 {
	return 0
}

func (s *ExecutionPayloadHeader) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(&s.ParentHash, &s.FeeRecipient, &s.StateRoot,
		&s.ReceiptsRoot, &s.LogsBloom, &s.PrevRandao, &s.BlockNumber, &s.GasLimit,
		&s.GasUsed, &s.Timestamp, &s.ExtraData, &s.BaseFeePerGas, &s.BlockHash, &s.TransactionsRoot)
}

func ExecutionPayloadType(spec *Spec) *ContainerTypeDef {
	return ContainerType("ExecutionPayload", []FieldDef{
		{"parent_hash", Hash32Type},
		{"fee_recipient", Eth1AddressType},
		{"state_root", Bytes32Type},
		{"receipts_root", Bytes32Type},
		{"logs_bloom", LogsBloomType},
		{"prev_randao", Bytes32Type},
		{"block_number", Uint64Type},
		{"gas_limit", Uint64Type},
		{"gas_used", Uint64Type},
		{"timestamp", TimestampType},
		{"extra_data", ExtraDataType},
		{"base_fee_per_gas", Uint256Type},
		{"block_hash", Hash32Type},
		{"transactions", PayloadTransactionsType(spec)},
	})
}

type ExecutionPayloadView struct {
	*ContainerView
}

func AsExecutionPayload(v View, err error) (*ExecutionPayloadView, error) {
	c, err := AsContainer(v, err)
	return &ExecutionPayloadView{c}, err
}

const MAX_EXTRA_DATA_BYTES = 32

var ExtraDataType = BasicListType(Uint8Type, MAX_EXTRA_DATA_BYTES)

type ExtraData []byte

func (otx *ExtraData) Deserialize(dr *codec.DecodingReader) error {
	return dr.ByteList((*[]byte)(otx), MAX_EXTRA_DATA_BYTES)
}

func (otx ExtraData) Serialize(w *codec.EncodingWriter) error {
	return w.Write(otx)
}

func (otx ExtraData) ByteLength() (out uint64) {
	return uint64(len(otx))
}

func (otx *ExtraData) FixedLength() uint64 {
	return 0
}

func (otx ExtraData) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.ByteListHTR(otx, MAX_EXTRA_DATA_BYTES)
}

func (otx ExtraData) MarshalText() ([]byte, error) {
	return conv.BytesMarshalText(otx[:])
}

func (otx ExtraData) String() string {
	return "0x" + hex.EncodeToString(otx[:])
}

func (otx *ExtraData) UnmarshalText(text []byte) error {
	if otx == nil {
		return errors.New("cannot decode into nil opaque transaction")
	}
	return conv.DynamicBytesUnmarshalText((*[]byte)(otx), text[:])
}

func (otx ExtraData) View() (*ExtraDataView, error) {
	dec := codec.NewDecodingReader(bytes.NewReader(otx), uint64(len(otx)))
	return AsExtraData(ExtraDataType.Deserialize(dec))
}

type ExtraDataView struct {
	*BasicListView
}

func AsExtraData(v View, err error) (*ExtraDataView, error) {
	c, err := AsBasicList(v, err)
	return &ExtraDataView{c}, err
}
func (v *ExtraDataView) Raw() (ExtraData, error) {
	var buf bytes.Buffer
	w := codec.NewEncodingWriter(&buf)
	if err := v.Serialize(w); err != nil {
		return nil, err
	}
	return ExtraData(buf.Bytes()), nil
}

type ExecutionPayload struct {
	ParentHash    Hash32              `json:"parent_hash" yaml:"parent_hash"`
	FeeRecipient  Eth1Address         `json:"fee_recipient" yaml:"fee_recipient"`
	StateRoot     Bytes32             `json:"state_root" yaml:"state_root"`
	ReceiptsRoot  Bytes32             `json:"receipts_root" yaml:"receipts_root"`
	LogsBloom     LogsBloom           `json:"logs_bloom" yaml:"logs_bloom"`
	PrevRandao    Bytes32             `json:"prev_randao" yaml:"prev_randao"`
	BlockNumber   Uint64View          `json:"block_number" yaml:"block_number"`
	GasLimit      Uint64View          `json:"gas_limit" yaml:"gas_limit"`
	GasUsed       Uint64View          `json:"gas_used" yaml:"gas_used"`
	Timestamp     Timestamp           `json:"timestamp" yaml:"timestamp"`
	ExtraData     ExtraData           `json:"extra_data" yaml:"extra_data"`
	BaseFeePerGas Uint256View         `json:"base_fee_per_gas" yaml:"base_fee_per_gas"`
	BlockHash     Hash32              `json:"block_hash" yaml:"block_hash"`
	Transactions  PayloadTransactions `json:"transactions" yaml:"transactions"`
}

func (s *ExecutionPayload) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.Container(&s.ParentHash, &s.FeeRecipient, &s.StateRoot,
		&s.ReceiptsRoot, &s.LogsBloom, &s.PrevRandao, &s.BlockNumber, &s.GasLimit,
		&s.GasUsed, &s.Timestamp, &s.ExtraData, &s.BaseFeePerGas, &s.BlockHash, spec.Wrap(&s.Transactions))
}

func (s *ExecutionPayload) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.Container(&s.ParentHash, &s.FeeRecipient, &s.StateRoot,
		&s.ReceiptsRoot, &s.LogsBloom, &s.PrevRandao, &s.BlockNumber, &s.GasLimit,
		&s.GasUsed, &s.Timestamp, &s.ExtraData, &s.BaseFeePerGas, &s.BlockHash, spec.Wrap(&s.Transactions))
}

func (s *ExecutionPayload) ByteLength(spec *Spec) uint64 {
	return codec.ContainerLength(&s.ParentHash, &s.FeeRecipient, &s.StateRoot,
		&s.ReceiptsRoot, &s.LogsBloom, &s.PrevRandao, &s.BlockNumber, &s.GasLimit,
		&s.GasUsed, &s.Timestamp, &s.ExtraData, &s.BaseFeePerGas, &s.BlockHash, spec.Wrap(&s.Transactions))
}

func (a *ExecutionPayload) FixedLength(*Spec) uint64 {
	// transactions list is not fixed length, so the whole thing is not fixed length.
	return 0
}

func (s *ExecutionPayload) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(&s.ParentHash, &s.FeeRecipient, &s.StateRoot,
		&s.ReceiptsRoot, &s.LogsBloom, &s.PrevRandao, &s.BlockNumber, &s.GasLimit,
		&s.GasUsed, &s.Timestamp, &s.ExtraData, &s.BaseFeePerGas, &s.BlockHash, spec.Wrap(&s.Transactions))
}

func (ep *ExecutionPayload) Header(spec *Spec) *ExecutionPayloadHeader {
	return &ExecutionPayloadHeader{
		ParentHash:       ep.ParentHash,
		FeeRecipient:     ep.FeeRecipient,
		StateRoot:        ep.StateRoot,
		ReceiptsRoot:     ep.ReceiptsRoot,
		LogsBloom:        ep.LogsBloom,
		PrevRandao:       ep.PrevRandao,
		BlockNumber:      ep.BlockNumber,
		GasLimit:         ep.GasLimit,
		GasUsed:          ep.GasUsed,
		Timestamp:        ep.Timestamp,
		ExtraData:        ep.ExtraData,
		BaseFeePerGas:    ep.BaseFeePerGas,
		BlockHash:        ep.BlockHash,
		TransactionsRoot: ep.Transactions.HashTreeRoot(spec, tree.GetHashFn()),
	}
}

type ExecutionEngine interface {
	ExecutePayload(ctx context.Context, executionPayload *ExecutionPayload) (valid bool, err error)
	// TODO: remaining interface parts
}
