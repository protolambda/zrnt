package deltas_computation

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/math"
	"sort"
)

func CrosslinksDeltas(state *BeaconState) *Deltas {
	deltas := NewDeltas(uint64(len(state.ValidatorRegistry)))

	totalActiveBalance := state.GetTotalBalanceOf(
		state.ValidatorRegistry.GetActiveValidatorIndices(state.Epoch()))

	totalBalanceSqRoot := Gwei(math.IntegerSquareroot(uint64(totalActiveBalance)))

	epoch := state.PreviousEpoch()
	count := Shard(state.GetEpochCommitteeCount(epoch))
	epochStartShard := state.GetEpochStartShard(epoch)
	for offset := Shard(0); offset < count; offset++ {
		shard := (epochStartShard + offset) % SHARD_COUNT

		crosslinkCommittee := state.GetCrosslinkCommittee(epoch, shard)
		committee := make(ValidatorSet, 0, len(crosslinkCommittee))
		committee = append(committee, crosslinkCommittee...)
		sort.Sort(committee)

		_, attestingIndices := state.GetWinningCrosslinkAndAttestingIndices(shard, epoch)
		attestingBalance := state.GetTotalBalanceOf(attestingIndices)
		totalBalance := state.GetTotalBalanceOf(committee)

		// reward/penalize using a zig-zag merge join.
		// ----------------------------------------------------
		committee.ZigZagJoin(attestingIndices,
			func(i ValidatorIndex) {
				// Committee member participated, reward them
				effectiveBalance := state.ValidatorRegistry[i].EffectiveBalance
				baseReward := effectiveBalance * BASE_REWARD_FACTOR / totalBalanceSqRoot / BASE_REWARDS_PER_EPOCH

				deltas.Rewards[i] += baseReward * attestingBalance / totalBalance
			}, func(i ValidatorIndex) {
				// Committee member did not participate, penalize them
				effectiveBalance := state.ValidatorRegistry[i].EffectiveBalance
				baseReward := effectiveBalance * BASE_REWARD_FACTOR / totalBalanceSqRoot / BASE_REWARDS_PER_EPOCH

				deltas.Penalties[i] += baseReward
			})
	}
	return deltas
}
