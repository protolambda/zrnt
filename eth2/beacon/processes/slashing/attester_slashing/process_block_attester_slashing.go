package attester_slashing

import (
	"errors"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/slashing"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func ProcessBlockAttesterSlashings(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.Attester_slashings) > beacon.MAX_ATTESTER_SLASHINGS {
		return errors.New("too many attester slashings")
	}
	for _, attester_slashing := range block.Body.Attester_slashings {
		if err := ProcessAttesterSlashing(state, &attester_slashing); err != nil {
			return err
		}
	}
}

func ProcessAttesterSlashing(state *beacon.BeaconState, attester_slashing *beacon.AttesterSlashing) error {
	sa1, sa2 := &attester_slashing.Slashable_attestation_1, &attester_slashing.Slashable_attestation_2
	// verify the attester_slashing
	if !(sa1.Data != sa2.Data && (
		Is_double_vote(&sa1.Data, &sa2.Data) ||
			Is_surround_vote(&sa1.Data, &sa2.Data)) &&
		Verify_slashable_attestation(state, sa1) &&
		Verify_slashable_attestation(state, sa2)) {
		return errors.New("attester slashing is invalid")
	}
	// keep track of effectiveness
	slashedAny := false
	// run slashings where applicable

	// indices are trusted, they have been verified by verify_slashable_attestation(...)
	for _, v1 := range sa1.Validator_indices {
		for _, v2 := range sa2.Validator_indices {
			if v1 == v2 && !state.Validator_registry[v1].Slashed {
				if err := slashing.Slash_validator(state, v1); err != nil {
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
}

// Verify validity of slashable_attestation fields.
func Verify_slashable_attestation(state *beacon.BeaconState, slashable_attestation *beacon.SlashableAttestation) bool {
	if size := len(slashable_attestation.Validator_indices); size == 0 || size > beacon.MAX_INDICES_PER_SLASHABLE_VOTE {
		return false
	}
	// [TO BE REMOVED IN PHASE 1]
	if !slashable_attestation.Custody_bitfield.IsZero() {
		return false
	}
	// verify the size of the bitfield: it must have exactly enough bits for the given amount of validators.
	if !slashable_attestation.Custody_bitfield.VerifySize(uint64(len(slashable_attestation.Validator_indices))) {
		return false
	}

	// simple check if the list is sorted.
	end := len(slashable_attestation.Validator_indices) - 1
	for i := 0; i < end; i++ {
		if slashable_attestation.Validator_indices[i] >= slashable_attestation.Validator_indices[i+1] {
			return false
		}
	}
	// Check the last item of the sorted list
	if !state.Validator_registry.Is_validator_index(slashable_attestation.Validator_indices[end]) {
		return false
	}

	custody_bit_0_pubkeys := make([]beacon.BLSPubkey, 0)
	custody_bit_1_pubkeys := make([]beacon.BLSPubkey, 0)

	for i, validator_index := range slashable_attestation.Validator_indices {
		// Update spec, or is this acceptable? (the bitfield verify size doesn't suffice here)
		if slashable_attestation.Custody_bitfield.GetBit(uint64(i)) == 0 {
			custody_bit_0_pubkeys = append(custody_bit_0_pubkeys, state.Validator_registry[validator_index].Pubkey)
		} else {
			custody_bit_1_pubkeys = append(custody_bit_1_pubkeys, state.Validator_registry[validator_index].Pubkey)
		}
	}
	// don't trust, verify
	return bls.Bls_verify_multiple(
		[]beacon.BLSPubkey{
			bls.Bls_aggregate_pubkeys(custody_bit_0_pubkeys),
			bls.Bls_aggregate_pubkeys(custody_bit_1_pubkeys)},
		[]beacon.Root{
			ssz.Hash_tree_root(beacon.AttestationDataAndCustodyBit{Data: slashable_attestation.Data, Custody_bit: false}),
			ssz.Hash_tree_root(beacon.AttestationDataAndCustodyBit{Data: slashable_attestation.Data, Custody_bit: true})},
		slashable_attestation.Aggregate_signature,
		beacon.Get_domain(state.Fork, slashable_attestation.Data.Slot.ToEpoch(), beacon.DOMAIN_ATTESTATION),
	)
}

// Check if a and b have the same target epoch.
func Is_double_vote(a *beacon.AttestationData, b *beacon.AttestationData) bool {
	return a.Slot.ToEpoch() == b.Slot.ToEpoch()
}

// Check if a surrounds b, i.E. source(a) < source(b) and target(a) > target(b)
func Is_surround_vote(a *beacon.AttestationData, b *beacon.AttestationData) bool {
	return a.Justified_epoch < b.Justified_epoch && a.Slot.ToEpoch() > b.Slot.ToEpoch()
}
