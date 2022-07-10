package gossipval

import (
	"context"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
)

type ProposerSlashingValBackend interface {
	Spec
	HeadInfo
	SeenProposerSlashing(proposer common.ValidatorIndex) bool
	MarkProposerSlashing(index common.ValidatorIndex)
}

func ValidateProposerSlashing(ctx context.Context, propSl *phase0.ProposerSlashing, propSlVal ProposerSlashingValBackend) GossipValidatorResult {
	spec := propSlVal.Spec()
	// [REJECT] All of the conditions within process_proposer_slashing pass validation.
	// Part 1: everything except the signature and the more exact "is slashable" check
	if err := phase0.ValidateProposerSlashingNoSignature(spec, propSl); err != nil {
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
	if err := phase0.ValidateProposerSlashing(spec, epc, state, propSl); err != nil {
		return GossipValidatorResult{REJECT, err}
	}
	propSlVal.MarkProposerSlashing(proposer)
	return GossipValidatorResult{ACCEPT, nil}
}
