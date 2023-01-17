package deneb

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"github.com/protolambda/ztyp/view"

	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type KZGCommitments []common.KZGCommitment

func (li *KZGCommitments) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*li)
		*li = append(*li, common.KZGCommitment{})
		return &((*li)[i])
	}, common.KZGCommitmentSize, uint64(spec.MAX_BLOBS_PER_BLOCK))
}

func (li KZGCommitments) Serialize(_ *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &li[i]
	}, common.KZGCommitmentSize, uint64(len(li)))
}

func (li KZGCommitments) ByteLength(_ *common.Spec) (out uint64) {
	return common.KZGCommitmentSize * uint64(len(li))
}

func (*KZGCommitments) FixedLength(*common.Spec) uint64 {
	return 0
}

func (li KZGCommitments) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, uint64(spec.MAX_BLOBS_PER_BLOCK))
}

func KZGCommitmentsType(spec *common.Spec) *view.ComplexListTypeDef {
	return view.ComplexListType(common.KZGCommitmentType, uint64(spec.MAX_BLOBS_PER_BLOCK))
}

type TransactionsAndBlobCommitments interface {
	GetTransactions() []common.Transaction
	GetBlobKZGCommitments() []common.KZGCommitment
}

const (
	blobTxMessageOffsetLen              = 4
	blobTxSignatureLen                  = 1 + 32 + 32
	blobTxMessageOffset                 = blobTxMessageOffsetLen + blobTxSignatureLen
	blobTxMessageLenTillVersionedHashes = 32 + 8 + 32 + 32 + 8 + 4 + 32 + 4 + 4 + 32
	blobTxVersionedHashesStart          = 1 + blobTxMessageOffset + blobTxMessageLenTillVersionedHashes
)

func TxPeekBlobVersionedHashes(opaqueTx common.Transaction) ([]common.Hash32, error) {
	if len(opaqueTx) < blobTxVersionedHashesStart+4 {
		return nil, fmt.Errorf("blob tx is too small: %d, expected at least %d bytes to read versioned hashes", len(opaqueTx), blobTxVersionedHashesStart+4)
	}
	if opaqueTx[0] != common.BLOB_TX_TYPE {
		return nil, fmt.Errorf("tx is not a blob tx type: %d", opaqueTx[0])
	}
	if messageOffset := binary.LittleEndian.Uint32(opaqueTx[1:5]); messageOffset != blobTxMessageOffset {
		return nil, fmt.Errorf("blob tx has invalid message offset: %d, expected %d", messageOffset, blobTxMessageOffset)
	}
	blobVersionedHashesOffset := 1 + blobTxMessageOffset + uint64(binary.LittleEndian.Uint32(opaqueTx[blobTxVersionedHashesStart:blobTxVersionedHashesStart+4]))
	if blobVersionedHashesOffset > uint64(len(opaqueTx)) || (uint64(len(opaqueTx))-blobVersionedHashesOffset)%32 != 0 {
		return nil, fmt.Errorf("versioned hashes start at byte %d, but have %d bytes", blobVersionedHashesOffset, uint64(len(opaqueTx)))
	}
	out := make([]common.Hash32, (uint64(len(opaqueTx))-blobVersionedHashesOffset)/32)
	for i := range out {
		x := blobVersionedHashesOffset + uint64(i*32)
		copy(out[i][:], opaqueTx[x:x+32])
	}
	return out, nil
}

func ProcessBlobKZGCommitments(ctx context.Context, spec *common.Spec, state *BeaconStateView, txsAndBlobs TransactionsAndBlobCommitments) error {
	var allVersionedHashes []common.Hash32
	for i, tx := range txsAndBlobs.GetTransactions() {
		if len(tx) > 0 && tx[0] == common.BLOB_TX_TYPE {
			txVersionedHashes, err := TxPeekBlobVersionedHashes(tx)
			if err != nil {
				return fmt.Errorf("failed to peek into tx %d of block for versioned hashes: %w", i, err)
			}
			allVersionedHashes = append(allVersionedHashes, txVersionedHashes...)
		}
	}
	kzgCommitments := txsAndBlobs.GetBlobKZGCommitments()
	if len(allVersionedHashes) != len(kzgCommitments) {
		return fmt.Errorf("got %d versioned hashes, but have %d kzg commitments", len(allVersionedHashes), len(kzgCommitments))
	}
	for i, commitment := range kzgCommitments {
		if x := commitment.ToVersionedHash(); x != allVersionedHashes[i] {
			return fmt.Errorf("entry %d does not match: versioned hash %s does not match hash %s computed from commitment %s", i, allVersionedHashes[i], x, commitment)
		}
	}
	return nil
}
