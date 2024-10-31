package phase0

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	. "github.com/protolambda/ztyp/view"
)

var DepositRootsType = ComplexListType(RootType, 1<<common.DEPOSIT_CONTRACT_TREE_DEPTH)

type DepositRootsView struct {
	*ComplexListView
}

func AsDepositRootsView(v View, err error) (*DepositRootsView, error) {
	c, err := AsComplexList(v, err)
	return &DepositRootsView{c}, err
}

func NewDepositRootsView() *DepositRootsView {
	return &DepositRootsView{DepositRootsType.New()}
}

func IsValidGenesisState(spec *common.Spec, state common.BeaconState) (bool, error) {
	genTime, err := state.GenesisTime()
	if err != nil {
		return false, err
	}
	if genTime < spec.MIN_GENESIS_TIME {
		return false, nil
	}

	// outside of genesis we have this precomputed at all times. Just compute it manually this time.
	activeCount := uint64(0)
	{
		validators, err := state.Validators()
		if err != nil {
			return false, err
		}
		valIterNext := validators.Iter()
		for {
			val, ok, err := valIterNext()
			if err != nil {
				return false, err
			}
			if !ok {
				break
			}
			if active, err := IsActive(val, common.GENESIS_EPOCH); err != nil {
				return false, err
			} else if active {
				activeCount += 1
			}
		}
	}
	return activeCount >= uint64(spec.MIN_GENESIS_ACTIVE_VALIDATOR_COUNT), nil
}
