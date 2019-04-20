package deltas_computation

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/math"
	"sort"
)

func DeltasCrosslinks(state *BeaconState) *Deltas {
	deltas := NewDeltas(uint64(len(state.ValidatorRegistry)))

	previousTotalBalance := state.GetTotalBalanceOf(
		state.ValidatorRegistry.GetActiveValidatorIndices(state.Epoch() - 1))

	adjustedQuotient := math.IntegerSquareroot(uint64(previousTotalBalance)) / BASE_REWARD_QUOTIENT

	// From previous epoch start, to current epoch start
	start := state.PreviousEpoch().GetStartSlot()
	end := state.Epoch().GetStartSlot()
	for slot := start; slot < end; slot++ {
		for _, shardCommittee := range state.GetCrosslinkCommitteesAtSlot(slot) {
			_, participants := state.GetWinningRootAndParticipants(shardCommittee.Shard)
			participatingBalance := state.GetTotalBalanceOf(participants)
			totalBalance := state.GetTotalBalanceOf(shardCommittee.Committee)

			committee := make(ValidatorSet, 0, len(shardCommittee.Committee))
			committee = append(committee, shardCommittee.Committee...)
			sort.Sort(committee)

			// reward/penalize using a zig-zag merge join.
			// ----------------------------------------------------
			committee.ZigZagJoin(participants,
				func(i ValidatorIndex) {
					// Committee member participated, reward them
					effectiveBalance := state.GetEffectiveBalance(i)
					baseReward := effectiveBalance / Gwei(adjustedQuotient) / 5

					deltas.Rewards[i] += baseReward * participatingBalance / totalBalance
				}, func(i ValidatorIndex) {
					// Committee member did not participate, penalize them
					effectiveBalance := state.GetEffectiveBalance(i)
					baseReward := effectiveBalance / Gwei(adjustedQuotient) / 5

					deltas.Penalties[i] += baseReward
				})
		}
	}
	return deltas
}
