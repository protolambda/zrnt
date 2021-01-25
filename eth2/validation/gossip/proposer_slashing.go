package gossip

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
)

type ProposerSlashableIndexSeenFn func(proposer beacon.ValidatorIndex) bool

func (gv *GossipValidator) ValidateProposerSlashing(ctx context.Context, propSl *beacon.ProposerSlashing, seenFn ProposerSlashableIndexSeenFn) GossipValidatorResult {
	// [REJECT] All of the conditions within process_proposer_slashing pass validation.
	// Part 1: everything except the signature and the more exact "is slashable" check
	if err := gv.Spec.ValidateProposerSlashingNoSignature(propSl); err != nil {
		return GossipValidatorResult{REJECT, err}
	}
	// [IGNORE] The proposer slashing is the first valid proposer slashing received for the proposer with index
	proposer := propSl.SignedHeader1.Message.ProposerIndex
	if seenFn(proposer) {
		return GossipValidatorResult{IGNORE, fmt.Errorf("already seen proposer %d slashing", proposer)}
	}

	// [REJECT] All of the conditions within process_proposer_slashing pass validation.
	// Part 2: now the full check
	headRef, err := gv.Chain.Head()
	if err != nil {
		return GossipValidatorResult{IGNORE, errors.New("could not fetch head ref for validation")}
	}
	epc, err := headRef.EpochsContext(ctx)
	if err != nil {
		return GossipValidatorResult{IGNORE, errors.New("could not fetch head EPC for validation")}
	}
	state, err := headRef.State(ctx)
	if err != nil {
		return GossipValidatorResult{IGNORE, errors.New("could not fetch head state for validation")}
	}
	if err := gv.Spec.ValidateProposerSlashing(epc, state, propSl); err != nil {
		return GossipValidatorResult{REJECT, err}
	}
	return GossipValidatorResult{ACCEPT, nil}
}
