package transition

import (
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

// Verify validity of slashable_attestation fields.
func Verify_slashable_attestation(state *beacon.BeaconState, slashable_attestation *beacon.SlashableAttestation) bool {
	if size := len(slashable_attestation.Validator_indices); size == 0 || size > eth2.MAX_INDICES_PER_SLASHABLE_VOTE {
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
	if !Is_validator_index(state, slashable_attestation.Validator_indices[end]) {
		return false
	}

	custody_bit_0_pubkeys := make([]eth2.BLSPubkey, 0)
	custody_bit_1_pubkeys := make([]eth2.BLSPubkey, 0)

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
		[]eth2.BLSPubkey{
			bls.Bls_aggregate_pubkeys(custody_bit_0_pubkeys),
			bls.Bls_aggregate_pubkeys(custody_bit_1_pubkeys)},
		[]eth2.Root{
			ssz.Hash_tree_root(beacon.AttestationDataAndCustodyBit{Data: slashable_attestation.Data, Custody_bit: false}),
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
