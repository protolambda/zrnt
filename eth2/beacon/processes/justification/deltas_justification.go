package justification

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/stake"
)

func DeltasJustification(state *beacon.BeaconState) *beacon.Deltas {
	return stake.NewDeltas(uint64(len(state.Validator_registry)))
	// TODO: implement justification rewards/penalties as deltas
	//// > Justification and finalization
	//{
	//	if epochs_since_finality <= 4 {
	//		// >> case 1: finality was not too long ago
	//
	//		// Slash validators that were supposed to be active, but did not do their work
	//		{
	//			//Justification-non-participation R-penalty
	//			applyRewardOrSlash(previous_active_validator_indices.Minus(previous_epoch_attester_indices), false, base_reward)
	//
	//			//Boundary-attestation-non-participation R-penalty
	//			applyRewardOrSlash(previous_active_validator_indices.Minus(previous_epoch_boundary_attester_indices), false, base_reward)
	//
	//			//Non-canonical-participation R-penalty
	//			applyRewardOrSlash(previous_active_validator_indices.Minus(previous_epoch_head_attester_indices), false, base_reward)
	//		}
	//
	//		// Reward active validators that do their work
	//		{
	//			// Justification-participation reward
	//			applyRewardOrSlash(previous_epoch_attester_indices, true,
	//				scaled_value(base_reward, previous_epoch_attesting_balance/previous_total_balance))
	//
	//			// Boundary-attestation reward
	//			applyRewardOrSlash(previous_epoch_boundary_attester_indices, true,
	//				scaled_value(base_reward, previous_epoch_boundary_attesting_balance/previous_total_balance))
	//
	//			// Canonical-participation reward
	//			applyRewardOrSlash(previous_epoch_head_attester_indices, true,
	//				scaled_value(base_reward, previous_epoch_head_attesting_balance/previous_total_balance))
	//
	//			// Attestation-Inclusion-delay reward: quicker = more reward
	//			applyRewardOrSlash(previous_epoch_attester_indices, true,
	//				scale_by_inclusion(scaled_value(base_reward, beacon.Gwei(beacon.MIN_ATTESTATION_INCLUSION_DELAY))))
	//		}
	//	} else {
	//		// >> case 2: more than 4 epochs since finality
	//
	//		// Slash validators that were supposed to be active, but did not do their work
	//		{
	//			// Justification-inactivity penalty
	//			applyRewardOrSlash(previous_active_validator_indices.Minus(previous_epoch_attester_indices), false, inactivity_penalty)
	//			// Boundary-attestation-Inactivity penalty
	//			applyRewardOrSlash(previous_active_validator_indices.Minus(previous_epoch_boundary_attester_indices), false, inactivity_penalty)
	//			// Non-canonical-participation R-penalty
	//			applyRewardOrSlash(previous_active_validator_indices.Minus(previous_epoch_head_attester_indices), false, base_reward)
	//			// Penalization measure: double inactivity penalty + R-penalty
	//			applyRewardOrSlash(previous_active_validator_indices, false, func(index beacon.ValidatorIndex) beacon.Gwei {
	//				if state.Validator_registry[index].Slashed {
	//					return (2 * inactivity_penalty(index)) + base_reward(index)
	//				}
	//				return 0
	//			})
	//		}
	//
	//		// Attestation delay measure
	//		{
	//			// Attestation-Inclusion-delay measure: less reward for long delays
	//			applyRewardOrSlash(previous_epoch_attester_indices, false, func(index beacon.ValidatorIndex) beacon.Gwei {
	//				return base_reward(index) - scale_by_inclusion(scaled_value(base_reward, beacon.Gwei(beacon.MIN_ATTESTATION_INCLUSION_DELAY)))(index)
	//			})
	//		}
	//	}
	//}
	//
	//// > Attestation inclusion
	//{
	//	// Attestations should be included timely.
	//	// TODO Difference from spec: it is easier (and faster) to iterate through the precomputed map
	//	for attester_index, att_index := range previous_epoch_earliest_attestations {
	//		proposer_index := Get_beacon_proposer_index(state, state.Latest_attestations[att_index].Inclusion_slot, false)
	//		state.Validator_balances[proposer_index] += base_reward(attester_index) / beacon.ATTESTATION_INCLUSION_REWARD_QUOTIENT
	//	}
	//}
}