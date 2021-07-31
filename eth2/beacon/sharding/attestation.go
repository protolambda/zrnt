package sharding

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
	"sort"
)

func BlockAttestationsType(spec *common.Spec) ListTypeDef {
	return ListType(AttestationType(spec), spec.MAX_ATTESTATIONS)
}

func AttestationType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("Attestation", []FieldDef{
		{"aggregation_bits", phase0.AttestationBitsType(spec)},
		{"data", AttestationDataType},
		{"signature", common.BLSSignatureType},
	})
}

type Attestation struct {
	AggregationBits phase0.AttestationBits `json:"aggregation_bits" yaml:"aggregation_bits"`
	Data            AttestationData        `json:"data" yaml:"data"`
	Signature       common.BLSSignature    `json:"signature" yaml:"signature"`
}

func (a *Attestation) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&a.AggregationBits), &a.Data, &a.Signature)
}

func (a *Attestation) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&a.AggregationBits), &a.Data, &a.Signature)
}

func (a *Attestation) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&a.AggregationBits), &a.Data, &a.Signature)
}

func (a *Attestation) FixedLength(*common.Spec) uint64 {
	return 0
}

func (a *Attestation) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&a.AggregationBits), &a.Data, a.Signature)
}

type Attestations []Attestation

func (a *Attestations) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, Attestation{})
		return spec.Wrap(&((*a)[i]))
	}, 0, spec.MAX_ATTESTATIONS)
}

func (a Attestations) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return spec.Wrap(&a[i])
	}, 0, uint64(len(a)))
}

func (a Attestations) ByteLength(spec *common.Spec) (out uint64) {
	for _, v := range a {
		out += v.ByteLength(spec) + codec.OFFSET_SIZE
	}
	return
}

func (a *Attestations) FixedLength(*common.Spec) uint64 {
	return 0
}

func (li Attestations) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(&li[i])
		}
		return nil
	}, length, spec.MAX_ATTESTATIONS)
}

func ProcessAttestations(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, ops []Attestation) error {
	for i := range ops {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := ProcessAttestation(spec, epc, state, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

func ProcessAttestation(spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, attestation *Attestation) error {
	if err := ClassicProcessAttestation(spec, epc, state, attestation); err != nil {
		return err
	}
	return ProcessAttestedShardWork(spec, epc, state, attestation)
}

func ClassicProcessAttestation(spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, attestation *Attestation) error {
	data := &attestation.Data

	currentSlot, err := state.Slot()
	if err != nil {
		return err
	}

	currentEpoch := spec.SlotToEpoch(currentSlot)
	previousEpoch := currentEpoch.Previous()

	// Check target
	if data.Target.Epoch < previousEpoch {
		return errors.New("attestation data is invalid, target is too far in past")
	} else if data.Target.Epoch > currentEpoch {
		return errors.New("attestation data is invalid, target is in future")
	}
	// And if it matches the slot
	if data.Target.Epoch != spec.SlotToEpoch(data.Slot) {
		return errors.New("attestation data is invalid, slot epoch does not match target epoch")
	}

	// safe additions, slot converts to a valid epoch, thus must be low
	if !(currentSlot <= data.Slot+spec.SLOTS_PER_EPOCH) {
		return errors.New("attestation slot is too old")
	}
	if !(data.Slot+spec.MIN_ATTESTATION_INCLUSION_DELAY <= currentSlot) {
		return errors.New("attestation is too new")
	}

	// Check committee index
	if commCount, err := epc.GetCommitteeCountPerSlot(data.Target.Epoch); err != nil {
		return err
	} else if uint64(data.Index) >= commCount {
		return errors.New("attestation data is invalid, committee index out of range")
	}

	// Note: this checks the source checkpoint.
	applyFlags, err := GetApplicableAttestationParticipationFlags(spec, state, data, currentSlot-data.Slot)
	if err != nil {
		return err
	}

	// Check signature and bitfields
	committee, err := epc.GetBeaconCommittee(data.Slot, data.Index)
	if err != nil {
		return err
	}
	indexedAtt, err := attestation.ConvertToIndexed(spec, committee)
	if err != nil {
		return fmt.Errorf("attestation could not be converted to an indexed attestation: %v", err)
	} else if err := ValidateIndexedAttestation(spec, epc, state, indexedAtt); err != nil {
		return fmt.Errorf("attestation could not be verified in its indexed form: %v", err)
	}

	var epochParticipation *altair.ParticipationRegistryView
	// Check source
	if data.Target.Epoch == currentEpoch {
		epochParticipation, err = state.CurrentEpochParticipation()
		if err != nil {
			return err
		}
	} else {
		epochParticipation, err = state.PreviousEpochParticipation()
		if err != nil {
			return err
		}
	}

	// TODO: probably better to batch flag changes, needs optimization, tree structure not good for this.
	proposerRewardNumerator := common.Gwei(0)
	baseRewardPerIncrement := spec.EFFECTIVE_BALANCE_INCREMENT * common.Gwei(spec.BASE_REWARD_FACTOR) / epc.TotalActiveStakeSqRoot
	for _, vi := range indexedAtt.AttestingIndices {
		if applyFlags == 0 { // no work to do, just skip ahead
			continue
		}
		increments := epc.EffectiveBalances[vi] / spec.EFFECTIVE_BALANCE_INCREMENT
		baseReward := increments * baseRewardPerIncrement
		existingFlags, err := epochParticipation.GetFlags(vi)
		if err != nil {
			return err
		}
		if (applyFlags&altair.TIMELY_SOURCE_FLAG != 0) && (existingFlags&altair.TIMELY_SOURCE_FLAG == 0) {
			proposerRewardNumerator += baseReward * altair.TIMELY_SOURCE_WEIGHT
		}
		if (applyFlags&altair.TIMELY_TARGET_FLAG != 0) && (existingFlags&altair.TIMELY_TARGET_FLAG == 0) {
			proposerRewardNumerator += baseReward * altair.TIMELY_TARGET_WEIGHT
		}
		if (applyFlags&altair.TIMELY_HEAD_FLAG != 0) && (existingFlags&altair.TIMELY_HEAD_FLAG == 0) {
			proposerRewardNumerator += baseReward * altair.TIMELY_HEAD_WEIGHT
		}
		if err := epochParticipation.SetFlags(vi, existingFlags|applyFlags); err != nil {
			return err
		}
	}
	proposerRewardDenominator := ((altair.WEIGHT_DENOMINATOR - altair.PROPOSER_WEIGHT) * altair.WEIGHT_DENOMINATOR) / altair.PROPOSER_WEIGHT
	proposerReward := proposerRewardNumerator / proposerRewardDenominator
	proposerIndex, err := epc.GetBeaconProposer(currentSlot)
	if err != nil {
		return err
	}
	bals, err := state.Balances()
	if err != nil {
		return err
	}
	if err := common.IncreaseBalance(bals, proposerIndex, proposerReward); err != nil {
		return err
	}
	return nil
}

func GetApplicableAttestationParticipationFlags(
	spec *common.Spec, state altair.AltairLikeBeaconState,
	data *AttestationData, inclusionDelay common.Slot) (out altair.ParticipationFlags, err error) {

	currentSlot, err := state.Slot()
	if err != nil {
		return 0, err
	}

	currentEpoch := spec.SlotToEpoch(currentSlot)

	var justifiedCheckpoint common.Checkpoint
	if data.Target.Epoch == currentEpoch {
		justifiedCheckpoint, err = state.CurrentJustifiedCheckpoint()
		if err != nil {
			return 0, err
		}
	} else {
		justifiedCheckpoint, err = state.PreviousJustifiedCheckpoint()
		if err != nil {
			return 0, err
		}
	}
	expectedHead, err := common.GetBlockRootAtSlot(spec, state, data.Slot)
	if err != nil {
		return 0, err
	}
	expectedTarget, err := common.GetBlockRoot(spec, state, data.Target.Epoch)
	if err != nil {
		return 0, err
	}

	isMatchingSource := data.Source == justifiedCheckpoint
	isMatchingTarget := isMatchingSource && expectedTarget == data.Target.Root
	isMatchingHead := isMatchingTarget && expectedHead == data.BeaconBlockRoot

	if !isMatchingSource {
		return 0, fmt.Errorf("source %s must match justified %s", data.Source, justifiedCheckpoint)
	}

	if isMatchingSource && inclusionDelay <= common.Slot(math.IntegerSquareroot(uint64(spec.SLOTS_PER_EPOCH))) {
		out |= altair.TIMELY_SOURCE_FLAG
	}
	if isMatchingTarget && inclusionDelay <= spec.SLOTS_PER_EPOCH {
		out |= altair.TIMELY_TARGET_FLAG
	}
	if isMatchingHead && inclusionDelay == spec.MIN_ATTESTATION_INCLUSION_DELAY {
		out |= altair.TIMELY_HEAD_FLAG
	}
	return out, nil
}

// Convert attestation to (almost) indexed-verifiable form
func (attestation *Attestation) ConvertToIndexed(spec *common.Spec, committee []common.ValidatorIndex) (*IndexedAttestation, error) {
	bitLen := attestation.AggregationBits.BitLen()
	if uint64(len(committee)) != bitLen {
		return nil, fmt.Errorf("committee size does not match bits size: %d <> %d", len(committee), bitLen)
	}

	participants := make([]common.ValidatorIndex, 0, len(committee))
	for i := uint64(0); i < bitLen; i++ {
		if attestation.AggregationBits.GetBit(i) {
			participants = append(participants, committee[i])
		}
	}
	sort.Slice(participants, func(i int, j int) bool {
		return participants[i] < participants[j]
	})

	return &IndexedAttestation{
		AttestingIndices: participants,
		Data:             attestation.Data,
		Signature:        attestation.Signature,
	}, nil
}

const TIMELY_SHARD_FLAG_INDEX uint8 = 3

const TIMELY_SHARD_FLAG altair.ParticipationFlags = 1 << TIMELY_SHARD_FLAG_INDEX

func batchApplyParticipationFlag(spec *common.Spec, state *BeaconStateView, bits phase0.AttestationBits, epoch common.Epoch, fullCommittee common.CommitteeIndices, flag altair.ParticipationFlags) error {
	slot, err := state.Slot()
	if err != nil {
		return err
	}
	currentEpoch := spec.SlotToEpoch(slot)
	var epochParticipation *altair.ParticipationRegistryView
	if epoch == currentEpoch {
		epochParticipation, err = state.CurrentEpochParticipation()
		if err != nil {
			return err
		}
	} else {
		epochParticipation, err = state.PreviousEpochParticipation()
		if err != nil {
			return err
		}
	}
	for i, index := range fullCommittee {
		if bits.GetBit(uint64(i)) {
			prev, err := epochParticipation.GetFlags(index)
			if err != nil {
				return err
			}
			epochParticipation.SetFlags(index, prev|flag)
		}
	}
	return nil
}

func ProcessAttestedShardWork(spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, attestation *Attestation) error {
	attestationShard, err := epc.ComputeShardFromCommitteeIndex(attestation.Data.Slot, attestation.Data.Index)
	if err != nil {
		return err
	}

	fullCommittee, err := epc.GetBeaconCommittee(attestation.Data.Slot, attestation.Data.Index)
	if err != nil {
		return err
	}

	bufferIndex := uint64(attestation.Data.Slot % spec.SHARD_STATE_MEMORY_SLOTS)
	// Check that this data is still pending
	buffer, err := state.ShardBuffer()
	if err != nil {
		return err
	}
	column, err := buffer.Column(bufferIndex)
	if err != nil {
		return err
	}
	committeeWork, err := column.GetWork(attestationShard)
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
	// Skip attestation vote accounting if the header is not pending
	if selector != SHARD_WORK_PENDING {
		// If the data was already confirmed, check if this matches, to apply the flag to the attesters.
		if selector == SHARD_WORK_CONFIRMED {
			attested, err := AsAttestedDataCommitment(workStatus.Value())
			if err != nil {
				return err
			}
			attestedRoot, err := attested.Root()
			if err != nil {
				return err
			}
			if attestedRoot == attestation.Data.ShardBlobRoot {
				batchApplyParticipationFlag(spec, state,
					attestation.AggregationBits, attestation.Data.Target.Epoch,
					fullCommittee, TIMELY_SHARD_FLAG)
			}
		}
		return nil
	}

	currentHeaders, err := AsPendingShardHeaders(workStatus.Value())
	if err != nil {
		return err
	}
	// Find the corresponding header, abort if it cannot be found
	headerIndex := uint64(0)
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
		attested, err := pendingHeader.Attested()
		if err != nil {
			return err
		}
		attestedHeaderRoot, err := attested.Root()
		if err != nil {
			return err
		}
		if attestedHeaderRoot == attestation.Data.ShardBlobRoot {
			break
		}
		headerIndex++
	}
	currentHeadersLen, err := currentHeaders.Length()
	if err != nil {
		return err
	}
	if headerIndex == currentHeadersLen {
		// Not an error, the attestation may be early (header not included yet), or too late (too many headers, attested header did not make it in)
		return nil
	}

	pendingHeader, err := AsPendingShardHeader(currentHeaders.Get(headerIndex))
	if err != nil {
		return err
	}

	// The weight may be outdated if it is not the initial weight, and from a previous epoch
	weight, err := pendingHeader.Weight()
	if err != nil {
		return err
	}
	if weight != 0 {
		updateSlot, err := pendingHeader.UpdateSlot()
		if err != nil {
			return err
		}
		if spec.SlotToEpoch(updateSlot) < epc.CurrentEpoch.Epoch {
			pendingBitsView, err := pendingHeader.Votes()
			if err != nil {
				return err
			}
			pendingBits, err := pendingBitsView.Raw()
			if err != nil {
				return err
			}
			weight := common.Gwei(0)
			for i, valIndex := range fullCommittee {
				if pendingBits.GetBit(uint64(i)) {
					weight += epc.EffectiveBalances[valIndex]
				}
			}
			if err := pendingHeader.SetWeight(weight); err != nil {
				return err
			}
		}
	}

	slot, err := state.Slot()
	if err != nil {
		return err
	}
	if err := pendingHeader.SetUpdateSlot(slot); err != nil {
		return err
	}

	// Update votes bitfield in the state, update weights
	pendingBitsView, err := pendingHeader.Votes()
	if err != nil {
		return err
	}
	pendingBits, err := pendingBitsView.Raw()
	if err != nil {
		return err
	}
	fullCommitteeBalance := common.Gwei(0)
	anyChange := false
	for i, valIndex := range fullCommittee {
		eff := epc.EffectiveBalances[valIndex]
		fullCommitteeBalance += eff
		if attestation.AggregationBits.GetBit(uint64(i)) {
			if !pendingBits.GetBit(uint64(i)) {
				weight += epc.EffectiveBalances[valIndex]
				pendingBits.SetBit(uint64(i), true)
				anyChange = true
			}
		}
	}
	if anyChange {
		if err := pendingHeader.SetWeight(weight); err != nil {
			return err
		}
		if err := pendingHeader.SetVotes(pendingBits.View(spec)); err != nil {
			return err
		}
	}

	// Check if the PendingShardHeader is eligible for expedited confirmation, requiring 2/3 of balance attesting
	if weight*3 > fullCommitteeBalance*2 {
		// participants of the winning header are remembered with participation flags
		batchApplyParticipationFlag(spec, state, pendingBits, attestation.Data.Target.Epoch,
			fullCommittee, TIMELY_SHARD_FLAG)

		attView, err := pendingHeader.Attested()
		if err != nil {
			return err
		}
		commView, err := attView.Commitment()
		if err != nil {
			return err
		}
		commitment, err := commView.Raw()
		if err != nil {
			return err
		}
		if *commitment == (DataCommitment{}) {
			// The committee voted to not confirm anything
			if err := workStatus.Change(SHARD_WORK_UNCONFIRMED, nil); err != nil {
				return err
			}
		} else {
			if err := workStatus.Change(SHARD_WORK_CONFIRMED, attView); err != nil {
				return err
			}
		}
	}

	return nil
}
