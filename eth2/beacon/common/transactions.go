package common

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/conv"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

const MAX_BYTES_PER_OPAQUE_TRANSACTION = 1 << 20
const MAX_EXECUTION_TRANSACTIONS = 1 << 14

func PayloadTransactionsType(spec *Spec) ListTypeDef {
	return ListType(TransactionType(spec), MAX_EXECUTION_TRANSACTIONS)
}

type PayloadTransactions []Transaction

func (txs *PayloadTransactions) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*txs)
		*txs = append(*txs, Transaction{})
		return spec.Wrap(&((*txs)[i]))
	}, 0, MAX_EXECUTION_TRANSACTIONS)
}

func (txs PayloadTransactions) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return spec.Wrap(&txs[i])
	}, 0, uint64(len(txs)))
}

func (txs PayloadTransactions) ByteLength(spec *Spec) (out uint64) {
	for _, v := range txs {
		out += v.ByteLength(spec) + codec.OFFSET_SIZE
	}
	return
}

func (txs *PayloadTransactions) FixedLength(*Spec) uint64 {
	return 0
}

func (txs PayloadTransactions) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	length := uint64(len(txs))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(&txs[i])
		}
		return nil
	}, length, MAX_EXECUTION_TRANSACTIONS)
}

var OpaqueTransactionType = BasicListType(Uint8Type, MAX_BYTES_PER_OPAQUE_TRANSACTION)

type OpaqueTransaction []byte

func (otx *OpaqueTransaction) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.ByteList((*[]byte)(otx), MAX_BYTES_PER_OPAQUE_TRANSACTION)
}

func (otx OpaqueTransaction) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.Write(otx)
}

func (otx OpaqueTransaction) ByteLength(spec *Spec) (out uint64) {
	return uint64(len(otx))
}

func (otx *OpaqueTransaction) FixedLength(*Spec) uint64 {
	return 0
}

func (otx OpaqueTransaction) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.ByteListHTR(otx, MAX_BYTES_PER_OPAQUE_TRANSACTION)
}

func (otx OpaqueTransaction) MarshalText() ([]byte, error) {
	return conv.BytesMarshalText(otx[:])
}

func (otx OpaqueTransaction) String() string {
	return "0x" + hex.EncodeToString(otx[:])
}

func (otx *OpaqueTransaction) UnmarshalText(text []byte) error {
	if otx == nil {
		return errors.New("cannot decode into nil opaque transaction")
	}
	return conv.DynamicBytesUnmarshalText((*[]byte)(otx), text[:])
}

func (otx OpaqueTransaction) View(spec *Spec) (*OpaqueTransactionView, error) {
	dec := codec.NewDecodingReader(bytes.NewReader(otx), uint64(len(otx)))
	return AsOpaqueTransaction(OpaqueTransactionType.Deserialize(dec))
}

type OpaqueTransactionView struct {
	*UnionView
}

func AsOpaqueTransaction(v View, err error) (*OpaqueTransactionView, error) {
	c, err := AsUnion(v, err)
	return &OpaqueTransactionView{c}, err
}

// Union[OpaqueTransaction]
type Transaction struct {
	Selector uint8 `json:"selector" yaml:"selector"`
	// *OpaqueTransaction, and future different types
	Value interface{} `json:"value" yaml:"value"`
}

func (h *Transaction) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.Union(func(selector uint8) (codec.Deserializable, error) {
		h.Selector = selector
		switch selector {
		case 0:
			dat := new(OpaqueTransaction)
			h.Value = dat
			return spec.Wrap(dat), nil
		default:
			return nil, errors.New("bad selector value")
		}
	})
}

func (h *Transaction) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	switch h.Selector {
	case 0:
		otx, ok := h.Value.(*OpaqueTransaction)
		if !ok {
			return fmt.Errorf("invalid value type for 0 (opaque transaction) selector: %T", h.Value)
		}
		return w.Union(0, spec.Wrap(otx))
	default:
		return errors.New("bad selector value")
	}
}

func (h *Transaction) ByteLength(spec *Spec) uint64 {
	switch h.Selector {
	case 0:
		otx, ok := h.Value.(*OpaqueTransaction)
		if !ok {
			panic(fmt.Errorf("invalid value type for 0 (opaque transaction) selector: %T", h.Value))
		}
		return 1 + otx.ByteLength(spec)
	default:
		panic(errors.New("bad selector value"))
	}
}

func (h *Transaction) FixedLength(spec *Spec) uint64 {
	return 0
}

func (h *Transaction) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	switch h.Selector {
	case 0:
		if h.Value == nil {
			return hFn.Union(h.Selector, spec.Wrap(&OpaqueTransaction{}))
		}
		otx, ok := h.Value.(*OpaqueTransaction)
		if !ok {
			panic(fmt.Errorf("invalid value type for 0 (opaque transaction) selector: %T", h.Value))
		}
		return hFn.Union(h.Selector, spec.Wrap(otx))
	default:
		panic(errors.New("bad selector value"))
	}
}

func (h *Transaction) View(spec *Spec) (*TransactionView, error) {
	switch h.Selector {
	case 0:
		otx, ok := h.Value.(*OpaqueTransaction)
		if !ok {
			return nil, fmt.Errorf("invalid value type for 0 (opaque transaction) selector: %T", h.Value)
		}
		otxView, err := otx.View(spec)
		if err != nil {
			return nil, err
		}
		return AsTransaction(TransactionType(spec).FromView(0, otxView))
	default:
		return nil, errors.New("bad selector value")
	}
}

type TransactionView struct {
	*UnionView
}

func TransactionType(spec *Spec) *UnionTypeDef {
	return UnionType([]TypeDef{OpaqueTransactionType})
}

func AsTransaction(v View, err error) (*TransactionView, error) {
	c, err := AsUnion(v, err)
	return &TransactionView{c}, err
}
