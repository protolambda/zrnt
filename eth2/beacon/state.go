package beacon

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/protolambda/go-beacon-transition/eth2/util/bitfield"
	"github.com/protolambda/go-beacon-transition/eth2/util/hash"
	"github.com/protolambda/go-beacon-transition/eth2/util/math"
	"sort"
)

type BeaconState struct {
	// Misc
	Slot         Slot
	Genesis_time Timestamp
	Fork         Fork

	// Validator registry
	Validator_registry              ValidatorRegistry
	Validator_balances              ValidatorBalances
	Validator_registry_update_epoch Epoch

	// Randomness and committees
	Latest_randao_mixes            [LATEST_RANDAO_MIXES_LENGTH]Bytes32
	Previous_shuffling_start_shard Shard
	Current_shuffling_start_shard  Shard
	Previous_shuffling_epoch       Epoch
	Current_shuffling_epoch        Epoch
	Previous_shuffling_seed        Bytes32
	Current_shuffling_seed         Bytes32

	// Finality
	PreviousEpochAttestations []PendingAttestation
	CurrentEpochAttestations  []PendingAttestation
	Previous_justified_epoch  Epoch
	Justified_epoch           Epoch
	Justification_bitfield    uint64
	Finalized_epoch           Epoch

	// Recent state
	Latest_crosslinks         [SHARD_COUNT]Crosslink
	Latest_block_roots        [SLOTS_PER_HISTORICAL_ROOT]Root
	Latest_state_roots        [SLOTS_PER_HISTORICAL_ROOT]Root
	Latest_active_index_roots [LATEST_ACTIVE_INDEX_ROOTS_LENGTH]Root
	// Balances slashed at every withdrawal period
	Latest_slashed_balances [LATEST_SLASHED_EXIT_LENGTH]Gwei
	LatestBlockHeader       BeaconBlockHeader
	HistoricalRoots         []Root

	// Ethereum 1.0 chain data
	Latest_eth1_data Eth1Data
	Eth1_data_votes  []Eth1DataVote
	Deposit_index    DepositIndex
}

// Make a deep copy of the state object
func (state *BeaconState) Copy() *BeaconState {
	// copy over state
	stUn := *state
	res := &stUn
	// manually copy over slices
	// validators
	copy(res.Validator_registry, state.Validator_registry)
	copy(res.Validator_balances, state.Validator_balances)
	// finality
	copy(res.PreviousEpochAttestations, state.PreviousEpochAttestations)
	copy(res.CurrentEpochAttestations, state.CurrentEpochAttestations)
	// recent state
	copy(res.HistoricalRoots, state.HistoricalRoots)
	// eth1
	copy(res.Eth1_data_votes, state.Eth1_data_votes)
	return res
}

// Get current epoch
func (state *BeaconState) Epoch() Epoch {
	return state.Slot.ToEpoch()
}

// Return previous epoch. Not just current - 1: it's clipped to genesis.
func (state *BeaconState) PreviousEpoch() Epoch {
	epoch := state.Epoch()
	if epoch < GENESIS_EPOCH {
		return GENESIS_EPOCH
	} else {
		return epoch
	}
}

// Set the validator with the given index as withdrawable
// MIN_VALIDATOR_WITHDRAWABILITY_DELAY after the current epoch.
func (state *BeaconState) Prepare_validator_for_withdrawal(index ValidatorIndex) {
	state.Validator_registry[index].Withdrawable_epoch = state.Epoch() + MIN_VALIDATOR_WITHDRAWABILITY_DELAY
}

// Initiate the validator of the given index
func (state *BeaconState) Initiate_validator_exit(index ValidatorIndex) {
	state.Validator_registry[index].Initiated_exit = true
}

// Activate the validator of the given index
func (state *BeaconState) Activate_validator(index ValidatorIndex, is_genesis bool) {
	validator := &state.Validator_registry[index]

	if is_genesis {
		validator.Activation_epoch = GENESIS_EPOCH
	} else {
		validator.Activation_epoch = state.Epoch().Get_delayed_activation_exit_epoch()
	}
}

func Get_epoch_committee_count(active_validator_count uint64) uint64 {
	return math.MaxU64(1,
		math.MinU64(
			uint64(SHARD_COUNT)/uint64(SLOTS_PER_EPOCH),
			active_validator_count/uint64(SLOTS_PER_EPOCH)/TARGET_COMMITTEE_SIZE,
		)) * uint64(SLOTS_PER_EPOCH)
}

// Return the number of committees in the previous epoch
func (state *BeaconState) Get_previous_epoch_committee_count() uint64 {
	return Get_epoch_committee_count(
		state.Validator_registry.Get_active_validator_count(
			state.Previous_shuffling_epoch,
		))
}

// Return the number of committees in the current epoch
func (state *BeaconState) Get_current_epoch_committee_count() uint64 {
	return Get_epoch_committee_count(
		state.Validator_registry.Get_active_validator_count(
			state.Current_shuffling_epoch,
		))
}

// Return the number of committees in the next epoch
func (state *BeaconState) Get_next_epoch_committee_count() uint64 {
	return Get_epoch_committee_count(
		state.Validator_registry.Get_active_validator_count(
			state.Epoch() + 1,
		))
}

// Return the beacon proposer index for the slot.
func (state *BeaconState) Get_beacon_proposer_index(slot Slot, registryChange bool) ValidatorIndex {
	epoch := slot.ToEpoch()
	currentEpoch := state.Epoch()
	if currentEpoch-1 <= epoch && epoch <= currentEpoch+1 {
		panic("epoch of given slot out of range")
	}
	committeeData := state.Get_crosslink_committees_at_slot(slot, registryChange)
	first_committee_data := committeeData[0]
	return first_committee_data.Committee[slot%Slot(len(first_committee_data.Committee))]
}

//  Return the randao mix at a recent epoch
func (state *BeaconState) Get_randao_mix(epoch Epoch) Bytes32 {
	// Every usage is a trusted input (i.e. state is already up to date to handle the requested epoch number).
	// If something is wrong due to unforeseen usage, panic to catch it during development.
	if !(state.Epoch()-LATEST_RANDAO_MIXES_LENGTH < epoch && epoch <= state.Epoch()) {
		panic("cannot get randao mix for out-of-bounds epoch")
	}
	return state.Latest_randao_mixes[epoch%LATEST_RANDAO_MIXES_LENGTH]
}

func (state *BeaconState) Get_active_index_root(epoch Epoch) Root {
	return state.Latest_active_index_roots[epoch%LATEST_ACTIVE_INDEX_ROOTS_LENGTH]
}

// Generate a seed for the given epoch
func (state *BeaconState) Generate_seed(epoch Epoch) Bytes32 {
	buf := make([]byte, 32*3)
	mix := state.Get_randao_mix(epoch - MIN_SEED_LOOKAHEAD)
	copy(buf[0:32], mix[:])
	// get_active_index_root in spec, but only used once, and the assertion is unnecessary, since epoch input is always trusted
	activeIndexRoot := state.Get_active_index_root(epoch)
	copy(buf[32:64], activeIndexRoot[:])
	binary.LittleEndian.PutUint64(buf[64:], uint64(epoch))
	return hash.Hash(buf)
}

// Return the block root at a recent slot
func (state *BeaconState) Get_block_root(slot Slot) (Root, error) {
	if slot+SLOTS_PER_HISTORICAL_ROOT < state.Slot || slot > state.Slot {
		return Root{}, errors.New("cannot get block root for given slot")
	}
	return state.Latest_block_roots[slot%SLOTS_PER_HISTORICAL_ROOT], nil
}

// Return the state root at a recent
func (state *BeaconState) Get_state_root(slot Slot) (Root, error) {
	if slot+SLOTS_PER_HISTORICAL_ROOT < state.Slot || slot > state.Slot {
		return Root{}, errors.New("cannot get state root for given slot")
	}
	return state.Latest_state_roots[slot%SLOTS_PER_HISTORICAL_ROOT], nil
}

type CrosslinkCommittee struct {
	Committee []ValidatorIndex
	Shard     Shard
}

// Return the list of (committee, shard) tuples for the slot.
//
// Note: There are two possible shufflings for crosslink committees for a
//  slot in the next epoch -- with and without a registryChange
func (state *BeaconState) Get_crosslink_committees_at_slot(slot Slot, registryChange bool) []CrosslinkCommittee {
	epoch, current_epoch, previous_epoch := slot.ToEpoch(), state.Epoch(), state.PreviousEpoch()
	next_epoch := current_epoch + 1

	if !(previous_epoch <= epoch && epoch <= next_epoch) {
		panic("could not retrieve crosslink committee for out of range slot")
	}

	var committees_per_epoch uint64
	var seed Bytes32
	var shuffling_epoch Epoch
	var shuffling_start_shard Shard
	if epoch == current_epoch {
		committees_per_epoch = state.Get_current_epoch_committee_count()
		seed = state.Current_shuffling_seed
		shuffling_epoch = state.Current_shuffling_epoch
		shuffling_start_shard = state.Current_shuffling_start_shard
	} else if epoch == previous_epoch {
		committees_per_epoch = state.Get_previous_epoch_committee_count()
		seed = state.Previous_shuffling_seed
		shuffling_epoch = state.Previous_shuffling_epoch
		shuffling_start_shard = state.Previous_shuffling_start_shard
	} else if epoch == next_epoch {
		epochs_since_last_registry_update := current_epoch - state.Validator_registry_update_epoch
		if registryChange {
			committees_per_epoch = state.Get_next_epoch_committee_count()
			seed = state.Generate_seed(next_epoch)
			shuffling_epoch = next_epoch
			current_committees_per_epoch := state.Get_current_epoch_committee_count()
			shuffling_start_shard = (state.Current_shuffling_start_shard + Shard(current_committees_per_epoch)) % SHARD_COUNT
		} else if epochs_since_last_registry_update > 1 && math.Is_power_of_two(uint64(epochs_since_last_registry_update)) {
			committees_per_epoch = state.Get_next_epoch_committee_count()
			seed = state.Generate_seed(next_epoch)
			shuffling_epoch = next_epoch
			shuffling_start_shard = state.Current_shuffling_start_shard
		} else {
			committees_per_epoch = state.Get_current_epoch_committee_count()
			seed = state.Current_shuffling_seed
			shuffling_epoch = state.Current_shuffling_epoch
			shuffling_start_shard = state.Current_shuffling_start_shard
		}
	}
	shuffling := state.Validator_registry.Get_shuffling(seed, shuffling_epoch)
	offset := slot % SLOTS_PER_EPOCH
	committees_per_slot := committees_per_epoch / uint64(SLOTS_PER_EPOCH)
	slot_start_shard := (shuffling_start_shard + Shard(committees_per_slot)*Shard(offset)) % SHARD_COUNT

	crosslink_committees := make([]CrosslinkCommittee, committees_per_slot)
	for i := uint64(0); i < committees_per_slot; i++ {
		crosslink_committees[i] = CrosslinkCommittee{
			Committee: shuffling[committees_per_slot*uint64(offset)+i],
			Shard:     (slot_start_shard + Shard(i)) % SHARD_COUNT}
	}
	return crosslink_committees
}

func (state *BeaconState) Get_winning_root_and_participants(shard Shard) (Root, []ValidatorIndex) {
	weightedCrosslinks := make(map[Root]Gwei)

	updateCrosslinkWeights := func(att *PendingAttestation) {
		if att.Data.Latest_crosslink == state.Latest_crosslinks[shard] {
			participants, _ := state.Get_attestation_participants(&att.Data, &att.Aggregation_bitfield)
			for _, participant := range participants {
				weightedCrosslinks[att.Data.Crosslink_data_root] += state.Validator_balances.Get_effective_balance(participant)
			}
		}
	}
	for i := 0; i < len(state.PreviousEpochAttestations); i++ {
		updateCrosslinkWeights(&state.PreviousEpochAttestations[i])
	}
	for i := 0; i < len(state.CurrentEpochAttestations); i++ {
		updateCrosslinkWeights(&state.CurrentEpochAttestations[i])
	}

	// handle when no attestations for shard available
	if len(weightedCrosslinks) == 0 {
		return Root{}, nil
	}
	// Now determine the best root, by total weight (votes, weighted by balance)
	var winning_root Root
	winning_weight := Gwei(0)
	for root, weight := range weightedCrosslinks {
		if weight > winning_weight {
			winning_root = root
		}
		if weight == winning_weight {
			// break tie lexicographically
			for i := 0; i < 32; i++ {
				if root[i] > winning_root[i] {
					winning_root = root
					break
				}
			}
		}
	}

	// now retrieve all the attesters of this winning root
	winning_attesters_set := make(map[ValidatorIndex]struct{})
	findWinners := func(att *PendingAttestation) {
		if att.Data.Crosslink_data_root == winning_root {
			participants, _ := state.Get_attestation_participants(&att.Data, &att.Aggregation_bitfield)
			for _, participant := range participants {
				winning_attesters_set[participant] = struct{}{}
			}
		}
	}
	for i := 0; i < len(state.PreviousEpochAttestations); i++ {
		findWinners(&state.PreviousEpochAttestations[i])
	}
	for i := 0; i < len(state.CurrentEpochAttestations); i++ {
		findWinners(&state.CurrentEpochAttestations[i])
	}
	winning_attesters := make([]ValidatorIndex, len(winning_attesters_set))
	i := 0
	for attester := range winning_attesters_set {
		winning_attesters[i] = attester
		i++
	}
	// Spec returns it in sorted order, although not strictly necessary (TODO)
	sort.Slice(winning_attesters, func(i int, j int) bool {
		return winning_attesters[i] < winning_attesters[j]
	})

	return winning_root, winning_attesters
}

// Exit the validator of the given index
func (state *BeaconState) Exit_validator(index ValidatorIndex) {
	validator := &state.Validator_registry[index]
	delayed_activation_exit_epoch := state.Epoch().Get_delayed_activation_exit_epoch()
	// The following updates only occur if not previous exited
	if validator.Exit_epoch > delayed_activation_exit_epoch {
		return
	}
	validator.Exit_epoch = delayed_activation_exit_epoch
}

// Update validator registry.
func (state *BeaconState) Update_validator_registry() {
	// The total effective balance of active validators
	total_balance := state.Validator_balances.Get_total_balance(state.Validator_registry.Get_active_validator_indices(state.Epoch()))

	// The maximum balance churn in Gwei (for deposits and exits separately)
	max_balance_churn := Max(MAX_DEPOSIT_AMOUNT, total_balance/(2*MAX_BALANCE_CHURN_QUOTIENT))

	// Activate validators within the allowable balance churn
	{
		balance_churn := Gwei(0)
		for index, validator := range state.Validator_registry {
			if validator.Activation_epoch == FAR_FUTURE_EPOCH && state.Validator_balances[index] >= MAX_DEPOSIT_AMOUNT {
				// Check the balance churn would be within the allowance
				balance_churn += state.Validator_balances.Get_effective_balance(ValidatorIndex(index))
				if balance_churn > max_balance_churn {
					break
				}
				//  Activate validator
				validator.Activation_epoch = state.Epoch().Get_delayed_activation_exit_epoch()
			}
		}
	}

	// Exit validators within the allowable balance churn
	{
		balance_churn := Gwei(0)
		for index, validator := range state.Validator_registry {
			if validator.Exit_epoch == FAR_FUTURE_EPOCH && validator.Initiated_exit {
				// Check the balance churn would be within the allowance
				balance_churn += state.Validator_balances.Get_effective_balance(ValidatorIndex(index))
				if balance_churn > max_balance_churn {
					break
				}
				// Exit validator
				state.Exit_validator(ValidatorIndex(index))
			}
		}
	}
}

// Return the participant indices at for the attestation_data and bitfield
func (state *BeaconState) Get_attestation_participants(attestation_data *AttestationData, bitfield *bitfield.Bitfield) ([]ValidatorIndex, error) {
	// Find the committee in the list with the desired shard
	crosslink_committees := state.Get_crosslink_committees_at_slot(attestation_data.Slot, false)

	var crosslink_committee []ValidatorIndex
	for _, cross_comm := range crosslink_committees {
		if cross_comm.Shard == attestation_data.Shard {
			crosslink_committee = cross_comm.Committee
			break
		}
	}
	if len(crosslink_committee) == 0 {
		return nil, errors.New(fmt.Sprintf("cannot find crosslink committee at slot %d for shard %d", attestation_data.Slot, attestation_data.Shard))
	}
	if !bitfield.VerifySize(uint64(len(crosslink_committee))) {
		return nil, errors.New("bitfield has wrong size for corresponding crosslink committee")
	}

	// Find the participating attesters in the committee
	participants := make([]ValidatorIndex, 0)
	for i, vIndex := range crosslink_committee {
		if bitfield.GetBit(uint64(i)) == 1 {
			participants = append(participants, vIndex)
		}
	}
	return participants, nil
}

// Slash the validator with index index.
func (state *BeaconState) Slash_validator(index ValidatorIndex) error {
	validator := &state.Validator_registry[index]
	// [TO BE REMOVED IN PHASE 2] (this is to make phase 0 and phase 1 consistent with behavior in phase 2)
	if state.Slot >= validator.Withdrawable_epoch.GetStartSlot() {
		return errors.New("cannot slash validator after withdrawal epoch")
	}
	state.Exit_validator(index)
	state.Latest_slashed_balances[state.Epoch()%LATEST_SLASHED_EXIT_LENGTH] += state.Validator_balances.Get_effective_balance(index)

	whistleblower_reward := state.Validator_balances.Get_effective_balance(index) / WHISTLEBLOWER_REWARD_QUOTIENT
	prop_index := state.Get_beacon_proposer_index(state.Slot, false)
	state.Validator_balances[prop_index] += whistleblower_reward
	state.Validator_balances[index] -= whistleblower_reward
	validator.Slashed = true
	validator.Withdrawable_epoch = state.Epoch() + LATEST_SLASHED_EXIT_LENGTH
	return nil
}
