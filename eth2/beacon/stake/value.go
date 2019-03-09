package stake

// TODO: bunch of unused old code, as reference to what needs to be updated to new version

//// attestation-source-index for a given epoch, by validator index.
//// The earliest attestation (by inclusion_slot) is referenced in this map.
//previous_epoch_earliest_attestations := make(map[beacon.ValidatorIndex]uint64)
//for i, att := range state.Latest_attestations {
//	// error ignored, attestation is trusted.
//	participants, _ := Get_attestation_participants(state, &att.Data, &att.Aggregation_bitfield)
//	for _, participant := range participants {
//		if att.Data.Slot.ToEpoch() == previous_epoch {
//			if existingIndex, ok := previous_epoch_earliest_attestations[participant]; !ok || state.Latest_attestations[existingIndex].Inclusion_slot < att.Inclusion_slot {
//				previous_epoch_earliest_attestations[participant] = uint64(i)
//			}
//		}
//	}
//}

//base_reward_quotient := beacon.Gwei(math.Integer_squareroot(uint64(previous_total_balance))) / beacon.BASE_REWARD_QUOTIENT
//
//base_reward := func(index beacon.ValidatorIndex) beacon.Gwei {
//	// magic number 5 is from spec. (TODO add reasoning?)
//	return Get_effective_balance(state, index) / base_reward_quotient / 5
//}
//
//epochs_since_finality := next_epoch - state.Finalized_epoch
//
//inactivity_penalty := func(index beacon.ValidatorIndex) beacon.Gwei {
//	return base_reward(index) + (Get_effective_balance(state, index) * beacon.Gwei(epochs_since_finality) / beacon.INACTIVITY_PENALTY_QUOTIENT / 2)
//}
//
//scaled_value := func(valueFn beacon.ValueFunction, scale beacon.Gwei) beacon.ValueFunction {
//	return func(index beacon.ValidatorIndex) beacon.Gwei {
//		return valueFn(index) * scale
//	}
//}
//inclusion_distance := func(index beacon.ValidatorIndex) beacon.Slot {
//	a := &state.Latest_attestations[previous_epoch_earliest_attestations[index]]
//	return a.Inclusion_slot - a.Data.Slot
//}
//
//scale_by_inclusion := func(valueFn beacon.ValueFunction) beacon.ValueFunction {
//	return func(index beacon.ValidatorIndex) beacon.Gwei {
//		return valueFn(index) / beacon.Gwei(inclusion_distance(index))
//	}
//}
//
//// rewardOrSlash: true = reward, false = slash
//applyRewardOrSlash := func(indices validatorset.ValidatorIndexSet, rewardOrSlash bool, valueFn beacon.ValueFunction) {
//	for _, vIndex := range indices {
//		if rewardOrSlash {
//			state.Validator_balances[vIndex] += valueFn(vIndex)
//		} else {
//			state.Validator_balances[vIndex] -= valueFn(vIndex)
//		}
//	}
//}
