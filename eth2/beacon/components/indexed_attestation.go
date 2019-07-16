package components

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	"sort"
)

var AttestationDataAndCustodyBitSSZ = zssz.GetSSZ((*AttestationDataAndCustodyBit)(nil))

type AttestationDataAndCustodyBit struct {
	Data       AttestationData
	CustodyBit bool // Challengeable bit (SSZ-bool, 1 byte) for the custody of crosslink data
}

type CommitteeIndices []ValidatorIndex

func (ci *CommitteeIndices) Limit() uint64 {
	return MAX_VALIDATORS_PER_COMMITTEE
}

type IndexedAttestation struct {
	// Indices with custody bit equal to 0
	CustodyBit0Indices CommitteeIndices
	// Indices with custody bit equal to 1
	CustodyBit1Indices CommitteeIndices

	Data      AttestationData
	Signature BLSSignature
}

// Verify validity of slashable_attestation fields.
func (state *BeaconState) ValidateIndexedAttestation(indexedAttestation *IndexedAttestation) error {
	// wrap it in validator-sets. Does not sort it, but does make checking if it is a lot easier.
	bit0Indices := ValidatorSet(indexedAttestation.CustodyBit0Indices)
	bit1Indices := ValidatorSet(indexedAttestation.CustodyBit1Indices)

	// To be removed in Phase 1.
	if len(bit1Indices) != 0 {
		return errors.New("validators cannot have a custody bit set to 1 during phase 0")
	}

	// Verify max number of indices
	if count := len(bit1Indices) + len(bit0Indices); count > MAX_VALIDATORS_PER_COMMITTEE {
		return fmt.Errorf("invalid indices count in indexed attestation: %d", count)
	}

	// The indices must be sorted
	if !sort.IsSorted(bit0Indices) {
		return errors.New("custody bit 0 indices are not sorted")
	}

	if !sort.IsSorted(bit1Indices) {
		return errors.New("custody bit 1 indices are not sorted")
	}

	// Verify index sets are disjoint
	if bit0Indices.Intersects(bit1Indices) {
		return errors.New("validator set for custody bit 1 intersects with validator set for custody bit 0")
	}

	// Check the last item of the sorted list to be a valid index,
	// if this one is valid, the others are as well.
	if len(bit0Indices) > 0 && !state.Validators.IsValidatorIndex(bit0Indices[len(bit0Indices)-1]) {
		return errors.New("index in custody bit 1 indices is invalid")
	}

	if len(bit1Indices) > 0 && !state.Validators.IsValidatorIndex(bit1Indices[len(bit1Indices)-1]) {
		return errors.New("index in custody bit 1 indices is invalid")
	}

	custodyBit0Pubkeys := make([]BLSPubkey, 0)
	for _, i := range bit0Indices {
		custodyBit0Pubkeys = append(custodyBit0Pubkeys, state.Validators[i].Pubkey)
	}
	custodyBit1Pubkeys := make([]BLSPubkey, 0)
	for _, i := range bit1Indices {
		custodyBit1Pubkeys = append(custodyBit1Pubkeys, state.Validators[i].Pubkey)
	}

	// don't trust, verify
	if bls.BlsVerifyMultiple(
		[]BLSPubkey{
			bls.BlsAggregatePubkeys(custodyBit0Pubkeys),
			bls.BlsAggregatePubkeys(custodyBit1Pubkeys)},
		[]Root{
			ssz.HashTreeRoot(&AttestationDataAndCustodyBit{Data: indexedAttestation.Data, CustodyBit: false}, AttestationDataAndCustodyBitSSZ),
			ssz.HashTreeRoot(&AttestationDataAndCustodyBit{Data: indexedAttestation.Data, CustodyBit: true}, AttestationDataAndCustodyBitSSZ)},
		indexedAttestation.Signature,
		state.GetDomain(DOMAIN_ATTESTATION, indexedAttestation.Data.Target.Epoch),
	) {
		return nil
	}

	return errors.New("could not verify BLS signature for indexed attestation")
}
