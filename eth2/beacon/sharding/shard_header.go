package sharding

import (
	"context"
	"errors"
	"fmt"
	kbls "github.com/kilic/bls12-381"
	blsu "github.com/protolambda/bls12-381-util"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"

	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

var ShardBlobHeaderType = ContainerType("ShardBlobHeader", []FieldDef{
	{"slot", common.SlotType},
	{"shard", common.ShardType},
	{"builder_index", common.BuilderIndexType},
	{"proposer_index", common.ValidatorIndexType},
	{"body_summary", ShardBlobBodySummaryType},
})

type ShardBlobHeader struct {
	// Slot that this header is intended for
	Slot common.Slot `json:"slot" yaml:"slot"`
	// Shard that this header is intended for
	Shard common.Shard `json:"shard" yaml:"shard"`
	// Builder of the data, pays data-fee to proposer
	BuilderIndex common.BuilderIndex `json:"builder_index" yaml:"builder_index"`
	// Proposer of the shard-blob
	ProposerIndex common.ValidatorIndex `json:"proposer_index" yaml:"proposer_index"`
	// Blob contents, without the full data
	BodySummary ShardBlobBodySummary `json:"body_summary" yaml:"body_summary"`
}

func (v *ShardBlobHeader) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&v.Slot, &v.Shard, &v.BuilderIndex, &v.ProposerIndex, &v.BodySummary)
}

func (v *ShardBlobHeader) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&v.Slot, &v.Shard, &v.BuilderIndex, &v.ProposerIndex, &v.BodySummary)
}

func (v *ShardBlobHeader) ByteLength() uint64 {
	return ShardBlobHeaderType.TypeByteLength()
}

func (*ShardBlobHeader) FixedLength() uint64 {
	return ShardBlobHeaderType.TypeByteLength()
}

func (v *ShardBlobHeader) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&v.Slot, &v.Shard, &v.BuilderIndex, &v.ProposerIndex, &v.BodySummary)
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
	slot := header.Slot
	shard := header.Shard

	// Verify the header is not 0, and not from the future.
	if slot == 0 {
		return errors.New("shard blob header slot must be non-zero")
	}
	stateSlot, err := state.Slot()
	if err != nil {
		return err
	}
	if slot > stateSlot {
		return fmt.Errorf("shard blob header slot must not be from the future, got %d, expected <= %d", slot, stateSlot)
	}
	headerEpoch := spec.SlotToEpoch(slot)
	currentEpoch := spec.SlotToEpoch(stateSlot)
	previousEpoch := currentEpoch.Previous()
	// Verify that the header is within the processing time window
	if headerEpoch != currentEpoch && headerEpoch != previousEpoch {
		return fmt.Errorf("expected shard blob header to be of current (%d) or previous (%d) epoch, but got %d", currentEpoch, previousEpoch, headerEpoch)
	}
	// Verify that the shard is active
	shardCount := spec.ActiveShardCount(headerEpoch)
	if uint64(header.Shard) >= shardCount {
		return fmt.Errorf("shard blob header shard field is out of bounds: %d, shard count is %d", shard, shardCount)
	}
	// Verify that a committee is able to attest this (slot, shard)
	startShard, err := epc.StartShard(slot)
	if err != nil {
		return fmt.Errorf("failed to get start shard: %v", err)
	}
	committeeIndex := (shardCount + uint64(shard) - uint64(startShard)) % shardCount
	committeesPerSlot, err := epc.GetCommitteeCountPerSlot(headerEpoch)
	if err != nil {
		return fmt.Errorf("failed to get committees per slot: %v", err)
	}
	if committeeIndex >= committeesPerSlot {
		return fmt.Errorf("no committee active for slot %d shard %d, would be committee %d, but only have %d", slot, shard, committeeIndex, committeesPerSlot)
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

	// Check that this header is not yet in the pending list
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
		att, err := pendingHeader.Attested()
		if err != nil {
			return err
		}
		pendingHeaderRoot, err := att.Root()
		if err != nil {
			return err
		}
		if pendingHeaderRoot == headerRoot {
			return fmt.Errorf("shard blob header cannot be added to pending list, header is already present")
		}
	}

	// Verify proposer matches
	expectedProposer, err := epc.GetShardProposer(header.Slot, header.Shard)
	if err != nil {
		return err
	}
	if header.ProposerIndex != expectedProposer {
		return fmt.Errorf("shard blob header proposer should be %d, but got %d", expectedProposer, header.ProposerIndex)
	}

	// Verify builder and proposer aggregate signature
	dom, err := common.GetDomain(state, common.DOMAIN_SHARD_PROPOSER, headerEpoch)
	if err != nil {
		return err
	}
	signingRoot := common.ComputeSigningRoot(header.HashTreeRoot(tree.GetHashFn()), dom)
	builderPubkey, ok := epc.BuilderPubkeyCache.Pubkey(header.BuilderIndex)
	if !ok {
		return fmt.Errorf("could not find pubkey of shard blob builder %d", header.BuilderIndex)
	}
	proposerPubkey, ok := epc.ValidatorPubkeyCache.Pubkey(header.ProposerIndex)
	if !ok {
		return fmt.Errorf("could not find pubkey of shard blob proposer %d", header.ProposerIndex)
	}

	blsBuilderPub, err := builderPubkey.Pubkey()
	if err != nil {
		return fmt.Errorf("failed to deserialize cached builder pubkey: %v", err)
	}
	blsProposerPub, err := proposerPubkey.Pubkey()
	if err != nil {
		return fmt.Errorf("failed to deserialize cached validator pubkey: %v", err)
	}
	sig, err := signedHeader.Signature.Signature()
	if err != nil {
		return fmt.Errorf("failed to deserialize and sub-group check shard header signature: %v", err)
	}
	if !blsu.FastAggregateVerify([]*blsu.Pubkey{blsBuilderPub, blsProposerPub}, signingRoot[:], sig) {
		return errors.New("shard blob header has invalid signature")
	}

	// Verify the length by verifying the degree.
	bodySummary := &header.BodySummary
	if bodySummary.Commitment.Length == 0 {
		if bodySummary.DegreeProof != spec.G1_SETUP.Serialized[0] {
			return fmt.Errorf("degree proof for 0-length shard data should be first setup element, got: %s", bodySummary.DegreeProof.String())
		}
	}
	{
		engine := kbls.NewEngine()
		degreeProof, err := bodySummary.DegreeProof.Pubkey()
		if err != nil {
			return fmt.Errorf("failed to deserialize shard blob degree proof: %v", err)
		}
		engine.AddPair((*kbls.PointG1)(degreeProof), &spec.G2_SETUP.Points[0])
		commitment, err := bodySummary.Commitment.Point.Pubkey()
		if err != nil {
			return fmt.Errorf("failed to deserialize shard blob commitment: %v", err)
		}
		engine.AddPairInv((*kbls.PointG1)(commitment), &spec.G2_SETUP.Points[uint64(len(spec.G2_SETUP.Points))-uint64(bodySummary.Commitment.Length)])
		if !engine.Check() {
			return fmt.Errorf("failed to verify shard blob commitment %s (length %d)", bodySummary.Commitment.Point, bodySummary.Commitment.Length)
		}
	}

	// Charge EIP 1559 fee, builder pays for opportunity, and is responsible for later availability,
	// or fail to publish at their own expense.
	samples := bodySummary.Commitment.Length
	// TODO: overflows, need bigger int type (see spec)
	maxFee := bodySummary.MaxFeePerSample * common.Gwei(samples)

	// Builder must have sufficient balance, even if max_fee is not completely utilized
	builderBals, err := state.BlobBuilderBalances()
	if err != nil {
		return err
	}
	builderBalance, err := builderBals.GetBalance(header.BuilderIndex)
	if err != nil {
		return fmt.Errorf("failed to retrieve builder (%d) balance: %v", header.BuilderIndex, err)
	}
	if builderBalance < maxFee {
		return fmt.Errorf("builder does not have sufficient funds to pay for blob: got %d, max fee: %d", builderBalance, maxFee)
	}

	samplePrice, err := state.ShardSamplePrice()
	if err != nil {
		return err
	}
	baseFee := samplePrice * common.Gwei(samples)
	// Base fee must be paid
	if maxFee < baseFee {
		return fmt.Errorf("base fee cannot be covered: base fee %d for sample data is higher than max fee %d in header", baseFee, maxFee)
	}

	// Remaining fee goes towards proposer for prioritizing, up to a maximum
	maxPriorityFee := bodySummary.MaxPriorityFeePerSample * common.Gwei(samples)
	priorityFee := maxFee - baseFee
	if maxPriorityFee < priorityFee {
		priorityFee = maxPriorityFee
	}

	// Burn base fee, take priority fee
	// priority_fee <= max_fee - base_fee, thus priority_fee + base_fee <= max_fee, thus sufficient balance.
	builderBals.SetBalance(header.BuilderIndex, builderBalance-(baseFee+priorityFee))

	// Pay out priority fee
	valBals, err := state.Balances()
	if err != nil {
		return err
	}
	if err := common.IncreaseBalance(valBals, header.ProposerIndex, priorityFee); err != nil {
		return err
	}

	// Initialize the pending header
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

	beaconProposer, err := epc.GetBeaconProposer(stateSlot)
	if err != nil {
		return err
	}
	pendingHeader := PendingShardHeader{
		Attested: AttestedDataCommitment{
			Commitment:    header.BodySummary.Commitment,
			Root:          headerRoot,
			IncluderIndex: beaconProposer,
		},
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
