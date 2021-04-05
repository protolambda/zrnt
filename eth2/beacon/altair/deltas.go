package altair

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

func ProcessEpochRewardsAndPenalties(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, attesterData *EpochAttesterData, state *BeaconStateView) error {
	// TODO source, target, head flag deltas
	// TODO inactivity deltas

	return nil
}
