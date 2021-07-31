package sharding

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

func ProcessEpochRewardsAndPenalties(ctx context.Context, spec *common.Spec, epc *common.EpochsContext,
	attesterData *altair.EpochAttesterData, state *BeaconStateView) error {
	currentEpoch := epc.CurrentEpoch.Epoch
	if currentEpoch == common.GENESIS_EPOCH {
		return nil
	}

	rewAndPenalties, err := altair.AttestationRewardsAndPenalties(ctx, spec, epc, attesterData, state)
	if err != nil {
		return err
	}

	valCount := uint64(len(attesterData.Flats))
	sum := common.NewDeltas(valCount)
	sum.Add(rewAndPenalties.Source)
	sum.Add(rewAndPenalties.Target)
	sum.Add(rewAndPenalties.Head)
	sum.Add(rewAndPenalties.Inactivity)
	// TODO: add sharding rewards/penalties
	balances, err := common.ApplyDeltas(state, sum)
	if err != nil {
		return err
	}
	return state.SetBalances(balances)
}
