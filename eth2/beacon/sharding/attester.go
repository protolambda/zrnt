package sharding

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type EpochAttesterData struct {
	Altair altair.EpochAttesterData
}

func ComputeEpochAttesterData(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, flats []common.FlatValidator, state *BeaconStateView) (out *EpochAttesterData, err error) {
	altairData, err := altair.ComputeEpochAttesterData(ctx, spec, epc, flats, state)
	if err != nil {
		return nil, err
	}
	out = &EpochAttesterData{Altair: *altairData}

	// TODO: add sharding stats

	return out, nil
}
