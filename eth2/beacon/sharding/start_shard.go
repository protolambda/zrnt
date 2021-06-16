package sharding

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

func ProcessShardEpochIncrement(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	slot, err := state.Slot()
	if err != nil {
		return err
	}
	nextStartShard, err := epc.StartShard(slot + 1)
	if err != nil {
		return err
	}
	return state.SetCurrentEpochStartShard(nextStartShard)
}
