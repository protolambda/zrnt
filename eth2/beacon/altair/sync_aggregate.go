package altair

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/ztyp/bitfields"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/conv"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

// SyncCommitteeBits is formatted as a serialized SSZ bitvector,
// with trailing zero bits if length does not align with byte length.
type SyncCommitteeBits []byte

func (li SyncCommitteeBits) View(spec *common.Spec) *SyncCommitteeBitsView {
	v, _ := SyncCommitteeBitsType(spec).Deserialize(codec.NewDecodingReader(bytes.NewReader(li), uint64(len(li))))
	return &SyncCommitteeBitsView{v.(*BitVectorView)}
}

func (li *SyncCommitteeBits) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.BitVector((*[]byte)(li), spec.SYNC_COMMITTEE_SIZE)
}

func (a SyncCommitteeBits) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.BitVector(a[:])
}

func (a SyncCommitteeBits) ByteLength(spec *common.Spec) uint64 {
	return (spec.SYNC_COMMITTEE_SIZE + 7) / 8
}

func (a *SyncCommitteeBits) FixedLength(spec *common.Spec) uint64 {
	return (spec.SYNC_COMMITTEE_SIZE + 7) / 8
}

func (li SyncCommitteeBits) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.BitVectorHTR(li)
}

func (cb SyncCommitteeBits) MarshalText() ([]byte, error) {
	return conv.BytesMarshalText(cb[:])
}

func (cb *SyncCommitteeBits) UnmarshalText(text []byte) error {
	return conv.DynamicBytesUnmarshalText((*[]byte)(cb), text)
}

func (cb SyncCommitteeBits) String() string {
	return conv.BytesString(cb[:])
}

func (cb SyncCommitteeBits) GetBit(i uint64) bool {
	return bitfields.GetBit(cb, i)
}

func (cb SyncCommitteeBits) SetBit(i uint64, v bool) {
	bitfields.SetBit(cb, i, v)
}

type SyncCommitteeBitsView struct {
	*BitVectorView
}

func AsSyncCommitteeBits(v View, err error) (*SyncCommitteeBitsView, error) {
	c, err := AsBitVector(v, err)
	return &SyncCommitteeBitsView{c}, err
}

func (v *SyncCommitteeBitsView) Raw(spec *common.Spec) (SyncCommitteeBits, error) {
	byteLen := int((spec.SYNC_COMMITTEE_SIZE + 7) / 8)
	var buf bytes.Buffer
	buf.Grow(byteLen)
	if err := v.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		return nil, err
	}
	out := SyncCommitteeBits(buf.Bytes())
	if len(out) != byteLen {
		return nil, fmt.Errorf("failed to convert sync committee tree bits view to raw bits")
	}
	return out, nil
}

func SyncCommitteeBitsType(spec *common.Spec) *BitVectorTypeDef {
	return BitVectorType(spec.SYNC_COMMITTEE_SIZE)
}

func SyncAggregateType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("SyncAggregate", []FieldDef{
		{"sync_committee_bits", SyncCommitteeBitsType(spec)},
		{"sync_committee_signature", common.BLSSignatureType},
	})
}

type SyncAggregate struct {
	SyncCommitteeBits      SyncCommitteeBits   `json:"sync_committee_bits" yaml:"sync_committee_bits"`
	SyncCommitteeSignature common.BLSSignature `json:"sync_committee_signature" yaml:"sync_committee_signature"`
}

func (agg *SyncAggregate) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(
		spec.Wrap(&agg.SyncCommitteeBits),
		&agg.SyncCommitteeSignature,
	)
}

func (agg *SyncAggregate) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(
		spec.Wrap(&agg.SyncCommitteeBits),
		&agg.SyncCommitteeSignature,
	)
}

func (agg *SyncAggregate) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(
		spec.Wrap(&agg.SyncCommitteeBits),
		&agg.SyncCommitteeSignature,
	)
}

func (agg *SyncAggregate) FixedLength(spec *common.Spec) uint64 {
	// bitvector + signature
	return (spec.SYNC_COMMITTEE_SIZE+7)/8 + 96
}

func (agg *SyncAggregate) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		spec.Wrap(&agg.SyncCommitteeBits),
		&agg.SyncCommitteeSignature,
	)
}

type SyncAggregateView struct {
	*ContainerView
}

func AsSyncAggregate(v View, err error) (*SyncAggregateView, error) {
	c, err := AsContainer(v, err)
	return &SyncAggregateView{c}, err
}

func ProcessSyncAggregate(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, agg *SyncAggregate) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	currentSlot, err := state.Slot()
	if err != nil {
		return err
	}
	if err := bitfields.BitvectorCheck(agg.SyncCommitteeBits, spec.SYNC_COMMITTEE_SIZE); err != nil {
		return fmt.Errorf("input bypassed deserialization checks, sanity check on sync committee bitvector length failed: %v", err)
	}

	if epc.CurrentSyncCommittee == nil {
		return fmt.Errorf("missing current sync committee info in EPC")
	}

	participantPubkeys := make([]*common.CachedPubkey, 0, spec.SYNC_COMMITTEE_SIZE)
	for i := uint64(0); i < spec.SYNC_COMMITTEE_SIZE; i++ {
		if agg.SyncCommitteeBits.GetBit(i) {
			participantPubkeys = append(participantPubkeys, epc.CurrentSyncCommittee.CachedPubkeys[i])
		}
	}

	prevSlot := currentSlot.Previous()
	domain, err := common.GetDomain(state, common.DOMAIN_SYNC_COMMITTEE, spec.SlotToEpoch(prevSlot))
	if err != nil {
		return err
	}
	blockRoot, err := common.GetBlockRootAtSlot(spec, state, prevSlot)
	if err != nil {
		return err
	}
	signingRoot := common.ComputeSigningRoot(blockRoot, domain)
	if !bls.Eth2FastAggregateVerify(participantPubkeys, signingRoot, agg.SyncCommitteeSignature) {
		return errors.New("invalid sync committee signature")
	}

	// Compute participant and proposer rewards
	totalActiveIncrements := epc.TotalActiveStake / spec.EFFECTIVE_BALANCE_INCREMENT
	baseRewardPerIncrement := (spec.EFFECTIVE_BALANCE_INCREMENT * common.Gwei(spec.BASE_REWARD_FACTOR)) / epc.TotalActiveStakeSqRoot
	totalBaseRewards := baseRewardPerIncrement * totalActiveIncrements
	maxParticipantRewards := (totalBaseRewards * SYNC_REWARD_WEIGHT) / WEIGHT_DENOMINATOR / common.Gwei(spec.SLOTS_PER_EPOCH)
	participantReward := maxParticipantRewards / common.Gwei(spec.SYNC_COMMITTEE_SIZE)
	proposerReward := participantReward * PROPOSER_WEIGHT / (WEIGHT_DENOMINATOR - PROPOSER_WEIGHT)

	// Apply participant rewards and penalties
	bals, err := state.Balances()
	if err != nil {
		return err
	}
	// Note: the minimum effective balance of the proposer is sufficient
	// to not result in differences from spec operations
	for i := uint64(0); i < spec.SYNC_COMMITTEE_SIZE; i++ {
		validatorIndex := epc.CurrentSyncCommittee.Indices[i]
		if agg.SyncCommitteeBits.GetBit(i) {
			if err := common.IncreaseBalance(bals, validatorIndex, participantReward); err != nil {
				return err
			}
		} else {
			if err := common.DecreaseBalance(bals, validatorIndex, participantReward); err != nil {
				return err
			}
		}
	}
	// Apply proposer rewards
	proposer, err := epc.GetBeaconProposer(currentSlot)
	if err != nil {
		return err
	}
	proposerRewardSum := proposerReward * common.Gwei(len(participantPubkeys))
	if err := common.IncreaseBalance(bals, proposer, proposerRewardSum); err != nil {
		return err
	}
	return nil
}

func ProcessSyncCommitteeUpdates(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, state common.SyncCommitteeBeaconState) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	nextEpoch := epc.NextEpoch.Epoch
	if nextEpoch%spec.EPOCHS_PER_SYNC_COMMITTEE_PERIOD == 0 {
		next, err := common.ComputeNextSyncCommittee(spec, epc, state)
		if err != nil {
			return fmt.Errorf("failed to update sync committee: %v", next)
		}
		nextView, err := next.View(spec)
		if err != nil {
			return fmt.Errorf("failed to convert sync committee to state tree representation")
		}
		if err := state.RotateSyncCommittee(nextView); err != nil {
			return fmt.Errorf("failed to rotate sync committee: %v", err)
		}
	}
	return nil
}
