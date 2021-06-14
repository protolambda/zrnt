package sharding

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

func ProcessParticipationRecordUpdates(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Rotate current/previous epoch attestations
	prevAtts, err := state.PreviousEpochAttestations()
	if err != nil {
		return err
	}
	currAtts, err := state.CurrentEpochAttestations()
	if err != nil {
		return err
	}
	if err := prevAtts.SetBacking(currAtts.Backing()); err != nil {
		return err
	}
	if err := currAtts.SetBacking(PendingAttestationsType(spec).DefaultNode()); err != nil {
		return err
	}

	return nil
}
