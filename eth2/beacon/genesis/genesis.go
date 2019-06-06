package genesis

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func GetGenesisBeaconState(validatorDeposits []Deposit, time Timestamp, eth1Data Eth1Data) *BeaconState {
	state := &BeaconState{
		GenesisTime: time,
		// Ethereum 1.0 chain data
		LatestEth1Data: eth1Data,
	}
	// Process genesis deposits
	for _, dep := range validatorDeposits {
		// in the rare case someone tries to create a genesis block using invalid data, panic.
		if err := block_processing.ProcessDeposit(state, &dep); err != nil {
			panic(err)
		}
	}
	// Process genesis activations
	for _, v := range state.ValidatorRegistry {
		if v.EffectiveBalance >= MAX_EFFECTIVE_BALANCE {
			v.ActivationEligibilityEpoch = GENESIS_EPOCH
			v.ActivationEpoch = GENESIS_EPOCH
		}
	}
	genesisActiveIndexRoot := ssz.HashTreeRoot(
		state.ValidatorRegistry.GetActiveValidatorIndices(GENESIS_EPOCH), ValidatorIndexListSSZ)
	for i := Epoch(0); i < LATEST_ACTIVE_INDEX_ROOTS_LENGTH; i++ {
		state.LatestActiveIndexRoots[i] = genesisActiveIndexRoot
	}
	return state
}
