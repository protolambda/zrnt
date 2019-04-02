package block_processing

import (
	"errors"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessBlockAttesterSlashings(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.AttesterSlashings) > beacon.MAX_ATTESTER_SLASHINGS {
		return errors.New("too many attester slashings")
	}
	for _, attesterSlashing := range block.Body.AttesterSlashings {
		if err := ProcessAttesterSlashing(state, &attesterSlashing); err != nil {
			return err
		}
	}
	return nil
}

func ProcessAttesterSlashing(state *beacon.BeaconState, attesterSlashing *beacon.AttesterSlashing) error {
	sa1, sa2 := &attesterSlashing.SlashableAttestation1, &attesterSlashing.SlashableAttestation2
	// verify the attester_slashing
	if !(sa1.Data != sa2.Data && (
		IsDoubleVote(&sa1.Data, &sa2.Data) ||
			IsSurroundVote(&sa1.Data, &sa2.Data)) &&
		VerifySlashableAttestation(state, sa1) &&
		VerifySlashableAttestation(state, sa2)) {
		return errors.New("attester slashing is invalid")
	}
	// keep track of effectiveness
	slashedAny := false
	// run slashings where applicable

	// indices are trusted, they have been verified by verify_slashable_attestation(...)
	for _, v1 := range sa1.ValidatorIndices {
		for _, v2 := range sa2.ValidatorIndices {
			if v1 == v2 && !state.ValidatorRegistry[v1].Slashed {
				if err := state.SlashValidator(v1); err != nil {
					return err
				}
				slashedAny = true
				continue
			}
		}
	}
	// "Verify that len(slashable_indices) >= 1."
	if !slashedAny {
		return errors.New("attester slashing %d is not effective, hence invalid")
	}
	return nil
}

// Verify validity of slashable_attestation fields.
func VerifyIndexedAttestation(state *beacon.BeaconState, indexedAttestation *beacon.IndexedAttestation) bool {
	custodyBit0Indices := indexedAttestation.CustodyBit0Indexes
	custodyBit1Indices := indexedAttestation.CustodyBit1Indexes

	// [TO BE REMOVED IN PHASE 1]
	if len(custodyBit1Indices) != 0 {
		return false
	}

	totalAttestingIndices := len(custodyBit1Indices) + len(custodyBit0Indices)
	if !(1 <= totalAttestingIndices && totalAttestingIndices <= beacon.MAX_ATTESTATION_PARTICIPANTS) {
		return false
	}

	// simple check if the lists are sorted.
	verifyAttestIndexList := func (indices []beacon.ValidatorIndex) bool {
		end := len(indices) - 1
		for i := 0; i < end; i++ {
			if indices[i] >= indices[i+1] {
				return false
			}
		}

		// Check the last item of the sorted list to be a valid index
		if !state.ValidatorRegistry.IsValidatorIndex(indices[end]) {
			return false
		}
		return true
	}
	if !verifyAttestIndexList(custodyBit0Indices) || !verifyAttestIndexList(custodyBit1Indices) {
		return false
	}

	custodyBit0_pubkeys := make([]beacon.BLSPubkey, 0)
	for _, i := range custodyBit0Indices {
		custodyBit0_pubkeys = append(custodyBit0_pubkeys, state.ValidatorRegistry[i].Pubkey)
	}
	custodyBit1_pubkeys := make([]beacon.BLSPubkey, 0)
	for _, i := range custodyBit1Indices {
		custodyBit1_pubkeys = append(custodyBit1_pubkeys, state.ValidatorRegistry[i].Pubkey)
	}

	// don't trust, verify
	return bls.BlsVerifyMultiple(
		[]beacon.BLSPubkey{
			bls.BlsAggregatePubkeys(custodyBit0_pubkeys),
			bls.BlsAggregatePubkeys(custodyBit1_pubkeys)},
		[]beacon.Root{
			ssz.HashTreeRoot(beacon.AttestationDataAndCustodyBit{Data: indexedAttestation.Data, CustodyBit: false}),
			ssz.HashTreeRoot(beacon.AttestationDataAndCustodyBit{Data: indexedAttestation.Data, CustodyBit: true})},
		indexedAttestation.AggregateSignature,
		beacon.GetDomain(state.Fork, indexedAttestation.Data.Slot.ToEpoch(), beacon.DOMAIN_ATTESTATION),
	)
}

// Check if a and b have the same target epoch.
func IsDoubleVote(a *beacon.AttestationData, b *beacon.AttestationData) bool {
	return a.Slot.ToEpoch() == b.Slot.ToEpoch()
}

// Check if a surrounds b, i.E. source(a) < source(b) and target(a) > target(b)
func IsSurroundVote(a *beacon.AttestationData, b *beacon.AttestationData) bool {
	return a.SourceEpoch < b.SourceEpoch && a.Slot.ToEpoch() > b.Slot.ToEpoch()
}
