package gossipval

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
)

type ProposerSlashingValBackend interface {
	Spec
	HeadInfo
	SeenProposerSlashing(proposer beacon.ValidatorIndex) bool
}

func ValidateProposerSlashing(ctx context.Context, propSl *beacon.ProposerSlashing, propSlVal ProposerSlashingValBackend) GossipValidatorResult {
	spec := propSlVal.Spec()
	// [REJECT] All of the conditions within process_proposer_slashing pass validation.
	// Part 1: everything except the signature and the more exact "is slashable" check
	if err := spec.ValidateProposerSlashingNoSignature(propSl); err != nil {
		return GossipValidatorResult{REJECT, err}
	}
	// [IGNORE] The proposer slashing is the first valid proposer slashing received for the proposer with index
	proposer := propSl.SignedHeader1.Message.ProposerIndex
	if propSlVal.SeenProposerSlashing(proposer) {
		return GossipValidatorResult{IGNORE, fmt.Errorf("already seen proposer %d slashing", proposer)}
	}

	// [REJECT] All of the conditions within process_proposer_slashing pass validation.
	// Part 2: now the full check
	_, epc, state, err := propSlVal.HeadInfo(ctx)
	if err != nil {
		return GossipValidatorResult{IGNORE, err}
	}
	if err := spec.ValidateProposerSlashing(epc, state, propSl); err != nil {
		return GossipValidatorResult{REJECT, err}
	}
	return GossipValidatorResult{ACCEPT, nil}
}
