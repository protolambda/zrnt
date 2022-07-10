package gossipval

import (
	"context"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
)

type VoluntaryExitValBackend interface {
	Spec
	HeadInfo
	// Checks if a valid exit for the given validator has been seen before.
	SeenExit(index common.ValidatorIndex) bool
	// Marks exit as seen
	MarkExit(index common.ValidatorIndex)
}

func ValidateVoluntaryExit(ctx context.Context, volExit *phase0.SignedVoluntaryExit, exitVal VoluntaryExitValBackend) GossipValidatorResult {
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
	if err := phase0.ValidateVoluntaryExit(exitVal.Spec(), epc, state, volExit); err != nil {
		return GossipValidatorResult{REJECT, err}
	}

	exitVal.MarkExit(volExit.Message.ValidatorIndex)

	return GossipValidatorResult{ACCEPT, nil}
}
