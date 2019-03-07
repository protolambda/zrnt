package attestations

import (
	"errors"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func ProcessBlockAttestations(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.Attestations) > beacon.MAX_ATTESTATIONS {
		return errors.New("too many attestations")
	}
	for _, attestation := range block.Body.Attestations {
		if err := ProcessAttestation(state, &attestation); err != nil {
			return err
		}
	}
	return nil
}

func ProcessAttestation(state *beacon.BeaconState, attestation *beacon.Attestation) error {
	justified_epoch := state.Previous_justified_epoch
	if (attestation.Data.Slot + 1).ToEpoch() >= state.Epoch() {
		justified_epoch = state.Justified_epoch
	}
	blockRoot, blockRootErr := state.Get_block_root(attestation.Data.Justified_epoch.GetStartSlot())
	if !(attestation.Data.Slot >= beacon.GENESIS_SLOT && attestation.Data.Slot+beacon.MIN_ATTESTATION_INCLUSION_DELAY <= state.Slot &&
		state.Slot < attestation.Data.Slot+beacon.SLOTS_PER_EPOCH && attestation.Data.Justified_epoch == justified_epoch &&
		(blockRootErr == nil && attestation.Data.Justified_block_root == blockRoot) &&
		(state.Latest_crosslinks[attestation.Data.Shard] == attestation.Data.Latest_crosslink ||
			state.Latest_crosslinks[attestation.Data.Shard] == beacon.Crosslink{Crosslink_data_root: attestation.Data.Crosslink_data_root, Epoch: attestation.Data.Slot.ToEpoch()})) {
		return errors.New("attestation is not valid")
	}
	// Verify bitfields and aggregate signature
	// custody bitfield is phase 0 only:
	if attestation.Aggregation_bitfield.IsZero() || !attestation.Custody_bitfield.IsZero() {
		return errors.New("attestation %d has incorrect bitfield(s)")
	}

	crosslink_committees, err := state.Get_crosslink_committees_at_slot(attestation.Data.Slot, false)
	if err != nil {
		return err
	}
	crosslink_committee := beacon.CrosslinkCommittee{}
	for _, committee := range crosslink_committees {
		if committee.Shard == attestation.Data.Shard {
			crosslink_committee = committee
			break
		}
	}
	// TODO spec is weak here: it's not very explicit about length of bitfields.
	//  Let's just make sure they are the size of the committee
	if !attestation.Aggregation_bitfield.VerifySize(uint64(len(crosslink_committee.Committee))) ||
		!attestation.Custody_bitfield.VerifySize(uint64(len(crosslink_committee.Committee))) {
		return errors.New("attestation %d has bitfield(s) with incorrect size")
	}
	// phase 0 only
	if !attestation.Aggregation_bitfield.IsZero() || !attestation.Custody_bitfield.IsZero() {
		return errors.New("attestation %d has non-zero bitfield(s)")
	}

	participants, err := state.Get_attestation_participants(&attestation.Data, &attestation.Aggregation_bitfield)
	if err != nil {
		return errors.New("participants could not be derived from aggregation_bitfield")
	}
	custody_bit_1_participants, err := state.Get_attestation_participants(&attestation.Data, &attestation.Custody_bitfield)
	if err != nil {
		return errors.New("participants could not be derived from custody_bitfield")
	}
	custody_bit_0_participants := participants.Minus(custody_bit_1_participants)

	// get lists of pubkeys for both 0 and 1 custody-bits
	custody_bit_0_pubkeys := make([]beacon.BLSPubkey, len(custody_bit_0_participants))
	for i, v := range custody_bit_0_participants {
		custody_bit_0_pubkeys[i] = state.Validator_registry[v].Pubkey
	}
	custody_bit_1_pubkeys := make([]beacon.BLSPubkey, len(custody_bit_1_participants))
	for i, v := range custody_bit_1_participants {
		custody_bit_1_pubkeys[i] = state.Validator_registry[v].Pubkey
	}
	// aggregate each of the two lists
	pubKeys := []beacon.BLSPubkey{bls.Bls_aggregate_pubkeys(custody_bit_0_pubkeys), bls.Bls_aggregate_pubkeys(custody_bit_1_pubkeys)}
	// hash the attestation data with 0 and 1 as bit
	hashes := []beacon.Root{
		ssz.Hash_tree_root(beacon.AttestationDataAndCustodyBit{attestation.Data, false}),
		ssz.Hash_tree_root(beacon.AttestationDataAndCustodyBit{attestation.Data, true}),
	}
	// now verify the two
	if !bls.Bls_verify_multiple(pubKeys, hashes, attestation.Aggregate_signature,
		beacon.Get_domain(state.Fork, attestation.Data.Slot.ToEpoch(), beacon.DOMAIN_ATTESTATION)) {
		return errors.New("attestation has invalid aggregated BLS signature")
	}

	// phase 0 only:
	if attestation.Data.Crosslink_data_root != (beacon.Root{}) {
		return errors.New("attestation has invalid crosslink: root must be 0 in phase 0")
	}
	return nil
}
