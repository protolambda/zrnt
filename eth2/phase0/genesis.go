package phase0

import (
	. "github.com/protolambda/zrnt/eth2/beacon/deposits"
	. "github.com/protolambda/zrnt/eth2/beacon/eth1"
	. "github.com/protolambda/zrnt/eth2/beacon/registry"
	. "github.com/protolambda/zrnt/eth2/beacon/versioning"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func Genesis(deps []Deposit, time Timestamp, eth1Data Eth1Data) *BeaconState {
	state := &BeaconState{
		VersioningState: VersioningState{
			GenesisTime: time,
		},
		// Ethereum 1.0 chain data
		Eth1State: Eth1State{
			Eth1Data: eth1Data,
		},
	}
	depProcessor := &DepositFeature{Meta: state}
	// Process genesis deposits
	for i := range deps {
		// in the rare case someone tries to create a genesis block using invalid data, panic.
		if err := depProcessor.ProcessDeposit(&deps[i]); err != nil {
			panic(err)
		}
	}
	// Process genesis activations
	for _, v := range state.Validators {
		if v.EffectiveBalance >= MAX_EFFECTIVE_BALANCE {
			v.ActivationEligibilityEpoch = GENESIS_EPOCH
			v.ActivationEpoch = GENESIS_EPOCH
		}
	}
	indices := state.GetActiveValidatorIndices(GENESIS_EPOCH)
	genesisActiveIndexRoot := ssz.HashTreeRoot(indices, RegistryIndicesSSZ)
	for i := Epoch(0); i < EPOCHS_PER_HISTORICAL_VECTOR; i++ {
		state.LatestActiveIndexRoots[i] = genesisActiveIndexRoot
	}
	return state
}