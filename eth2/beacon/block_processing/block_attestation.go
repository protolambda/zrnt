package block_processing

import (
	"errors"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
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
	justifiedEpoch := state.PreviousJustifiedEpoch
	if (attestation.Data.Slot + 1).ToEpoch() >= state.Epoch() {
		justifiedEpoch = state.CurrentJustifiedEpoch
	}
	blockRoot, blockRootErr := state.GetBlockRoot(attestation.Data.SourceEpoch.GetStartSlot())
	if !(attestation.Data.Slot >= beacon.GENESIS_SLOT && attestation.Data.Slot+beacon.MIN_ATTESTATION_INCLUSION_DELAY <= state.Slot &&
		state.Slot < attestation.Data.Slot+beacon.SLOTS_PER_EPOCH && attestation.Data.SourceEpoch == justifiedEpoch &&
		(blockRootErr == nil && attestation.Data.SourceRoot == blockRoot) &&
		(state.LatestCrosslinks[attestation.Data.Shard] == attestation.Data.PreviousCrosslink ||
			state.LatestCrosslinks[attestation.Data.Shard] == beacon.Crosslink{CrosslinkDataRoot: attestation.Data.CrosslinkDataRoot, Epoch: attestation.Data.Slot.ToEpoch()})) {
		return errors.New("attestation is not valid")
	}
	// Verify bitfields and aggregate signature
	// custody bitfield is phase 0 only:
	if attestation.AggregationBitfield.IsZero() || !attestation.CustodyBitfield.IsZero() {
		return errors.New("attestation %d has incorrect bitfield(s)")
	}

	crosslinkCommittees := state.GetCrosslinkCommitteesAtSlot(attestation.Data.Slot, false)
	var crosslinkCommittee *beacon.CrosslinkCommittee
	for i := 0; i < len(crosslinkCommittees); i++ {
		committee := &crosslinkCommittees[i]
		if committee.Shard == attestation.Data.Shard {
			crosslinkCommittee = committee
			break
		}
	}
	// TODO spec is weak here: it's not very explicit about length of bitfields.
	//  Let's just make sure they are the size of the committee
	if !attestation.AggregationBitfield.VerifySize(uint64(len(crosslinkCommittee.Committee))) ||
		!attestation.CustodyBitfield.VerifySize(uint64(len(crosslinkCommittee.Committee))) {
		return errors.New("attestation %d has bitfield(s) with incorrect size")
	}
	// phase 0 only
	if !attestation.AggregationBitfield.IsZero() || !attestation.CustodyBitfield.IsZero() {
		return errors.New("attestation %d has non-zero bitfield(s)")
	}

	participants, err := state.GetAttestationParticipants(&attestation.Data, &attestation.AggregationBitfield)
	if err != nil {
		return errors.New("participants could not be derived from aggregation_bitfield")
	}
	custodyBit1_participants, err := state.GetAttestationParticipants(&attestation.Data, &attestation.CustodyBitfield)
	if err != nil {
		return errors.New("participants could not be derived from custody_bitfield")
	}
	_, custodyBit0_participants := beacon.FindInAndOutValidators(participants, custodyBit1_participants)

	// get lists of pubkeys for both 0 and 1 custody-bits
	custodyBit0_pubkeys := make([]beacon.BLSPubkey, len(custodyBit0_participants))
	for i, v := range custodyBit0_participants {
		custodyBit0_pubkeys[i] = state.ValidatorRegistry[v].Pubkey
	}
	custodyBit1_pubkeys := make([]beacon.BLSPubkey, len(custodyBit1_participants))
	for i, v := range custodyBit1_participants {
		custodyBit1_pubkeys[i] = state.ValidatorRegistry[v].Pubkey
	}
	// aggregate each of the two lists
	pubKeys := []beacon.BLSPubkey{bls.BlsAggregatePubkeys(custodyBit0_pubkeys), bls.BlsAggregatePubkeys(custodyBit1_pubkeys)}
	// hash the attestation data with 0 and 1 as bit
	hashes := []beacon.Root{
		ssz.HashTreeRoot(beacon.AttestationDataAndCustodyBit{attestation.Data, false}),
		ssz.HashTreeRoot(beacon.AttestationDataAndCustodyBit{attestation.Data, true}),
	}
	// now verify the two
	if !bls.BlsVerifyMultiple(pubKeys, hashes, attestation.AggregateSignature,
		beacon.GetDomain(state.Fork, attestation.Data.Slot.ToEpoch(), beacon.DOMAIN_ATTESTATION)) {
		return errors.New("attestation has invalid aggregated BLS signature")
	}

	// phase 0 only:
	if attestation.Data.CrosslinkDataRoot != (beacon.Root{}) {
		return errors.New("attestation has invalid crosslink: root must be 0 in phase 0")
	}

	// Apply the attestation
	pendingAttestation := beacon.PendingAttestation{
		Data:                attestation.Data,
		AggregationBitfield: attestation.AggregationBitfield,
		CustodyBitfield:     attestation.CustodyBitfield,
		InclusionSlot:       state.Slot,
	}
	if attEpoch := attestation.Data.Slot.ToEpoch(); attEpoch == state.Epoch() {
		state.CurrentEpochAttestations = append(state.CurrentEpochAttestations, pendingAttestation)
	} else if attEpoch == state.PreviousEpoch() {
		state.PreviousEpochAttestations = append(state.CurrentEpochAttestations, pendingAttestation)
	}
	return nil
}
