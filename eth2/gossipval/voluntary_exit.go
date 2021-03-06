package gossipval

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
)

// Checks if a valid exit for the given validator has been seen before.
type VoluntaryExitSeenFn func(index beacon.ValidatorIndex) bool

func (gv *GossipValidator) ValidateVoluntaryExit(ctx context.Context, volExit *beacon.SignedVoluntaryExit, seenFn VoluntaryExitSeenFn) GossipValidatorResult {
	// [IGNORE] The voluntary exit is the first valid voluntary exit received
	// for the validator with index signed_voluntary_exit.message.validator_index
	if seenFn(volExit.Message.ValidatorIndex) {
		return GossipValidatorResult{IGNORE, fmt.Errorf("already seen exit for validator %d", volExit.Message.ValidatorIndex)}
	}

	// REJECT] All of the conditions within process_voluntary_exit pass validation.
	_, epc, state, err := gv.HeadInfo(ctx)
	if err != nil {
		return GossipValidatorResult{IGNORE, err}
	}
	if err := gv.Spec.ValidateVoluntaryExit(epc, state, volExit); err != nil {
		return GossipValidatorResult{REJECT, err}
	}

	return GossipValidatorResult{ACCEPT, nil}
}
