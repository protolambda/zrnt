package altair

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	hbls "github.com/herumi/bls-eth-go-binary/bls"
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

func SyncCommitteePubkeysType(spec *common.Spec) *ComplexVectorTypeDef {
	return ComplexVectorType(common.BLSPubkeyType, spec.SYNC_COMMITTEE_SIZE)
}

type SyncCommitteePubkeys []common.BLSPubkey

func (li *SyncCommitteePubkeys) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	*li = make([]common.BLSPubkey, spec.SYNC_COMMITTEE_SIZE, spec.SYNC_COMMITTEE_SIZE)
	return dr.Vector(func(i uint64) codec.Deserializable {
		return &(*li)[i]
	}, common.BLSPubkeyType.Size, spec.SYNC_COMMITTEE_SIZE)
}

func (a SyncCommitteePubkeys) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Vector(func(i uint64) codec.Serializable {
		return &a[i]
	}, common.BLSPubkeyType.Size, spec.SYNC_COMMITTEE_SIZE)
}

func (a SyncCommitteePubkeys) ByteLength(spec *common.Spec) uint64 {
	return spec.SYNC_COMMITTEE_SIZE * common.BLSPubkeyType.Size
}

func (a *SyncCommitteePubkeys) FixedLength(spec *common.Spec) uint64 {
	return spec.SYNC_COMMITTEE_SIZE * common.BLSPubkeyType.Size
}

func (li SyncCommitteePubkeys) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.ComplexVectorHTR(func(i uint64) tree.HTR {
		return &li[i]
	}, spec.SYNC_COMMITTEE_SIZE)
}

type SyncCommitteePubkeysView struct {
	*ComplexVectorView
}

func AsSyncCommitteePubkeys(v View, err error) (*SyncCommitteePubkeysView, error) {
	c, err := AsComplexVector(v, err)
	return &SyncCommitteePubkeysView{c}, err
}

func SyncCommitteeType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("SyncCommittee", []FieldDef{
		{"pubkeys", SyncCommitteePubkeysType(spec)},
		{"aggregate_pubkey", common.BLSPubkeyType},
	})
}

type SyncCommittee struct {
	Pubkeys         SyncCommitteePubkeys `json:"pubkeys" yaml:"pubkeys"`
	AggregatePubkey common.BLSPubkey     `json:"aggregate_pubkey" yaml:"aggregate_pubkey"`
}

func (a *SyncCommittee) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(spec.Wrap(&a.Pubkeys), &a.AggregatePubkey)
}

func (a *SyncCommittee) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.FixedLenContainer(spec.Wrap(&a.Pubkeys), &a.AggregatePubkey)
}

func (a *SyncCommittee) ByteLength(spec *common.Spec) uint64 {
	return (spec.SYNC_COMMITTEE_SIZE + 1) * common.BLSPubkeyType.Size
}

func (*SyncCommittee) FixedLength(spec *common.Spec) uint64 {
	return (spec.SYNC_COMMITTEE_SIZE + 1) * common.BLSPubkeyType.Size
}

func (p *SyncCommittee) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&p.Pubkeys), &p.AggregatePubkey)
}

func (p *SyncCommittee) View(spec *common.Spec) (*SyncCommitteeView, error) {
	elems := make([]View, len(p.Pubkeys), len(p.Pubkeys))
	for i := 0; i < len(p.Pubkeys); i++ {
		elems[i] = common.ViewPubkey(&p.Pubkeys[i])
	}
	pubs, err := SyncCommitteePubkeysType(spec).FromElements(elems...)
	if err != nil {
		return nil, err
	}
	return AsSyncCommittee(SyncCommitteeType(spec).FromFields(pubs, common.ViewPubkey(&p.AggregatePubkey)))
}

type SyncCommitteeView struct {
	*ContainerView
}

func AsSyncCommittee(v View, err error) (*SyncCommitteeView, error) {
	c, err := AsContainer(v, err)
	return &SyncCommitteeView{c}, err
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

func (agg *SyncAggregate) FixedLength(*common.Spec) uint64 {
	return 0
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

func ProcessSyncCommittee(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, agg *SyncAggregate) error {
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
	prevSlot := currentSlot.Previous()
	// TODO syncCommitteeIndices = get_sync_committee_indices(state, get_current_epoch(state))
	syncCommitteeIndices := make([]common.ValidatorIndex, spec.SYNC_COMMITTEE_SIZE, spec.SYNC_COMMITTEE_SIZE)
	// TODO
	syncCommitteePubkeys := make([]*common.CachedPubkey, spec.SYNC_COMMITTEE_SIZE, spec.SYNC_COMMITTEE_SIZE)

	includedIndices := make([]common.ValidatorIndex, 0, len(syncCommitteeIndices))
	includedPubkeys := make([]*common.CachedPubkey, 0, len(syncCommitteeIndices))
	for i := uint64(0); i < spec.SYNC_COMMITTEE_SIZE; i++ {
		if agg.SyncCommitteeBits.GetBit(i) {
			includedIndices = append(includedIndices, syncCommitteeIndices[i])
			includedPubkeys = append(includedPubkeys, syncCommitteePubkeys[i])
		}
	}
	domain, err := common.GetDomain(state, spec.DOMAIN_SYNC_COMMITTEE, spec.SlotToEpoch(prevSlot))
	if err != nil {
		return err
	}
	blockRoot, err := common.GetBlockRootAtSlot(spec, state, prevSlot)
	if err != nil {
		return err
	}
	signingRoot := common.ComputeSigningRoot(blockRoot, domain)
	if !bls.Eth2FastAggregateVerify(includedPubkeys, signingRoot, agg.SyncCommitteeSignature) {
		return errors.New("invalid sync committee signature")
	}
	totalActiveIncrements := epc.TotalActiveStake / spec.EFFECTIVE_BALANCE_INCREMENT
	baseRewardPerIncrement := (spec.EFFECTIVE_BALANCE_INCREMENT * common.Gwei(spec.BASE_REWARD_FACTOR)) / epc.TotalActiveStakeSqRoot
	totalBaseRewards := baseRewardPerIncrement * totalActiveIncrements
	maxEpochRewards := (totalBaseRewards * common.Gwei(SYNC_REWARD_WEIGHT)) / common.Gwei(WEIGHT_DENOMINATOR)
	maxSlotRewards := ((maxEpochRewards * common.Gwei(len(includedIndices))) /
		common.Gwei(len(syncCommitteeIndices))) / common.Gwei(spec.SLOTS_PER_EPOCH)

	committeeEffBalance := common.Gwei(0)
	for _, ci := range includedIndices {
		committeeEffBalance += epc.EffectiveBalances[ci]
	}
	if committeeEffBalance < spec.EFFECTIVE_BALANCE_INCREMENT {
		committeeEffBalance = spec.EFFECTIVE_BALANCE_INCREMENT
	}
	bals, err := state.Balances()
	if err != nil {
		return err
	}
	proposerRewardSum := common.Gwei(0)
	for _, ci := range includedIndices {
		effectiveBalance := epc.EffectiveBalances[ci]
		inclusionReward := (maxSlotRewards * effectiveBalance) / committeeEffBalance
		proposerReward := inclusionReward / common.Gwei(spec.PROPOSER_REWARD_QUOTIENT)
		proposerRewardSum += proposerReward
		if err := common.IncreaseBalance(bals, ci, inclusionReward-proposerReward); err != nil {
			return err
		}
	}
	proposer, err := epc.GetBeaconProposer(currentSlot)
	if err != nil {
		return err
	}
	if err := common.IncreaseBalance(bals, proposer, proposerRewardSum); err != nil {
		return err
	}
	return nil
}

func ComputeNextSyncCommittee(spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView) (*SyncCommittee, error) {
	indices, err := common.ComputeSyncCommitteeIndices(spec, state, epc.NextEpoch.Epoch, epc.NextEpoch.ActiveIndices)
	if err != nil {
		return nil, fmt.Errorf("failed to compute sync committee indices for next epoch: %d", indices)
	}
	var pubs []common.BLSPubkey
	var blsAggregate hbls.PublicKey
	for _, idx := range indices {
		pub, ok := epc.PubkeyCache.Pubkey(idx)
		if !ok {
			return nil, fmt.Errorf("failed to get sync committee data, pubkey cache is missing pubkey for index %d: %v", idx, err)
		}
		blsPub, err := pub.Pubkey()
		if err != nil {
			return nil, fmt.Errorf("pubkey cache contains invalid pubkey at index %d: %v", idx, err)
		}
		blsAggregate.Add(blsPub)
		pubs = append(pubs, pub.Compressed)
	}
	var aggregate common.BLSPubkey
	copy(aggregate[:], blsAggregate.Serialize())
	return &SyncCommittee{
		Pubkeys:         pubs,
		AggregatePubkey: aggregate,
	}, nil
}

func ProcessSyncCommitteeUpdates(spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView) error {
	nextEpoch := epc.NextEpoch.Epoch
	if nextEpoch%spec.EPOCHS_PER_SYNC_COMMITTEE_PERIOD == 0 {
		next, err := ComputeNextSyncCommittee(spec, epc, state)
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
