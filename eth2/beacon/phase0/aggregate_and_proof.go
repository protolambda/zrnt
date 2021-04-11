package phase0

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
)

// Given some committee size at a given slot, and a signature (not validated here) for that same slot
func IsAggregator(spec *common.Spec, commSize uint64, selectionProof common.BLSSignature) bool {
	modulo := commSize / spec.TARGET_AGGREGATORS_PER_COMMITTEE
	if modulo == 0 {
		modulo = 1
	}
	hash := sha256.New()
	hash.Write(selectionProof[:])
	return binary.LittleEndian.Uint64(hash.Sum(nil)[:8])%modulo == 0
}

func AggregateSelectionProofSigningRoot(spec *common.Spec, domainFn common.BLSDomainFn, slot common.Slot) (common.Root, error) {
	domain, err := domainFn(spec.DOMAIN_SELECTION_PROOF, spec.SlotToEpoch(slot))
	if err != nil {
		return common.Root{}, err
	}
	return common.ComputeSigningRoot(slot.HashTreeRoot(tree.GetHashFn()), domain), nil
}

func ValidateAggregateSelectionProof(spec *common.Spec, epc *common.EpochsContext, state common.BeaconState,
	slot common.Slot, commIndex common.CommitteeIndex, aggregator common.ValidatorIndex, selectionProof common.BLSSignature) (bool, error) {
	// check if the aggregator even exists
	vals, err := state.Validators()
	if err != nil {
		return false, err
	}
	if valid, err := vals.IsValidIndex(aggregator); err != nil {
		return false, err
	} else if !valid {
		return false, nil
	}
	// ge the relevant committee
	comm, err := epc.GetBeaconCommittee(slot, commIndex)
	if err != nil {
		// not an error, just not a valid committee index. Mark it as invalid.
		return false, nil
	}
	// check if the aggregator is part of the committee
	inComm := false
	for _, v := range comm {
		if v == aggregator {
			inComm = true
			break
		}
	}
	if !inComm {
		return false, nil
	}
	// check if the aggregator may actually aggregate
	if !IsAggregator(spec, uint64(len(comm)), selectionProof) {
		return false, nil
	}
	// check the selection proof
	sigRoot, err := AggregateSelectionProofSigningRoot(spec,
		func(typ common.BLSDomainType, epoch common.Epoch) (common.BLSDomain, error) {
			return common.GetDomain(state, typ, epoch)
		}, slot)
	if err != nil {
		return false, err
	}
	pub, ok := epc.PubkeyCache.Pubkey(aggregator)
	if !ok {
		return false, fmt.Errorf("could not fetch pubkey for aggregator %d", aggregator)
	}
	return bls.Verify(pub, sigRoot, selectionProof), nil
}

type SignedAggregateAndProof struct {
	Message   AggregateAndProof   `json:"message"`
	Signature common.BLSSignature `json:"signature"`
}

func (a *SignedAggregateAndProof) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&a.Message), &a.Signature)
}

func (a *SignedAggregateAndProof) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&a.Message), &a.Signature)
}

func (a *SignedAggregateAndProof) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&a.Message), &a.Signature)
}

func (a *SignedAggregateAndProof) FixedLength(*common.Spec) uint64 {
	return 0
}

func (a *SignedAggregateAndProof) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&a.Message), &a.Signature)
}

type AggregateAndProof struct {
	AggregatorIndex common.ValidatorIndex `json:"aggregator_index"`
	Aggregate       Attestation           `json:"aggregate"`
	SelectionProof  common.BLSSignature   `json:"selection_proof"`
}

func (a *AggregateAndProof) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&a.AggregatorIndex, spec.Wrap(&a.Aggregate), &a.SelectionProof)
}

func (a *AggregateAndProof) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&a.AggregatorIndex, spec.Wrap(&a.Aggregate), &a.SelectionProof)
}

func (a *AggregateAndProof) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&a.AggregatorIndex, spec.Wrap(&a.Aggregate), &a.SelectionProof)
}

func (a *AggregateAndProof) FixedLength(*common.Spec) uint64 {
	return 0
}

func (a *AggregateAndProof) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&a.AggregatorIndex, spec.Wrap(&a.Aggregate), &a.SelectionProof)
}
