package block_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/math"
)

func ProcessProposerAttestationRewards(state *beacon.BeaconState, _ *beacon.BeaconBlock) error {
	proposerIndex := state.GetBeaconProposerIndex()

	earliestAttestations := make([]beacon.Slot, len(state.ValidatorRegistry))
	for i := 0; i < len(earliestAttestations); i++ {
		earliestAttestations[i] = ^beacon.Slot(0)
	}
	findEarliest := func(att *beacon.PendingAttestation) {
		participants, _ := state.GetAttestationParticipants(&att.Data, &att.AggregationBitfield)
		for _, p := range participants {
			if !state.ValidatorRegistry[p].Slashed {
				// If the attestation is the earliest:
				if earliestAttestations[p] > att.InclusionSlot {
					earliestAttestations[p] = att.InclusionSlot
				}
			}
		}
	}
	totalBalance := state.GetTotalBalanceOf(
		state.ValidatorRegistry.GetActiveValidatorIndices(state.Epoch()))
	adjustedQuotient := math.IntegerSquareroot(uint64(totalBalance)) / beacon.BASE_REWARD_QUOTIENT

	rewardProposersOfEarliest := func(att *beacon.PendingAttestation) {
		if adjustedQuotient != 0 {
			for i, slot := range earliestAttestations {
				if slot == state.Slot {
					effectiveBalance := state.GetEffectiveBalance(beacon.ValidatorIndex(i))
					baseReward := effectiveBalance / beacon.Gwei(adjustedQuotient) / 5
					state.IncreaseBalance(proposerIndex, baseReward / beacon.PROPOSER_REWARD_QUOTIENT)
				}
			}
		}
	}

	for _, att := range state.PreviousEpochAttestations {
		findEarliest(att)
	}
	for _, att := range state.CurrentEpochAttestations {
		findEarliest(att)
	}
	for _, att := range state.PreviousEpochAttestations {
		rewardProposersOfEarliest(att)
	}
	for _, att := range state.CurrentEpochAttestations {
		rewardProposersOfEarliest(att)
	}

	return nil
}
