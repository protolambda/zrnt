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
func VerifySlashableAttestation(state *beacon.BeaconState, slashableAttestation *beacon.SlashableAttestation) bool {
	if size := len(slashableAttestation.ValidatorIndices); size == 0 || size > beacon.MAX_INDICES_PER_SLASHABLE_VOTE {
		return false
	}
	// [TO BE REMOVED IN PHASE 1]
	if !slashableAttestation.CustodyBitfield.IsZero() {
		return false
	}
	// verify the size of the bitfield: it must have exactly enough bits for the given amount of validators.
	if !slashableAttestation.CustodyBitfield.VerifySize(uint64(len(slashableAttestation.ValidatorIndices))) {
		return false
	}

	// simple check if the list is sorted.
	end := len(slashableAttestation.ValidatorIndices) - 1
	for i := 0; i < end; i++ {
		if slashableAttestation.ValidatorIndices[i] >= slashableAttestation.ValidatorIndices[i+1] {
			return false
		}
	}
	// Check the last item of the sorted list
	if !state.ValidatorRegistry.IsValidatorIndex(slashableAttestation.ValidatorIndices[end]) {
		return false
	}

	custodyBit0_pubkeys := make([]beacon.BLSPubkey, 0)
	custodyBit1_pubkeys := make([]beacon.BLSPubkey, 0)

	for i, validatorIndex := range slashableAttestation.ValidatorIndices {
		// Update spec, or is this acceptable? (the bitfield verify size doesn't suffice here)
		if slashableAttestation.CustodyBitfield.GetBit(uint64(i)) == 0 {
			custodyBit0_pubkeys = append(custodyBit0_pubkeys, state.ValidatorRegistry[validatorIndex].Pubkey)
		} else {
			custodyBit1_pubkeys = append(custodyBit1_pubkeys, state.ValidatorRegistry[validatorIndex].Pubkey)
		}
	}
	// don't trust, verify
	return bls.BlsVerifyMultiple(
		[]beacon.BLSPubkey{
			bls.BlsAggregatePubkeys(custodyBit0_pubkeys),
			bls.BlsAggregatePubkeys(custodyBit1_pubkeys)},
		[]beacon.Root{
			ssz.HashTreeRoot(beacon.AttestationDataAndCustodyBit{Data: slashableAttestation.Data, CustodyBit: false}),
			ssz.HashTreeRoot(beacon.AttestationDataAndCustodyBit{Data: slashableAttestation.Data, CustodyBit: true})},
		slashableAttestation.AggregateSignature,
		beacon.GetDomain(state.Fork, slashableAttestation.Data.Slot.ToEpoch(), beacon.DOMAIN_ATTESTATION),
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
