package genesis

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/validator_balances/deposits"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func GetGenesisBeaconState(validatorDeposits []beacon.Deposit, time beacon.Timestamp, eth1Data beacon.Eth1Data) *beacon.BeaconState {
	state := &beacon.BeaconState{
		Slot:         beacon.GENESIS_SLOT,
		Genesis_time: time,
		Fork: beacon.Fork{
			Previous_version: beacon.GENESIS_FORK_VERSION,
			Current_version:  beacon.GENESIS_FORK_VERSION,
			Epoch:            beacon.GENESIS_EPOCH,
		},
		// Validator registry
		Validator_registry_update_epoch: beacon.GENESIS_EPOCH,
		// Randomness and committees
		Previous_shuffling_start_shard: beacon.GENESIS_START_SHARD,
		Current_shuffling_start_shard:  beacon.GENESIS_START_SHARD,
		Current_shuffling_epoch:        beacon.GENESIS_EPOCH,
		Previous_shuffling_epoch:       beacon.GENESIS_EPOCH,
		// Finality
		Previous_justified_epoch: beacon.GENESIS_EPOCH,
		Justified_epoch:          beacon.GENESIS_EPOCH,
		Finalized_epoch:          beacon.GENESIS_EPOCH,
		// Recent state
		LatestBlockHeader: beacon.GetEmptyBlock().GetTemporaryBlockHeader(),
		// Ethereum 1.0 chain data
		Latest_eth1_data: eth1Data,
	}
	// Initialize crosslinks
	for i := beacon.Shard(0); i < beacon.SHARD_COUNT; i++ {
		state.Latest_crosslinks[i] = beacon.Crosslink{Epoch: beacon.GENESIS_EPOCH}
	}
	// Process genesis deposits
	for _, dep := range validatorDeposits {
		// in the rare case someone tries to create a genesis block using invalid data, panic.
		if err := deposits.ProcessDeposit(state, &dep); err != nil {
			panic(err)
		}
	}
	for i := range state.Validator_registry {
		if state.Validator_balances.Get_effective_balance(beacon.ValidatorIndex(i)) >= beacon.MAX_DEPOSIT_AMOUNT {
			state.Activate_validator(beacon.ValidatorIndex(i), true)
		}
	}
	genesis_active_index_root := ssz.Hash_tree_root(
		state.Validator_registry.Get_active_validator_indices(beacon.GENESIS_EPOCH))
	for i := beacon.Epoch(0); i < beacon.LATEST_ACTIVE_INDEX_ROOTS_LENGTH; i++ {
		state.Latest_active_index_roots[i] = genesis_active_index_root
	}
	state.Current_shuffling_seed = state.Generate_seed(beacon.GENESIS_EPOCH)
	return state
}
