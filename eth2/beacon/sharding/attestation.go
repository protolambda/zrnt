package sharding

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
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
	phase0Att := phase0.Attestation{
		AggregationBits: attestation.AggregationBits,
		Data: phase0.AttestationData{
			Slot:            attestation.Data.Slot,
			Index:           attestation.Data.Index,
			BeaconBlockRoot: attestation.Data.BeaconBlockRoot,
			Source:          attestation.Data.Source,
			Target:          attestation.Data.Target,
		},
		Signature: attestation.Signature,
	}
	if err := phase0.ProcessAttestation(spec, epc, state, &phase0Att); err != nil {
		return err
	}
	return UpdatePendingShardWork(spec, epc, state, attestation)
}

func UpdatePendingShardWork(spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, attestation *Attestation) error {
	attestationShard, err := epc.ComputeShardFromCommitteeIndex(attestation.Data.Index)
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
	if selector != SHARD_WORK_PENDING {
		// Attestation doesn't need to be processed, shard/slot pair is unconfirmable or already confirmed
		return nil
	}
	currentHeaders, err := AsPendingShardHeaders(workStatus.Value())
	if err != nil {
		return err
	}
	// Find the corresponding header, abort if it cannot be found
	headerIndex := uint64(0)
	found := false
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
		if pendingHeaderRoot == attestation.Data.ShardHeaderRoot {
			found = true
			break
		}
		headerIndex++
	}
	if !found {
		return fmt.Errorf("attestation for unknown shard header is invalid, header root: %s", attestation.Data.ShardHeaderRoot)
	}

	pendingHeader, err := AsPendingShardHeader(currentHeaders.Get(headerIndex))
	if err != nil {
		return err
	}
	fullCommittee, err := epc.GetBeaconCommittee(attestation.Data.Slot, attestation.Data.Index)
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
		pendingBitsView, err := pendingHeader.Votes()
		if err != nil {
			return err
		}
		pendingBits, err := pendingBitsView.Raw()
		if err != nil {
			return err
		}
		if spec.SlotToEpoch(updateSlot) < epc.CurrentEpoch.Epoch {
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
		commView, err := pendingHeader.Commitment()
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
			if err := workStatus.Change(SHARD_WORK_CONFIRMED, commView); err != nil {
				return err
			}
		}
	}

	return nil
}
