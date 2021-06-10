package sharding

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

var ShardBlobHeaderType = ContainerType("ShardBlobHeader", []FieldDef{
	{"slot", common.SlotType},
	{"shard", common.ShardType},
	{"body_summary", ShardBlobBodySummaryType},
	{"proposer_index", common.ValidatorIndexType},
})

type ShardBlobHeader struct {
	// Slot that this header is intended for
	Slot common.Slot `json:"slot" yaml:"slot"`
	// Shard that this header is intended for
	Shard common.Shard `json:"shard" yaml:"shard"`

	// SSZ-summary of ShardBlobBody
	BodySummary ShardBlobBodySummary `json:"body_summary" yaml:"body_summary"`

	// Proposer of the shard-blob
	ProposerIndex common.ValidatorIndex `json:"proposer_index" yaml:"proposer_index"`
}

func (v *ShardBlobHeader) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&v.Slot, &v.Shard, &v.BodySummary, &v.ProposerIndex)
}

func (v *ShardBlobHeader) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&v.Slot, &v.Shard, &v.BodySummary, &v.ProposerIndex)
}

func (v *ShardBlobHeader) ByteLength() uint64 {
	return ShardBlobHeaderType.TypeByteLength()
}

func (*ShardBlobHeader) FixedLength() uint64 {
	return ShardBlobHeaderType.TypeByteLength()
}

func (v *ShardBlobHeader) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&v.Slot, &v.Shard, &v.BodySummary, &v.ProposerIndex)
}

var SignedShardBlobHeaderType = ContainerType("SignedShardBlobHeader", []FieldDef{
	{"message", ShardBlobHeaderType},
	{"signature", common.BLSSignatureType},
})

type SignedShardBlobHeader struct {
	Message   ShardBlobHeader     `json:"message" yaml:"message"`
	Signature common.BLSSignature `json:"signature" yaml:"signature"`
}

func (v *SignedShardBlobHeader) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&v.Message, &v.Signature)
}

func (v *SignedShardBlobHeader) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&v.Message, &v.Signature)
}

func (v *SignedShardBlobHeader) ByteLength() uint64 {
	return SignedShardBlobHeaderType.TypeByteLength()
}

func (*SignedShardBlobHeader) FixedLength() uint64 {
	return SignedShardBlobHeaderType.TypeByteLength()
}

func (v *SignedShardBlobHeader) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&v.Message, v.Signature)
}

func BlockShardHeadersType(spec *common.Spec) ListTypeDef {
	return ListType(SignedShardBlobHeaderType, spec.MAX_SHARDS*spec.MAX_SHARD_HEADERS_PER_SHARD)
}

type ShardHeaders []SignedShardBlobHeader

func (a *ShardHeaders) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, SignedShardBlobHeader{})
		return &((*a)[i])
	}, SignedShardBlobHeaderType.TypeByteLength(), spec.MAX_SHARDS*spec.MAX_SHARD_HEADERS_PER_SHARD)
}

func (a ShardHeaders) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, SignedShardBlobHeaderType.TypeByteLength(), uint64(len(a)))
}

func (a ShardHeaders) ByteLength(*common.Spec) (out uint64) {
	return SignedShardBlobHeaderType.TypeByteLength() * uint64(len(a))
}

func (a *ShardHeaders) FixedLength(*common.Spec) uint64 {
	return 0
}

func (li ShardHeaders) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, spec.MAX_SHARDS*spec.MAX_SHARD_HEADERS_PER_SHARD)
}

func ProcessShardHeaders(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, ops []SignedShardBlobHeader) error {
	for i := range ops {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := ProcessShardHeader(spec, epc, state, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

func ProcessShardHeader(spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, signedHeader *SignedShardBlobHeader) error {
	header := &signedHeader.Message
	// Verify the header is not 0, and not from the future.
	if header.Slot == 0 {
		return errors.New("shard blob header slot must be non-zero")
	}
	slot, err := state.Slot()
	if err != nil {
		return err
	}
	if header.Slot > slot {
		return fmt.Errorf("shard blob header slot must not be from the future, got %d, expected <= %d", header.Slot, slot)
	}
	headerEpoch := spec.SlotToEpoch(header.Slot)
	currentEpoch := spec.SlotToEpoch(slot)
	previousEpoch := currentEpoch.Previous()
	// Verify that the header is within the processing time window
	if headerEpoch != currentEpoch && headerEpoch != previousEpoch {
		return fmt.Errorf("expected shard blob header to be of current (%d) or previous (%d) epoch, but got %d", currentEpoch, previousEpoch, headerEpoch)
	}
	// Verify that the shard is active
	activeShardCount := spec.ActiveShardCount(headerEpoch)
	if uint64(header.Shard) >= activeShardCount {
		return fmt.Errorf("shard blob header shard field is out of bounds: %d, shard count is %d", header.Shard, activeShardCount)
	}
	// Verify that the block root matches,
	// to ensure the header will only be included in this specific Beacon Chain sub-tree.
	blockRoot, err := common.GetBlockRootAtSlot(spec, state, header.Slot-1)
	if err != nil {
		return err
	}
	if header.BodySummary.BeaconBlockRoot != blockRoot {
		return fmt.Errorf("shard blob header anchored at %s beacon block root, but expected %s", header.BodySummary.BeaconBlockRoot, blockRoot)
	}
	// Check that this data is still pending
	buffer, err := state.ShardBuffer()
	if err != nil {
		return err
	}
	column, err := buffer.Column(uint64(header.Slot % spec.SHARD_STATE_MEMORY_SLOTS))
	if err != nil {
		return err
	}
	committeeWork, err := column.GetWork(header.Shard)
	if err != nil {
		return err
	}
	workStatus, err := committeeWork.Status()
	if err != nil {
		return err
	}
	selector, err := workStatus.Selector()
	if err != nil {
		return err
	}
	if selector != SHARD_WORK_PENDING {
		return fmt.Errorf("shard blob header at slot %d shard %d is not pending", header.Slot, header.Shard)
	}
	currentHeaders, err := AsPendingShardHeaders(workStatus.Value())
	if err != nil {
		return err
	}
	headerRoot := header.HashTreeRoot(tree.GetHashFn())
	iter := currentHeaders.Iter()
	for {
		elem, ok, err := iter.Next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		pendingHeader, err := AsPendingShardHeader(elem, nil)
		if err != nil {
			return err
		}
		pendingHeaderRoot, err := pendingHeader.Root()
		if err != nil {
			return err
		}
		if pendingHeaderRoot == headerRoot {
			return fmt.Errorf("shard blob header cannot be added to pending list, header is already present")
		}
	}

	// Verify proposer
	expectedProposer, err := epc.GetShardProposer(header.Slot, header.Shard)
	if err != nil {
		return err
	}
	if header.ProposerIndex != expectedProposer {
		return fmt.Errorf("shard blob header proposer should be %d, but got %d", expectedProposer, header.ProposerIndex)
	}

	// Verify signature
	dom, err := common.GetDomain(state, common.DOMAIN_SHARD_PROPOSER, headerEpoch)
	if err != nil {
		return err
	}
	pubkey, ok := epc.PubkeyCache.Pubkey(header.ProposerIndex)
	if !ok {
		return fmt.Errorf("could not find pubkey of shard blob proposer %d", header.ProposerIndex)
	}
	signingRoot := common.ComputeSigningRoot(header.HashTreeRoot(tree.GetHashFn()), dom)
	if !bls.Verify(pubkey, signingRoot, signedHeader.Signature) {
		return errors.New("shard blob header has invalid signature")
	}

	// TODO
	//# Verify the length by verifying the degree.
	//body_summary = header.body_summary
	//if body_summary.commitment.length == 0:
	//assert body_summary.degree_proof == G1_SETUP[0]
	//assert (
	//	bls.Pairing(body_summary.degree_proof, G2_SETUP[0])
	//  == bls.Pairing(body_summary.commitment.point, G2_SETUP[-body_summary.commitment.length])
	//)

	index, err := ComputeCommitteeIndexFromShard(spec, epc, header.Slot, header.Shard)
	if err != nil {
		return err
	}
	committee, err := epc.GetBeaconCommittee(header.Slot, index)
	if err != nil {
		return err
	}
	// empty bitlist, packed in bytes, with delimiter bit
	emptyBits := make(phase0.AttestationBits, (len(committee)/8)+1)
	emptyBits[len(emptyBits)-1] = 1 << (uint8(len(committee)) & 7)

	pendingHeader := PendingShardHeader{
		Commitment: header.BodySummary.Commitment,
		Root:       common.Root{},
		Votes:      emptyBits,
		Weight:     0,
		UpdateSlot: header.Slot,
	}
	if err := currentHeaders.Append(pendingHeader.View(spec)); err != nil {
		return err
	}

	return nil
}

func ComputeCommitteeIndexFromShard(spec *common.Spec, epc *common.EpochsContext, slot common.Slot, shard common.Shard) (common.CommitteeIndex, error) {
	activeShards := spec.ActiveShardCount(epc.CurrentEpoch.Epoch)
	startShard, err := epc.StartShard(slot)
	if err != nil {
		return 0, err
	}
	index := common.CommitteeIndex((common.Shard(activeShards) + shard - startShard) % common.Shard(activeShards))
	committeeCount, err := epc.GetCommitteeCountPerSlot(epc.CurrentEpoch.Epoch)
	if err != nil {
		return 0, err
	}
	if index >= common.CommitteeIndex(committeeCount) {
		return 0, fmt.Errorf("slot %d shard %d pair does not have a valid committee index, got %d, count is %d", slot, shard, index, committeeCount)
	}
	return index, nil
}
