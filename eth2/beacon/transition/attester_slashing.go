package transition

import (
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

// Verify validity of slashable_attestation fields.
func Verify_slashable_attestation(state *beacon.BeaconState, slashable_attestation *beacon.SlashableAttestation) bool {
	// TODO Moved condition to top, compared to spec. Data can be way too big, get rid of that ASAP.
	if len(slashable_attestation.Validator_indices) == 0 ||
		len(slashable_attestation.Validator_indices) > eth2.MAX_INDICES_PER_SLASHABLE_VOTE ||
	// [TO BE REMOVED IN PHASE 1]
		!slashable_attestation.Custody_bitfield.IsZero() ||
	// verify the size of the bitfield: it must have exactly enough bits for the given amount of validators.
		!slashable_attestation.Custody_bitfield.VerifySize(uint64(len(slashable_attestation.Validator_indices))) {
		return false
	}

	// simple check if the list is sorted.
	for i := 0; i < len(slashable_attestation.Validator_indices)-1; i++ {
		if slashable_attestation.Validator_indices[i] >= slashable_attestation.Validator_indices[i+1] {
			return false
		}
	}

	custody_bit_0_pubkeys, custody_bit_1_pubkeys := make([]eth2.BLSPubkey, 0), make([]eth2.BLSPubkey, 0)
	for i, validator_index := range slashable_attestation.Validator_indices {
		// The slashable indices is one giant sorted list of numbers,
		//   bigger than the registry, causing a out-of-bounds panic for some of the indices.
		if !Is_validator_index(state, validator_index) {
			return false
		}
		// Update spec, or is this acceptable? (the bitfield verify size doesn't suffice here)
		if slashable_attestation.Custody_bitfield.GetBit(uint64(i)) == 0 {
			custody_bit_0_pubkeys = append(custody_bit_0_pubkeys, state.Validator_registry[validator_index].Pubkey)
		} else {
			custody_bit_1_pubkeys = append(custody_bit_1_pubkeys, state.Validator_registry[validator_index].Pubkey)
		}
	}
	// don't trust, verify
	return bls.Bls_verify_multiple(
		[]eth2.BLSPubkey{bls.Bls_aggregate_pubkeys(custody_bit_0_pubkeys), bls.Bls_aggregate_pubkeys(custody_bit_1_pubkeys)},
		[]eth2.Root{ssz.Hash_tree_root(beacon.AttestationDataAndCustodyBit{Data: slashable_attestation.Data, Custody_bit: false}),
			ssz.Hash_tree_root(beacon.AttestationDataAndCustodyBit{Data: slashable_attestation.Data, Custody_bit: true})},
		slashable_attestation.Aggregate_signature,
		Get_domain(state.Fork, slashable_attestation.Data.Slot.ToEpoch(), eth2.DOMAIN_ATTESTATION),
	)
}

// Check if a and b have the same target epoch. //TODO: spec has wrong wording here (?)
func Is_double_vote(a *beacon.AttestationData, b *beacon.AttestationData) bool {
	return a.Slot.ToEpoch() == b.Slot.ToEpoch()
}

// Check if a surrounds b, i.E. source(a) < source(b) and target(a) > target(b)
func Is_surround_vote(a *beacon.AttestationData, b *beacon.AttestationData) bool {
	return a.Justified_epoch < b.Justified_epoch && a.Slot.ToEpoch() > b.Slot.ToEpoch()
}
