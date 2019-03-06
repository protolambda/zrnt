package transition

import (
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/math"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
	"github.com/protolambda/go-beacon-transition/eth2/util/validatorset"
	"sort"
)

func EpochTransition(state *beacon.BeaconState) {
	current_epoch, previous_epoch := state.Epoch(), state.PreviousEpoch()
	next_epoch := current_epoch + 1

	// attestation-source-index for a given epoch, by validator index.
	// The earliest attestation (by inclusion_slot) is referenced in this map.
	previous_epoch_earliest_attestations := make(map[eth2.ValidatorIndex]uint64)
	for i, att := range state.Latest_attestations {
		// error ignored, attestation is trusted.
		participants, _ := Get_attestation_participants(state, &att.Data, &att.Aggregation_bitfield)
		for _, participant := range participants {
			if att.Data.Slot.ToEpoch() == previous_epoch {
				if existingIndex, ok := previous_epoch_earliest_attestations[participant]; !ok || state.Latest_attestations[existingIndex].Inclusion_slot < att.Inclusion_slot {
					previous_epoch_earliest_attestations[participant] = uint64(i)
				}
			}
		}
	}

	// Eth1 Data

	// Helper data
	// Note: Rewards and penalties are for participation in the previous epoch,
	//  so the "active validator" set is drawn from get_active calls on previous_epoch
	previous_active_validator_indices := validatorset.ValidatorIndexSet(Get_active_validator_indices(state.Validator_registry, previous_epoch))

	// Copy over the keys of our per-validator map to get a set of validator indices with previous epoch attestations.
	previous_epoch_attester_indices := make(validatorset.ValidatorIndexSet, 0, len(previous_epoch_earliest_attestations))
	for vIndex := range previous_epoch_earliest_attestations {
		previous_epoch_attester_indices = append(previous_epoch_attester_indices, vIndex)
	}
	previous_epoch_boundary_attester_indices, previous_epoch_head_attester_indices, current_epoch_boundary_attester_indices := make(validatorset.ValidatorIndexSet, 0), make(validatorset.ValidatorIndexSet, 0), make(validatorset.ValidatorIndexSet, 0)
	for _, att := range state.Latest_attestations {
		if ep := att.Data.Slot.ToEpoch(); ep == previous_epoch {

			boundary_block_root, err := Get_block_root(state, previous_epoch.GetStartSlot())
			isForBoundary := err == nil && att.Data.Epoch_boundary_root == boundary_block_root

			head_block_root, err := Get_block_root(state, att.Data.Slot)
			isForHead := err == nil && att.Data.Beacon_block_root == head_block_root

			// error ignored, attestation is trusted.
			participants, _ := Get_attestation_participants(state, &att.Data, &att.Aggregation_bitfield)
			for _, vIndex := range participants {

				// If the attestation is for a block boundary:
				if isForBoundary {
					previous_epoch_boundary_attester_indices = append(previous_epoch_boundary_attester_indices, vIndex)
				}

				if isForHead {
					previous_epoch_head_attester_indices = append(previous_epoch_head_attester_indices, vIndex)
				}
			}
		} else if ep == current_epoch {
			boundary_block_root, err := Get_block_root(state, current_epoch.GetStartSlot())
			isForBoundary := err == nil && att.Data.Epoch_boundary_root == boundary_block_root
			// error ignored, attestation is trusted.
			participants, _ := Get_attestation_participants(state, &att.Data, &att.Aggregation_bitfield)
			for _, vIndex := range participants {
				// If the attestation is for a block boundary:
				if isForBoundary {
					current_epoch_boundary_attester_indices = append(current_epoch_boundary_attester_indices, vIndex)
				}
			}
		}
	}

	// Justification and finalization
	{
		previous_epoch_boundary_attesting_balance := Get_total_balance(state, previous_epoch_boundary_attester_indices)
		current_epoch_boundary_attesting_balance := Get_total_balance(state, current_epoch_boundary_attester_indices)
		previous_total_balance := Get_total_balance(state, Get_active_validator_indices(state.Validator_registry, previous_epoch))
		current_total_balance := Get_total_balance(state, Get_active_validator_indices(state.Validator_registry, current_epoch))

		// > Justification
		new_justified_epoch := state.Justified_epoch
		state.Justification_bitfield = state.Justification_bitfield << 1
		if 3*previous_epoch_boundary_attesting_balance >= 2*previous_total_balance {
			state.Justification_bitfield |= 2
			new_justified_epoch = previous_epoch
		}
		if 3*current_epoch_boundary_attesting_balance >= 2*current_total_balance {
			state.Justification_bitfield |= 1
			new_justified_epoch = current_epoch
		}
		// > Finalization
		if (state.Justification_bitfield>>1)&7 == 7 && state.Previous_justified_epoch == previous_epoch-2 {
			state.Finalized_epoch = state.Previous_justified_epoch
		}
		if (state.Justification_bitfield>>1)&3 == 3 && state.Previous_justified_epoch == previous_epoch-1 {
			state.Finalized_epoch = state.Previous_justified_epoch
		}
		if (state.Justification_bitfield>>0)&7 == 7 && state.Justified_epoch == previous_epoch-1 {
			state.Finalized_epoch = state.Justified_epoch
		}
		if (state.Justification_bitfield>>0)&3 == 3 && state.Justified_epoch == previous_epoch {
			state.Finalized_epoch = state.Justified_epoch
		}
		// > Final part
		state.Previous_justified_epoch = state.Justified_epoch
		state.Justified_epoch = new_justified_epoch
	}

	// All recent winning crosslinks, regardless of weight.
	winning_roots := make(map[eth2.Shard]eth2.Root)
	// Remember the attesters of each winning crosslink root (1 per shard)
	// Also includes non-persisted winners (i.e. winning attesters not bigger than 2/3 of total committee weight)
	crosslink_winners := make(map[eth2.Root]validatorset.ValidatorIndexSet)
	crosslink_winners_weight := make(map[eth2.Root]eth2.Gwei)

	// Crosslinks
	{

		start, end := previous_epoch.GetStartSlot(), next_epoch.GetStartSlot()
		for slot := start; slot < end; slot++ {
			// epoch is trusted, ignore error
			crosslink_committees_at_slot, _ := Get_crosslink_committees_at_slot(state, slot, false)
			for _, cross_comm := range crosslink_committees_at_slot {

				// The spec is insane in making everything a helper function, ignoring scope/encapsulation, and not being to-the-point.
				// All we need is to determine a crosslink root,
				//  "winning_root" (from all attestations in previous or current epoch),
				//  and keep track of its weight.
				crosslink_data_root := state.Latest_crosslinks[cross_comm.Shard].Crosslink_data_root

				// First look at all attestations, and sum the weights per root.
				weightedCrosslinks := make(map[eth2.Root]eth2.Gwei)
				for _, att := range state.Latest_attestations {
					if ep := att.Data.Slot.ToEpoch(); ep == previous_epoch || ep == current_epoch &&
						att.Data.Shard == cross_comm.Shard &&
						att.Data.Crosslink_data_root == crosslink_data_root {
						// error ignored, attestation is trusted.
						participants, _ := Get_attestation_participants(state, &att.Data, &att.Aggregation_bitfield)
						for _, participant := range participants {
							weightedCrosslinks[att.Data.Crosslink_data_root] += Get_effective_balance(state, participant)
						}
					}
				}
				// Now determine the best root, by weight
				var winning_root eth2.Root
				winning_weight := eth2.Gwei(0)
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
				// we need to remember attesters of winning root (for later rewarding, and exclusion to slashing)
				winning_attesting_committee_members := make(validatorset.ValidatorIndexSet, 0)
				for _, att := range state.Latest_attestations {
					if ep := att.Data.Slot.ToEpoch(); ep == previous_epoch || ep == current_epoch &&
						att.Data.Shard == cross_comm.Shard &&
						att.Data.Crosslink_data_root == winning_root {
						// error ignored, attestation is trusted.
						participants, _ := Get_attestation_participants(state, &att.Data, &att.Aggregation_bitfield)
						for _, participant := range participants {
							for _, vIndex := range cross_comm.Committee {
								if participant == vIndex {
									winning_attesting_committee_members = append(winning_attesting_committee_members, vIndex)
								}
							}
						}
					}
				}
				crosslink_winners[winning_root] = winning_attesting_committee_members
				winning_roots[cross_comm.Shard] = winning_root
				crosslink_winners_weight[winning_root] = winning_weight

				// If it has sufficient weight, the crosslink is accepted.
				if 3*winning_weight >= 2*Get_total_balance(state, cross_comm.Committee) {
					state.Latest_crosslinks[cross_comm.Shard] = beacon.Crosslink{
						Epoch:               slot.ToEpoch(),
						Crosslink_data_root: winning_root}
				}
			}
		}
	}

	// Rewards & Penalties
	{
		// Sum balances of the sets of validators from earlier
		previous_epoch_attesting_balance := Get_total_balance(state, previous_epoch_attester_indices)
		previous_epoch_boundary_attesting_balance := Get_total_balance(state, previous_epoch_boundary_attester_indices)
		previous_epoch_head_attesting_balance := Get_total_balance(state, previous_epoch_head_attester_indices)

		// Note: previous_total_balance and previous_epoch_boundary_attesting_balance balance might be marginally
		// different than the actual balances during previous epoch transition.
		// Due to the tight bound on validator churn each epoch and small per-epoch rewards/penalties,
		// the potential balance difference is very low and only marginally affects consensus safety.
		previous_total_balance := Get_total_balance(state, Get_active_validator_indices(state.Validator_registry, previous_epoch))

		base_reward_quotient := eth2.Gwei(math.Integer_squareroot(uint64(previous_total_balance))) / eth2.BASE_REWARD_QUOTIENT

		base_reward := func(index eth2.ValidatorIndex) eth2.Gwei {
			// magic number 5 is from spec. (TODO add reasoning?)
			return Get_effective_balance(state, index) / base_reward_quotient / 5
		}

		epochs_since_finality := next_epoch - state.Finalized_epoch

		inactivity_penalty := func(index eth2.ValidatorIndex) eth2.Gwei {
			return base_reward(index) + (Get_effective_balance(state, index) * eth2.Gwei(epochs_since_finality) / eth2.INACTIVITY_PENALTY_QUOTIENT / 2)
		}

		scaled_value := func(valueFn eth2.ValueFunction, scale eth2.Gwei) eth2.ValueFunction {
			return func(index eth2.ValidatorIndex) eth2.Gwei {
				return valueFn(index) * scale
			}
		}
		inclusion_distance := func(index eth2.ValidatorIndex) eth2.Slot {
			a := &state.Latest_attestations[previous_epoch_earliest_attestations[index]]
			return a.Inclusion_slot - a.Data.Slot
		}

		scale_by_inclusion := func(valueFn eth2.ValueFunction) eth2.ValueFunction {
			return func(index eth2.ValidatorIndex) eth2.Gwei {
				return valueFn(index) / eth2.Gwei(inclusion_distance(index))
			}
		}

		// rewardOrSlash: true = reward, false = slash
		applyRewardOrSlash := func(indices validatorset.ValidatorIndexSet, rewardOrSlash bool, valueFn eth2.ValueFunction) {
			for _, vIndex := range indices {
				if rewardOrSlash {
					state.Validator_balances[vIndex] += valueFn(vIndex)
				} else {
					state.Validator_balances[vIndex] -= valueFn(vIndex)
				}
			}
		}

		// > Justification and finalization
		{
			if epochs_since_finality <= 4 {
				// >> case 1: finality was not too long ago

				// Slash validators that were supposed to be active, but did not do their work
				{
					//Justification-non-participation R-penalty
					applyRewardOrSlash(previous_active_validator_indices.Minus(previous_epoch_attester_indices), false, base_reward)

					//Boundary-attestation-non-participation R-penalty
					applyRewardOrSlash(previous_active_validator_indices.Minus(previous_epoch_boundary_attester_indices), false, base_reward)

					//Non-canonical-participation R-penalty
					applyRewardOrSlash(previous_active_validator_indices.Minus(previous_epoch_head_attester_indices), false, base_reward)
				}

				// Reward active validators that do their work
				{
					// Justification-participation reward
					applyRewardOrSlash(previous_epoch_attester_indices, true,
						scaled_value(base_reward, previous_epoch_attesting_balance/previous_total_balance))

					// Boundary-attestation reward
					applyRewardOrSlash(previous_epoch_boundary_attester_indices, true,
						scaled_value(base_reward, previous_epoch_boundary_attesting_balance/previous_total_balance))

					// Canonical-participation reward
					applyRewardOrSlash(previous_epoch_head_attester_indices, true,
						scaled_value(base_reward, previous_epoch_head_attesting_balance/previous_total_balance))

					// Attestation-Inclusion-delay reward: quicker = more reward
					applyRewardOrSlash(previous_epoch_attester_indices, true,
						scale_by_inclusion(scaled_value(base_reward, eth2.Gwei(eth2.MIN_ATTESTATION_INCLUSION_DELAY))))
				}
			} else {
				// >> case 2: more than 4 epochs since finality

				// Slash validators that were supposed to be active, but did not do their work
				{
					// Justification-inactivity penalty
					applyRewardOrSlash(previous_active_validator_indices.Minus(previous_epoch_attester_indices), false, inactivity_penalty)
					// Boundary-attestation-Inactivity penalty
					applyRewardOrSlash(previous_active_validator_indices.Minus(previous_epoch_boundary_attester_indices), false, inactivity_penalty)
					// Non-canonical-participation R-penalty
					applyRewardOrSlash(previous_active_validator_indices.Minus(previous_epoch_head_attester_indices), false, base_reward)
					// Penalization measure: double inactivity penalty + R-penalty
					applyRewardOrSlash(previous_active_validator_indices, false, func(index eth2.ValidatorIndex) eth2.Gwei {
						if state.Validator_registry[index].Slashed {
							return (2 * inactivity_penalty(index)) + base_reward(index)
						}
						return 0
					})
				}

				// Attestation delay measure
				{
					// Attestation-Inclusion-delay measure: less reward for long delays
					applyRewardOrSlash(previous_epoch_attester_indices, false, func(index eth2.ValidatorIndex) eth2.Gwei {
						return base_reward(index) - scale_by_inclusion(scaled_value(base_reward, eth2.Gwei(eth2.MIN_ATTESTATION_INCLUSION_DELAY)))(index)
					})
				}
			}
		}

		// > Attestation inclusion
		{
			// Attestations should be included timely.
			// TODO Difference from spec: it is easier (and faster) to iterate through the precomputed map
			for attester_index, att_index := range previous_epoch_earliest_attestations {
				proposer_index := Get_beacon_proposer_index(state, state.Latest_attestations[att_index].Inclusion_slot, false)
				state.Validator_balances[proposer_index] += base_reward(attester_index) / eth2.ATTESTATION_INCLUSION_REWARD_QUOTIENT
			}
		}

		// > Crosslinks
		{
			// Crosslinks should be created by the committees
			start, end := previous_epoch.GetStartSlot(), next_epoch.GetStartSlot()
			for slot := start; slot < end; slot++ {
				// epoch is trusted, ignore error
				crosslink_committees_at_slot, _ := Get_crosslink_committees_at_slot(state, slot, false)
				for _, cross_comm := range crosslink_committees_at_slot {

					// We remembered the winning root
					// (i.e. the most attested crosslink root, doesn't have to be 2/3 majority)
					winning_root := winning_roots[cross_comm.Shard]

					// We remembered the attesters of the crosslink
					crosslink_attesters := crosslink_winners[winning_root]

					// Note: non-committee validators still count as attesters for a crosslink,
					//  hence the extra work to filter for just the validators in the committee
					committee_non_participants := validatorset.ValidatorIndexSet(cross_comm.Committee).Minus(crosslink_attesters)

					committee_attesters_weight := crosslink_winners_weight[winning_root]
					total_committee_weight := Get_total_balance(state, cross_comm.Committee)

					// Reward those that contributed to finding a winning root.
					applyRewardOrSlash(validatorset.ValidatorIndexSet(cross_comm.Committee).Minus(committee_non_participants),
						true, func(index eth2.ValidatorIndex) eth2.Gwei {
							return base_reward(index) * committee_attesters_weight / total_committee_weight
						})
					// Slash those that opted for a different crosslink
					applyRewardOrSlash(committee_non_participants, false, base_reward)
				}
			}
		}

		// > Ejections
		{
			// After we are done slashing, eject the validators that don't have enough balance left.
			for _, vIndex := range Get_active_validator_indices(state.Validator_registry, current_epoch) {
				if state.Validator_balances[vIndex] < eth2.EJECTION_BALANCE {
					Exit_validator(state, vIndex)
				}
			}
		}
	}
	// Validator registry and shuffling data
	{
		// > update registry
		{
			state.Previous_shuffling_epoch = state.Current_shuffling_epoch
			state.Previous_shuffling_start_shard = state.Current_shuffling_start_shard
			state.Previous_shuffling_seed = state.Current_shuffling_seed

			if state.Finalized_epoch > state.Validator_registry_update_epoch {
				needsUpdate := true
				{
					committee_count := Get_epoch_committee_count(Get_active_validator_count(state.Validator_registry, current_epoch))
					for i := uint64(0); i < committee_count; i++ {
						if shard := (state.Current_shuffling_start_shard + eth2.Shard(i)) % eth2.SHARD_COUNT; state.Latest_crosslinks[shard].Epoch <= state.Validator_registry_update_epoch {
							needsUpdate = false
						}
					}
				}
				if needsUpdate {
					Update_validator_registry(state)
					state.Current_shuffling_epoch = next_epoch
					// recompute committee count, some validators may not be active anymore due to the above update.
					committee_count := Get_epoch_committee_count(Get_active_validator_count(state.Validator_registry, current_epoch))
					state.Current_shuffling_start_shard = (state.Current_shuffling_start_shard + eth2.Shard(committee_count)) % eth2.SHARD_COUNT
					// ignore error, current_shuffling_epoch is a trusted input
					state.Current_shuffling_seed = Generate_seed(state, state.Current_shuffling_epoch)
				} else {
					// If a validator registry update does not happen:
					epochs_since_last_registry_update := current_epoch - state.Validator_registry_update_epoch
					if epochs_since_last_registry_update > 1 && math.Is_power_of_two(uint64(epochs_since_last_registry_update)) {
						state.Current_shuffling_epoch = next_epoch
						// Note that state.Current_shuffling_start_shard is left unchanged
						state.Current_shuffling_seed = Generate_seed(state, state.Current_shuffling_epoch)
					}
				}
			}
		}

		// > process slashings
		{
			active_validator_indices := Get_active_validator_indices(state.Validator_registry, current_epoch)
			total_balance := Get_total_balance(state, active_validator_indices)

			for index, validator := range state.Validator_registry {
				if validator.Slashed &&
					current_epoch == validator.Withdrawable_epoch-(eth2.LATEST_SLASHED_EXIT_LENGTH/2) {
					epoch_index := current_epoch % eth2.LATEST_SLASHED_EXIT_LENGTH
					total_at_start := state.Latest_slashed_balances[(epoch_index+1)%eth2.LATEST_SLASHED_EXIT_LENGTH]
					total_at_end := state.Latest_slashed_balances[epoch_index]
					balance := Get_effective_balance(state, eth2.ValidatorIndex(index))
					state.Validator_balances[index] -= math.Max(balance*math.Min((total_at_end-total_at_start)*3, total_balance)/total_balance, balance/eth2.MIN_PENALTY_QUOTIENT)
				}
			}
		}

		// > process exit queue
		{
			eligible_indices := make(validatorset.ValidatorIndexSet, 0)
			for index, validator := range state.Validator_registry {
				if validator.Withdrawable_epoch != eth2.FAR_FUTURE_EPOCH && current_epoch > validator.Exit_epoch+eth2.MIN_VALIDATOR_WITHDRAWABILITY_DELAY {
					eligible_indices = append(eligible_indices, eth2.ValidatorIndex(index))
				}
			}
			// Sort in order of exit epoch, and validators that exit within the same epoch exit in order of validator index
			sort.Slice(eligible_indices, func(i int, j int) bool {
				return state.Validator_registry[eligible_indices[i]].Exit_epoch < state.Validator_registry[eligible_indices[j]].Exit_epoch
			})
			// eligible_indices is sorted here (in-place sorting)
			for i, end := uint64(0), uint64(len(eligible_indices)); i < eth2.MAX_EXIT_DEQUEUES_PER_EPOCH && i < end; i++ {
				Prepare_validator_for_withdrawal(state, eligible_indices[i])
			}
		}

		// > final updates
		{
			state.Latest_active_index_roots[(next_epoch+eth2.ACTIVATION_EXIT_DELAY)%eth2.LATEST_ACTIVE_INDEX_ROOTS_LENGTH] = ssz.Hash_tree_root(Get_active_validator_indices(state.Validator_registry, next_epoch+eth2.ACTIVATION_EXIT_DELAY))
			state.Latest_slashed_balances[next_epoch%eth2.LATEST_SLASHED_EXIT_LENGTH] = state.Latest_slashed_balances[current_epoch%eth2.LATEST_SLASHED_EXIT_LENGTH]

			// TODO randao

			// Remove any attestation in state.Latest_attestations such that slot_to_epoch(attestation.Data.Slot) < current_epoch
			attests := make([]beacon.PendingAttestation, 0)
			for _, a := range state.Latest_attestations {
				// only keep recent attestations. (for next epoch to process)
				if a.Data.Slot.ToEpoch() >= current_epoch {
					attests = append(attests, a)
				}
			}
			state.Latest_attestations = attests
		}
	}
}
