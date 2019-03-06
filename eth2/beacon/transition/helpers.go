package transition

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/protolambda/eth2-shuffle"
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/bitfield"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/hash"
	"github.com/protolambda/go-beacon-transition/eth2/util/math"
	"github.com/protolambda/go-beacon-transition/eth2/util/merkle"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
	"github.com/protolambda/go-beacon-transition/eth2/util/validatorset"
)

// Set the validator with the given index as withdrawable
// MIN_VALIDATOR_WITHDRAWABILITY_DELAY after the current epoch.
func Prepare_validator_for_withdrawal(state *beacon.BeaconState, index eth2.ValidatorIndex) {
	state.Validator_registry[index].Withdrawable_epoch = state.Epoch() + eth2.MIN_VALIDATOR_WITHDRAWABILITY_DELAY
}

// Return the epoch at which an activation or exit triggered in epoch takes effect.
func Get_delayed_activation_exit_epoch(epoch eth2.Epoch) eth2.Epoch {
	return epoch + 1 + eth2.ACTIVATION_EXIT_DELAY
}

// Exit the validator of the given index
func Exit_validator(state *beacon.BeaconState, index eth2.ValidatorIndex) {
	validator := &state.Validator_registry[index]
	delayed_activation_exit_epoch := Get_delayed_activation_exit_epoch(state.Epoch())
	// The following updates only occur if not previous exited
	if validator.Exit_epoch > delayed_activation_exit_epoch {
		return
	}
	validator.Exit_epoch = delayed_activation_exit_epoch
}

// Initiate the validator of the given index
func Initiate_validator_exit(state *beacon.BeaconState, index eth2.ValidatorIndex) {
	state.Validator_registry[index].Initiated_exit = true
}

// Activate the validator of the given index
func Activate_validator(state *beacon.BeaconState, index eth2.ValidatorIndex, is_genesis bool) {
	validator := &state.Validator_registry[index]

	if is_genesis {
		validator.Activation_epoch = eth2.GENESIS_EPOCH
	} else {
		validator.Activation_epoch = Get_delayed_activation_exit_epoch(state.Epoch())
	}
}

func Get_active_validator_count(validator_registry []beacon.Validator, epoch eth2.Epoch) (count uint64) {
	for _, v := range validator_registry {
		if v.IsActive(epoch) {
			count++
		}
	}
	return count
}

func Get_active_validator_indices(validator_registry []beacon.Validator, epoch eth2.Epoch) []eth2.ValidatorIndex {
	res := make([]eth2.ValidatorIndex, 0, len(validator_registry))
	for i, v := range validator_registry {
		if v.IsActive(epoch) {
			res = append(res, eth2.ValidatorIndex(i))
		}
	}
	return res
}

// Return the effective balance (also known as "balance at stake") for a validator with the given index.
func Get_effective_balance(state *beacon.BeaconState, index eth2.ValidatorIndex) eth2.Gwei {
	return math.Max(state.Validator_balances[index], eth2.MAX_DEPOSIT_AMOUNT)
}

// Return the combined effective balance of an array of validators.
func Get_total_balance(state *beacon.BeaconState, indices []eth2.ValidatorIndex) (sum eth2.Gwei) {
	for _, vIndex := range indices {
		sum += Get_effective_balance(state, vIndex)
	}
	return sum
}



// Update validator registry.
func Update_validator_registry(state *beacon.BeaconState) {
	// The total effective balance of active validators
	total_balance := Get_total_balance(state, Get_active_validator_indices(state.Validator_registry, state.Epoch()))

	// The maximum balance churn in Gwei (for deposits and exits separately)
	max_balance_churn := math.Max(eth2.MAX_DEPOSIT_AMOUNT, total_balance/(2*eth2.MAX_BALANCE_CHURN_QUOTIENT))

	// Activate validators within the allowable balance churn
	{
		balance_churn := eth2.Gwei(0)
		for index, validator := range state.Validator_registry {
			if validator.Activation_epoch == eth2.FAR_FUTURE_EPOCH && state.Validator_balances[index] >= eth2.MAX_DEPOSIT_AMOUNT {
				// Check the balance churn would be within the allowance
				balance_churn += Get_effective_balance(state, eth2.ValidatorIndex(index))
				if balance_churn > max_balance_churn {
					break
				}
				//  Activate validator
				validator.Activation_epoch = Get_delayed_activation_exit_epoch(state.Epoch())
			}
		}
	}

	// Exit validators within the allowable balance churn
	{
		balance_churn := eth2.Gwei(0)
		for index, validator := range state.Validator_registry {
			if validator.Exit_epoch == eth2.FAR_FUTURE_EPOCH && validator.Initiated_exit {
				// Check the balance churn would be within the allowance
				balance_churn += Get_effective_balance(state, eth2.ValidatorIndex(index))
				if balance_churn > max_balance_churn {
					break
				}
				// Exit validator
				Exit_validator(state, eth2.ValidatorIndex(index))
			}
		}
	}
}

// Return the participant indices at for the attestation_data and bitfield
func Get_attestation_participants(state *beacon.BeaconState, attestation_data *beacon.AttestationData, bitfield *bitfield.Bitfield) (validatorset.ValidatorIndexSet, error) {
	// Find the committee in the list with the desired shard
	crosslink_committees, err := Get_crosslink_committees_at_slot(state, attestation_data.Slot, false)
	if err != nil {
		return nil, err
	}

	var crosslink_committee []eth2.ValidatorIndex
	for _, cross_comm := range crosslink_committees {
		if cross_comm.Shard == attestation_data.Shard {
			crosslink_committee = cross_comm.Committee
			break
		}
	}
	if crosslink_committee == nil {
		return nil, errors.New(fmt.Sprintf("cannot find crosslink committee at slot %d for shard %d", attestation_data.Slot, attestation_data.Shard))
	}
	if !bitfield.VerifySize(uint64(len(crosslink_committee))) {
		return nil, errors.New("bitfield has wrong size for corresponding crosslink committee")
	}

	// Find the participating attesters in the committee
	participants := make(validatorset.ValidatorIndexSet, 0)
	for i, vIndex := range crosslink_committee {
		if bitfield.GetBit(uint64(i)) == 1 {
			participants = append(participants, vIndex)
		}
	}
	return participants, nil
}

// Generate a seed for the given epoch
func Generate_seed(state *beacon.BeaconState, epoch eth2.Epoch) eth2.Bytes32 {
	buf := make([]byte, 32*3)
	mix := Get_randao_mix(state, epoch-eth2.MIN_SEED_LOOKAHEAD)
	copy(buf[0:32], mix[:])
	// get_active_index_root in spec, but only used once, and the assertion is unnecessary, since epoch input is always trusted
	copy(buf[32:32*2], state.Latest_active_index_roots[epoch%eth2.LATEST_ACTIVE_INDEX_ROOTS_LENGTH][:])
	binary.LittleEndian.PutUint64(buf[32*3-8:], uint64(epoch))
	return hash.Hash(buf)
}

type CrosslinkCommittee struct {
	Committee []eth2.ValidatorIndex
	Shard     eth2.Shard
}

// Return the list of (committee, shard) tuples for the slot.
//
// Note: There are two possible shufflings for crosslink committees for a
//  slot in the next epoch -- with and without a registryChange
func Get_crosslink_committees_at_slot(state *beacon.BeaconState, slot eth2.Slot, registryChange bool) ([]CrosslinkCommittee, error) {
	epoch, current_epoch, previous_epoch := slot.ToEpoch(), state.Epoch(), state.PreviousEpoch()
	next_epoch := current_epoch + 1

	if !(previous_epoch <= epoch && epoch <= next_epoch) {
		return nil, errors.New("could not retrieve crosslink committee for out of range slot")
	}

	var committees_per_epoch uint64
	var seed eth2.Bytes32
	var shuffling_epoch eth2.Epoch
	var shuffling_start_shard eth2.Shard
	if epoch == current_epoch {
		committees_per_epoch = Get_current_epoch_committee_count(state)
		seed = state.Current_shuffling_seed
		shuffling_epoch = state.Current_shuffling_epoch
		shuffling_start_shard = state.Current_shuffling_start_shard
	} else if epoch == previous_epoch {
		committees_per_epoch = Get_previous_epoch_committee_count(state)
		seed = state.Previous_shuffling_seed
		shuffling_epoch = state.Previous_shuffling_epoch
		shuffling_start_shard = state.Previous_shuffling_start_shard
	} else if epoch == next_epoch {
		epochs_since_last_registry_update := current_epoch - state.Validator_registry_update_epoch
		if registryChange {
			committees_per_epoch = Get_next_epoch_committee_count(state)
			seed = Generate_seed(state, next_epoch)
			shuffling_epoch = next_epoch
			current_committees_per_epoch := Get_current_epoch_committee_count(state)
			shuffling_start_shard = (state.Current_shuffling_start_shard + eth2.Shard(current_committees_per_epoch)) % eth2.SHARD_COUNT
		} else if epochs_since_last_registry_update > 1 && math.Is_power_of_two(uint64(epochs_since_last_registry_update)) {
			committees_per_epoch = Get_next_epoch_committee_count(state)
			seed = Generate_seed(state, next_epoch)
			shuffling_epoch = next_epoch
			shuffling_start_shard = state.Current_shuffling_start_shard
		} else {
			committees_per_epoch = Get_current_epoch_committee_count(state)
			seed = state.Current_shuffling_seed
			shuffling_epoch = state.Current_shuffling_epoch
			shuffling_start_shard = state.Current_shuffling_start_shard

		}
	}
	shuffling := Get_shuffling(seed, state.Validator_registry, shuffling_epoch)
	offset := slot % eth2.SLOTS_PER_EPOCH
	committees_per_slot := committees_per_epoch / uint64(eth2.SLOTS_PER_EPOCH)
	slot_start_shard := (shuffling_start_shard + eth2.Shard(committees_per_slot)*eth2.Shard(offset)) % eth2.SHARD_COUNT

	crosslink_committees := make([]CrosslinkCommittee, committees_per_slot)
	for i := uint64(0); i < committees_per_slot; i++ {
		crosslink_committees[i] = CrosslinkCommittee{
			Committee: shuffling[committees_per_slot*uint64(offset)+i],
			Shard:     (slot_start_shard + eth2.Shard(i)) % eth2.SHARD_COUNT}
	}
	return crosslink_committees, nil
}

// Shuffle active validators and split into crosslink committees.
// Return a list of committees (each a list of validator indices).
func Get_shuffling(seed eth2.Bytes32, validators []beacon.Validator, epoch eth2.Epoch) [][]eth2.ValidatorIndex {
	active_validator_indices := Get_active_validator_indices(validators, epoch)
	committee_count := Get_epoch_committee_count(uint64(len(active_validator_indices)))
	commitees := make([][]eth2.ValidatorIndex, committee_count, committee_count)
	// Active validators, shuffled in-place.
	// TODO shuffleValidatorIndices(active_validator_indices, seed)
	hashFn := func(input []byte) []byte {
		res := hash.Hash(input)
		return res[:]
	}
	eth2_shuffle.ShuffleList(hashFn, eth2.ValidatorIndexList(active_validator_indices).RawIndexSlice(), eth2.SHUFFLE_ROUND_COUNT, seed)
	committee_size := uint64(len(active_validator_indices)) / committee_count
	for i := uint64(0); i < committee_count; i += committee_size {
		commitees[i] = active_validator_indices[i : i+committee_size]
	}
	return commitees
}

// Return the block root at a recent slot.
func Get_block_root(state *beacon.BeaconState, slot eth2.Slot) (eth2.Root, error) {
	if !(slot < state.Slot && state.Slot <= slot+eth2.SLOTS_PER_HISTORICAL_ROOT) {
		return eth2.Root{}, errors.New("slot is not a recent slot, cannot find block root")
	}
	return state.Latest_block_roots[slot%eth2.SLOTS_PER_HISTORICAL_ROOT], nil
}

// Return the state root at a recent slot.
func Get_state_root(state *beacon.BeaconState, slot eth2.Slot) (eth2.Root, error) {
	if !(slot < state.Slot && state.Slot <= slot+eth2.SLOTS_PER_HISTORICAL_ROOT) {
		return eth2.Root{}, errors.New("slot is not a recent slot, cannot find state root")
	}
	return state.Latest_state_roots[slot%eth2.SLOTS_PER_HISTORICAL_ROOT], nil
}


func Is_validator_index(state *beacon.BeaconState, index eth2.ValidatorIndex) bool {
	return index < eth2.ValidatorIndex(len(state.Validator_registry))
}

// Slash the validator with index index.
func Slash_validator(state *beacon.BeaconState, index eth2.ValidatorIndex) error {
	validator := &state.Validator_registry[index]
	// [TO BE REMOVED IN PHASE 2] // TODO: add reasoning, spec unclear
	if state.Slot >= validator.Withdrawable_epoch.GetStartSlot() {
		return errors.New("cannot slash validator after withdrawal epoch")
	}
	Exit_validator(state, index)
	state.Latest_slashed_balances[state.Epoch()%eth2.LATEST_SLASHED_EXIT_LENGTH] += Get_effective_balance(state, index)

	whistleblower_reward := Get_effective_balance(state, index) / eth2.WHISTLEBLOWER_REWARD_QUOTIENT
	prop_index, err := Get_beacon_proposer_index(state, state.Slot, false)
	if err != nil {
		return err
	}
	state.Validator_balances[prop_index] += whistleblower_reward
	state.Validator_balances[index] -= whistleblower_reward
	validator.Slashed = true
	validator.Withdrawable_epoch = state.Epoch() + eth2.LATEST_SLASHED_EXIT_LENGTH
	return nil
}


// Get the domain number that represents the fork meta and signature domain.
func Get_domain(fork beacon.Fork, epoch eth2.Epoch, dom eth2.BLSDomain) eth2.BLSDomain {
	// combine fork version with domain.
	// TODO: spec is unclear about input size expectations.
	// TODO And is "+" different than packing into 64 bits here? I.e. ((32 bits fork version << 32) | (dom 32 bits))
	return eth2.BLSDomain(fork.GetVersion(epoch)<<32) + dom
}

// Return the beacon proposer index for the slot.
func Get_beacon_proposer_index(state *beacon.BeaconState, slot eth2.Slot, registryChange bool) (eth2.ValidatorIndex, error) {
	epoch := slot.ToEpoch()
	currentEpoch := slot.ToEpoch()
	if currentEpoch-1 <= epoch && epoch <= currentEpoch+1 {
		return 0, errors.New("epoch of given slot out of range")
	}
	// ignore error, slot input is trusted here
	committeeData, _ := Get_crosslink_committees_at_slot(state, slot, registryChange)
	first_committee_data := committeeData[0]
	return first_committee_data.Committee[slot%eth2.Slot(len(first_committee_data.Committee))], nil
}

func Get_epoch_committee_count(active_validator_count uint64) uint64 {
	return math.MaxU64(1,
		math.MinU64(
			uint64(eth2.SHARD_COUNT)/uint64(eth2.SLOTS_PER_EPOCH),
			active_validator_count/uint64(eth2.SLOTS_PER_EPOCH)/eth2.TARGET_COMMITTEE_SIZE,
		)) * uint64(eth2.SLOTS_PER_EPOCH)
}

// Return the number of committees in the previous epoch of the given ``state``.
func Get_previous_epoch_committee_count(state *beacon.BeaconState) uint64 {
	return Get_epoch_committee_count(Get_active_validator_count(
		state.Validator_registry,
		state.Previous_shuffling_epoch,
	))
}
// Return the number of committees in the current epoch of the given ``state``.
func Get_current_epoch_committee_count(state *beacon.BeaconState) uint64 {
	return Get_epoch_committee_count(Get_active_validator_count(
		state.Validator_registry,
		state.Current_shuffling_epoch,
	))
}
// Return the number of committees in the next epoch of the given ``state``.
func Get_next_epoch_committee_count(state *beacon.BeaconState) uint64 {
	return Get_epoch_committee_count(Get_active_validator_count(
		state.Validator_registry,
		state.Epoch() + 1,
	))
}
