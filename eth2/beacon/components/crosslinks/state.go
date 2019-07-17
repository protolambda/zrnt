package crosslinks

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/meta"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/math"
	"sort"
)

type CrosslinksState struct {
	CurrentCrosslinks  [SHARD_COUNT]Crosslink
	PreviousCrosslinks [SHARD_COUNT]Crosslink
}

type CrosslinkDeltasReq interface {
	VersioningMeta
	RegistrySizeMeta
	StakingMeta
	CrosslinkMeta
	WinningCrosslinkMeta
}

func (state *CrosslinksState) CrosslinksDeltas(meta CrosslinkDeltasReq) *Deltas {
	deltas := NewDeltas(meta.ValidatorCount())

	totalActiveBalance := meta.GetTotalActiveEffectiveBalance(meta.Epoch())

	totalBalanceSqRoot := Gwei(math.IntegerSquareroot(uint64(totalActiveBalance)))

	epoch := meta.PreviousEpoch()
	count := Shard(meta.GetCommitteeCount(epoch))
	epochStartShard := meta.GetStartShard(epoch)
	for offset := Shard(0); offset < count; offset++ {
		shard := (epochStartShard + offset) % SHARD_COUNT

		crosslinkCommittee := meta.GetCrosslinkCommittee(epoch, shard)
		committee := make(ValidatorSet, 0, len(crosslinkCommittee))
		committee = append(committee, crosslinkCommittee...)
		sort.Sort(committee)

		_, attestingIndices := meta.GetWinningCrosslinkAndAttesters(epoch, shard)
		attestingBalance := meta.GetTotalEffectiveBalanceOf(attestingIndices)
		totalBalance := meta.GetTotalEffectiveBalanceOf(committee)

		// reward/penalize using a zig-zag merge join.
		// ----------------------------------------------------
		committee.ZigZagJoin(attestingIndices,
			func(i ValidatorIndex) {
				// Committee member participated, reward them
				effectiveBalance := meta.EffectiveBalance(i)
				baseReward := effectiveBalance * BASE_REWARD_FACTOR / totalBalanceSqRoot / BASE_REWARDS_PER_EPOCH

				deltas.Rewards[i] += baseReward * attestingBalance / totalBalance
			}, func(i ValidatorIndex) {
				// Committee member did not participate, penalize them
				effectiveBalance := meta.EffectiveBalance(i)
				baseReward := effectiveBalance * BASE_REWARD_FACTOR / totalBalanceSqRoot / BASE_REWARDS_PER_EPOCH

				deltas.Penalties[i] += baseReward
			})
	}
	return deltas
}
