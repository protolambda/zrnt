package merge

import (
	"encoding/hex"
	"errors"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/conv"
	"github.com/protolambda/ztyp/tree"
	"github.com/protolambda/ztyp/view"
)

const MAX_BYTES_PER_OPAQUE_TRANSACTION = 1 << 20
const MAX_APPLICATION_TRANSACTIONS = 1 << 14

var PayloadTransactionsType = view.ListType(OpaqueTransactionType, MAX_APPLICATION_TRANSACTIONS)

type PayloadTransactions []OpaqueTransaction

func (txs *PayloadTransactions) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*txs)
		*txs = append(*txs, OpaqueTransaction{})
		return spec.Wrap(&((*txs)[i]))
	}, 0, MAX_APPLICATION_TRANSACTIONS)
}

func (txs PayloadTransactions) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return spec.Wrap(&txs[i])
	}, 0, uint64(len(txs)))
}

func (txs PayloadTransactions) ByteLength(spec *common.Spec) (out uint64) {
	for _, v := range txs {
		out += v.ByteLength(spec) + codec.OFFSET_SIZE
	}
	return
}

func (txs *PayloadTransactions) FixedLength(*common.Spec) uint64 {
	return 0
}

func (txs PayloadTransactions) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(txs))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(&txs[i])
		}
		return nil
	}, length, MAX_APPLICATION_TRANSACTIONS)
}

var OpaqueTransactionType = view.BasicListType(view.Uint8Type, MAX_BYTES_PER_OPAQUE_TRANSACTION)

type OpaqueTransaction []byte

func (otx *OpaqueTransaction) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.ByteList((*[]byte)(otx), MAX_BYTES_PER_OPAQUE_TRANSACTION)
}

func (otx OpaqueTransaction) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Write(otx)
}

func (otx OpaqueTransaction) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(len(otx))
}

func (otx *OpaqueTransaction) FixedLength(*common.Spec) uint64 {
	return 0
}

func (otx OpaqueTransaction) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
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
		return errors.New("cannot decode into nil transaction")
	}
	return conv.DynamicBytesUnmarshalText((*[]byte)(otx), text[:])
}
