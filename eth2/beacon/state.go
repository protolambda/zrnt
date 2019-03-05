package beacon

import "github.com/protolambda/go-beacon-transition/eth2"

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
	//Previous_shuffling_start_shard eth2.Shard
	//Current_shuffling_start_shard  eth2.Shard
	//Previous_shuffling_epoch       eth2.Epoch
	//Current_shuffling_epoch        eth2.Epoch
	//Previous_shuffling_seed        eth2.Bytes32
	//Current_shuffling_seed         eth2.Bytes32

	// Finality
	//Previous_justified_epoch eth2.Epoch
	//Justified_epoch          eth2.Epoch
	//Justification_bitfield   uint64
	//Finalized_epoch          eth2.Epoch
	//
	//// Recent state
	//Latest_crosslinks         [eth2.SHARD_COUNT]Crosslink
	//Latest_block_roots        [eth2.LATEST_BLOCK_ROOTS_LENGTH]eth2.Root
	//Latest_active_index_roots [eth2.LATEST_ACTIVE_INDEX_ROOTS_LENGTH]eth2.Root
	//// Balances slashed at every withdrawal period
	//Latest_slashed_balances [eth2.LATEST_SLASHED_EXIT_LENGTH]eth2.Gwei
	//Latest_attestations     []PendingAttestation
	//Batched_block_roots     []eth2.Root
	//
	//// Ethereum 1.0 chain data
	////Latest_eth1_data Eth1Data
	//Eth1_data_votes  []Eth1DataVote
	//Deposit_index    eth2.DepositIndex
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
