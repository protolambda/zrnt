package beacon

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/validator_balances/deposits"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func GetGenesisBeaconState(validatorDeposits []Deposit, time Timestamp, eth1Data Eth1Data) *BeaconState {
	state := &BeaconState{
		Slot:         GENESIS_SLOT,
		Genesis_time: time,
		Fork: Fork{
			Previous_version: GENESIS_FORK_VERSION,
			Current_version:  GENESIS_FORK_VERSION,
			Epoch:            GENESIS_EPOCH,
		},
		// Validator registry
		// (all default values)
		// Randomness and committees
		// (all default values)
		// Finality
		Previous_justified_epoch: GENESIS_EPOCH,
		Justified_epoch:          GENESIS_EPOCH,
		Finalized_epoch:          GENESIS_EPOCH,
		// Recent state
		LatestBlockHeader: GetEmptyBlock().GetTemporaryBlockHeader(),
		// Ethereum 1.0 chain data
		Latest_eth1_data: eth1Data,
	}
	// Process genesis deposits
	for _, dep := range validatorDeposits {
		// in the rare case someone tries to create a genesis block using invalid data, panic.
		if err := deposits.ProcessDeposit(state, &dep); err != nil {
			panic(err)
		}
	}
	for i := range state.Validator_registry {
		if state.Validator_balances.Get_effective_balance(ValidatorIndex(i)) >= MAX_DEPOSIT_AMOUNT {
			state.Activate_validator(ValidatorIndex(i), true)
		}
	}
	genesis_active_index_root := ssz.Hash_tree_root(
		state.Validator_registry.Get_active_validator_indices(GENESIS_EPOCH))
	for i := Epoch(0); i < LATEST_ACTIVE_INDEX_ROOTS_LENGTH; i++ {
		state.Latest_active_index_roots[i] = genesis_active_index_root
	}
	state.Current_shuffling_seed = state.Generate_seed(GENESIS_EPOCH)
	return state
}
