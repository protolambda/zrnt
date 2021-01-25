package gossip

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
)

// Check if any of the indices has NOT been seen before, return true if so.
// May not be an index within valid range.
// It is up to the topic subscriber to mark indices as seen. Indices which are checked may not be valid,
// and should not be marked as seen because of just the check itself.
// It is recommended to regard any indices which were finalized as slashed, as seen.
type AttesterSlashedAnySeenFn func(indices []beacon.ValidatorIndex) bool

func (gv *GossipValidator) ValidateAttesterSlashing(ctx context.Context, attSl *beacon.AttesterSlashing, seenFn AttesterSlashedAnySeenFn) GossipValidatorResult {
	sa1 := &attSl.Attestation1
	sa2 := &attSl.Attestation2

	// [REJECT] All of the conditions within process_attester_slashing pass validation.
	// Part 1: just light checks, make sure the formatting is right, no signature checks yet.
	if !beacon.IsSlashableAttestationData(&sa1.Data, &sa2.Data) {
		return GossipValidatorResult{REJECT, errors.New("attester slashing has no valid reasoning")}
	}
	indices1, err := beacon.ValidateIndexedAttestationIndicesSet(sa1)
	if err != nil {
		return GossipValidatorResult{REJECT, errors.New("attestation 1 of attester slashing cannot be verified")}
	}
	indices2, err := beacon.ValidateIndexedAttestationIndicesSet(sa2)
	if err != nil {
		return GossipValidatorResult{REJECT, errors.New("attestation 2 of attester slashing cannot be verified")}
	}

	// [IGNORE] At least one index in the intersection of the attesting indices of each attestation has not yet been seen in any prior attester_slashing
	slashable := make([]beacon.ValidatorIndex, 0, len(indices1))
	indices1.ZigZagJoin(indices2, func(i beacon.ValidatorIndex) {
		slashable = append(slashable, i)
	}, nil)

	if !seenFn(slashable) {
		return GossipValidatorResult{IGNORE, errors.New("no unseen slashable attester indices")}
	}

	// [REJECT] All of the conditions within process_attester_slashing pass validation.
	// Part 2: signature checks
	dom1, err := gv.GetDomain(gv.Spec.DOMAIN_BEACON_ATTESTER, sa1.Data.Target.Epoch)
	if err != nil {
		return GossipValidatorResult{REJECT, fmt.Errorf("could not get attester BLS domain 1 for epoch %d: %v", sa1.Data.Target.Epoch, err)}
	}
	dom2, err := gv.GetDomain(gv.Spec.DOMAIN_BEACON_ATTESTER, sa2.Data.Target.Epoch)
	if err != nil {
		return GossipValidatorResult{REJECT, fmt.Errorf("could not get attester BLS domain 2 for epoch %d: %v", sa2.Data.Target.Epoch, err)}
	}

	// Note: the unrecognized indices in here are thrown away.
	// TODO: get pubkey cache via EPC or not?
	if err := beacon.ValidateIndexedAttestationSignature(dom1, nil, sa1); err != nil {
		return GossipValidatorResult{REJECT, fmt.Errorf("attester slashing att 1 signature is invalid: %v", err)}
	}
	if err := beacon.ValidateIndexedAttestationSignature(dom2, nil, sa2); err != nil {
		return GossipValidatorResult{REJECT, fmt.Errorf("attester slashing att 2 signature is invalid: %v", err)}
	}

	return GossipValidatorResult{ACCEPT, nil}
}
