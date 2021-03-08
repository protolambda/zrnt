package gossipval

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
)

type VoluntaryExitValBackend interface {
	Spec
	HeadInfo
	// Checks if a valid exit for the given validator has been seen before.
	SeenExit(index beacon.ValidatorIndex) bool
}

func ValidateVoluntaryExit(ctx context.Context, volExit *beacon.SignedVoluntaryExit, exitVal VoluntaryExitValBackend) GossipValidatorResult {
	// [IGNORE] The voluntary exit is the first valid voluntary exit received
	// for the validator with index signed_voluntary_exit.message.validator_index
	if exitVal.SeenExit(volExit.Message.ValidatorIndex) {
		return GossipValidatorResult{IGNORE, fmt.Errorf("already seen exit for validator %d", volExit.Message.ValidatorIndex)}
	}

	// REJECT] All of the conditions within process_voluntary_exit pass validation.
	_, epc, state, err := exitVal.HeadInfo(ctx)
	if err != nil {
		return GossipValidatorResult{IGNORE, err}
	}
	if err := exitVal.Spec().ValidateVoluntaryExit(epc, state, volExit); err != nil {
		return GossipValidatorResult{REJECT, err}
	}

	return GossipValidatorResult{ACCEPT, nil}
}
