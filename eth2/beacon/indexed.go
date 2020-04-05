package beacon

import (
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/meta"

	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	. "github.com/protolambda/ztyp/view"
	"sort"
)

type CommitteeIndices []ValidatorIndex
var CommitteeIndicesType = ListType(ValidatorIndexType, MAX_VALIDATORS_PER_COMMITTEE)

func (ci *CommitteeIndices) Limit() uint64 {
	return MAX_VALIDATORS_PER_COMMITTEE
}

type IndexedAttestation struct {
	AttestingIndices CommitteeIndices
	Data             AttestationData
	Signature        BLSSignature
}

var IndexedAttestationType = ContainerType("IndexedAttestation", []FieldDef{
	{"attesting_indices", CommitteeIndicesType},
	{"data", AttestationDataType},
	{"signature", BLSSignatureType},
})

// Verify validity of slashable_attestation fields.
func (indexedAttestation *IndexedAttestation) Validate(m AttestationValidator) error {
	// wrap it in validator-sets. Does not sort it, but does make checking if it is a lot easier.
	indices := ValidatorSet(indexedAttestation.AttestingIndices)

	// Verify max number of indices
	if count := len(indices); count > MAX_VALIDATORS_PER_COMMITTEE {
		return fmt.Errorf("invalid indices count in indexed attestation: %d", count)
	}

	// The indices must be sorted
	if !sort.IsSorted(indices) {
		return errors.New("attestation indices are not sorted")
	}

	// Verify if the indices are unique. Simple O(n) check, since they are already sorted.
	for i := 1; i < len(indices); i++ {
		if indices[i-1] == indices[i] {
			return fmt.Errorf("attestation indices at %d and %d are duplicate, both: %d", i-1, i, indices[i])
		}
	}

	// Check the last item of the sorted list to be a valid index,
	// if this one is valid, the others are as well, since they are lower.
	if len(indices) > 0 {
		valid, err := m.IsValidIndex(indices[len(indices)-1])
		if err != nil {
			return err
		} else if !valid {
			return errors.New("attestation indices contains out of range index")
		}
	}

	pubkeys := make([]BLSPubkey, 0, 2)
	for _, i := range indices {
		pub, err := m.Pubkey(i)
		if err != nil {
			return err
		}
		pubkeys = append(pubkeys, pub)
	}

	// empty attestation
	if len(pubkeys) <= 0 {
		// TODO: check if the signature is default
		return nil
	}

	dom, err := m.GetDomain(DOMAIN_BEACON_ATTESTER, indexedAttestation.Data.Target.Epoch)
	if err != nil {
		return err
	}
	if !bls.FastAggregateVerify(pubkeys,
		ComputeSigningRoot(ssz.HashTreeRoot(&indexedAttestation.Data, AttestationDataSSZ), dom),
		indexedAttestation.Signature,
	) {
		return errors.New("could not verify BLS signature for indexed attestation")
	}

	return nil
}
