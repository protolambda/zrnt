package transition

import (
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func GetGenesisBeaconState(validatorDeposits []beacon.Deposit, time eth2.Timestamp, eth1Data beacon.Eth1Data) *beacon.BeaconState {
	state := &beacon.BeaconState{
		Slot:         eth2.GENESIS_SLOT,
		Genesis_time: time,
		Fork: beacon.Fork{
			Previous_version: eth2.GENESIS_FORK_VERSION,
			Current_version:  eth2.GENESIS_FORK_VERSION,
			Epoch:            eth2.GENESIS_EPOCH,
		},
		// Validator registry
		// (all default values)
		// Randomness and committees
		// (all default values)
		// Finality
		Previous_justified_epoch: eth2.GENESIS_EPOCH,
		Justified_epoch:          eth2.GENESIS_EPOCH,
		Finalized_epoch:          eth2.GENESIS_EPOCH,
		// Recent state
		LatestBlockHeader: beacon.GetEmptyBlock().GetTemporaryBlockHeader(),
		// Ethereum 1.0 chain data
		Latest_eth1_data: eth1Data,
	}
	// Process genesis deposits
	for _, dep := range validatorDeposits {
		// in the rare case someone tries to create a genesis block using invalid data, panic.
		if err := Process_deposit(state, &dep); err != nil {
			panic(err)
		}
	}
	for i := range state.Validator_registry {
		if Get_effective_balance(state, eth2.ValidatorIndex(i)) >= eth2.MAX_DEPOSIT_AMOUNT {
			Activate_validator(state, eth2.ValidatorIndex(i), true)
		}
	}
	genesis_active_index_root := ssz.Hash_tree_root(
		Get_active_validator_indices(state.Validator_registry, eth2.GENESIS_EPOCH))
	for i := eth2.Epoch(0); i < eth2.LATEST_ACTIVE_INDEX_ROOTS_LENGTH; i++ {
		state.Latest_active_index_roots[i] = genesis_active_index_root
	}
	state.Current_shuffling_seed = Generate_seed(state, eth2.GENESIS_EPOCH)
	return state
}
