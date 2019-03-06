package beacon

import (
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/transition"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

type BeaconState struct {
	// Misc
	Slot         eth2.Slot
	Genesis_time eth2.Timestamp
	Fork         Fork // For versioning hard forks

	// Validator registry
	Validator_registry              []Validator
	Validator_balances              []eth2.Gwei
	Validator_registry_update_epoch eth2.Epoch

	// Randomness and committees
	Latest_randao_mixes            []eth2.Bytes32
	Previous_shuffling_start_shard eth2.Shard
	Current_shuffling_start_shard  eth2.Shard
	Previous_shuffling_epoch       eth2.Epoch
	Current_shuffling_epoch        eth2.Epoch
	Previous_shuffling_seed        eth2.Bytes32
	Current_shuffling_seed         eth2.Bytes32

	// Finality
	PreviousEpochAttestations []PendingAttestation
	CurrentEpochAttestations  []PendingAttestation
	Previous_justified_epoch  eth2.Epoch
	Justified_epoch           eth2.Epoch
	Justification_bitfield    uint64
	Finalized_epoch           eth2.Epoch

	// Recent state
	Latest_crosslinks         [eth2.SHARD_COUNT]Crosslink
	Latest_block_roots        [eth2.SLOTS_PER_HISTORICAL_ROOT]eth2.Root
	Latest_state_roots        [eth2.SLOTS_PER_HISTORICAL_ROOT]eth2.Root
	Latest_active_index_roots [eth2.LATEST_ACTIVE_INDEX_ROOTS_LENGTH]eth2.Root
	// Balances slashed at every withdrawal period
	Latest_slashed_balances [eth2.LATEST_SLASHED_EXIT_LENGTH]eth2.Gwei
	LatestBlockHeader       BeaconBlockHeader
	Latest_attestations     []PendingAttestation
	HistoricalRoots     []eth2.Root

	// Ethereum 1.0 chain data
	//Latest_eth1_data Eth1Data
	Latest_eth1_data Eth1Data
	Eth1_data_votes []Eth1DataVote
	Deposit_index   eth2.DepositIndex
}

func GetGenesisBeaconState(validatorDeposits []Deposit, time eth2.Timestamp, eth1Data Eth1Data) *BeaconState {
	state := &BeaconState{
		Slot:         eth2.GENESIS_SLOT,
		Genesis_time: time,
		Fork: Fork{
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
		Justified_epoch: eth2.GENESIS_EPOCH,
		Finalized_epoch: eth2.GENESIS_EPOCH,
		// Recent state
		LatestBlockHeader: nil,//TODO
		// Ethereum 1.0 chain data
		Latest_eth1_data: eth1Data,
	}
	// Process genesis deposits
	for _, dep := range validatorDeposits {
		// in the rare case someone tries to create a genesis block using invalid data, panic.
		if err := transition.Process_deposit(state, &dep); err != nil {
			panic(err)
		}
	}
	for i := range state.Validator_registry {
		if transition.Get_effective_balance(state, eth2.ValidatorIndex(i)) >= eth2.MAX_DEPOSIT_AMOUNT {
			transition.Activate_validator(state, eth2.ValidatorIndex(i), true)
		}
	}
	genesis_active_index_root := ssz.Hash_tree_root(
		transition.Get_active_validator_indices(state.Validator_registry, eth2.GENESIS_EPOCH))
	for i := eth2.Epoch(0); i < eth2.LATEST_ACTIVE_INDEX_ROOTS_LENGTH; i++ {
		state.Latest_active_index_roots[i] = genesis_active_index_root
	}
	state.Current_shuffling_seed = transition.Generate_seed(state, eth2.GENESIS_EPOCH)
	return state
}

// Make a deep copy of the state object
func (st *BeaconState) Copy() *BeaconState {
	// copy over state
	stUn := *st
	res := &stUn
	// manually copy over slices
	// validators
	copy(res.Validator_registry, st.Validator_registry)
	copy(res.Validator_balances, st.Validator_balances)
	// recent state
	//copy(res.Latest_attestations, st.Latest_attestations)
	//copy(res.Batched_block_roots, st.Batched_block_roots)
	//// eth1
	//copy(res.Eth1_data_votes, st.Eth1_data_votes)
	return res
}

// Get current epoch
func (st *BeaconState) Epoch() eth2.Epoch {
	return st.Slot.ToEpoch()
}

// Return previous epoch. Not just current - 1: it's clipped to genesis.
func (st *BeaconState) PreviousEpoch() eth2.Epoch {
	epoch := st.Epoch()
	if epoch < eth2.GENESIS_EPOCH {
		return eth2.GENESIS_EPOCH
	} else {
		return epoch
	}
}
