package fulu

import (
	"errors"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/electra"
)

func UpgradeToFulu(spec *common.Spec, epc *common.EpochsContext, pre *electra.BeaconStateView) (*BeaconStateView, error) {
	return nil, errors.New("upgrade of electra state to fulu state is not supported")
}