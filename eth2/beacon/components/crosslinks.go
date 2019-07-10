package components

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/zssz"
	"sort"
)

type CrosslinkingData interface {
	GetWinningCrosslinkAndAttesters(epoch Epoch, shard Shard) (*Crosslink, ValidatorSet)
}

var CrosslinkSSZ = zssz.GetSSZ((*Crosslink)(nil))

type Crosslink struct {
	Shard      Shard
	ParentRoot Root
	// Crosslinking data
	StartEpoch Epoch
	EndEpoch   Epoch
	DataRoot   Root
}

type CrosslinksState struct {
	CurrentCrosslinks  [SHARD_COUNT]Crosslink
	PreviousCrosslinks [SHARD_COUNT]Crosslink
}

func (state *BeaconState) CrosslinksDeltas() *Deltas {
	deltas := NewDeltas(uint64(len(state.Validators)))

	totalActiveBalance := state.Validators.GetTotalActiveEffectiveBalance(state.Epoch())

	totalBalanceSqRoot := Gwei(math.IntegerSquareroot(uint64(totalActiveBalance)))

	epoch := state.PreviousEpoch()
	count := Shard(state.Validators.GetEpochCommitteeCount(epoch))
	epochStartShard := state.GetEpochStartShard(epoch)
	for offset := Shard(0); offset < count; offset++ {
		shard := (epochStartShard + offset) % SHARD_COUNT

		crosslinkCommittee := state.PrecomputedData.GetCrosslinkCommittee(epoch, shard)
		committee := make(ValidatorSet, 0, len(crosslinkCommittee))
		committee = append(committee, crosslinkCommittee...)
		sort.Sort(committee)

		_, attestingIndices := state.PrecomputedData.GetWinningCrosslinkAndAttesters(epoch, shard)
		attestingBalance := state.Validators.GetTotalEffectiveBalanceOf(attestingIndices)
		totalBalance := state.Validators.GetTotalEffectiveBalanceOf(committee)

		// reward/penalize using a zig-zag merge join.
		// ----------------------------------------------------
		committee.ZigZagJoin(attestingIndices,
			func(i ValidatorIndex) {
				// Committee member participated, reward them
				effectiveBalance := state.Validators[i].EffectiveBalance
				baseReward := effectiveBalance * BASE_REWARD_FACTOR / totalBalanceSqRoot / BASE_REWARDS_PER_EPOCH

				deltas.Rewards[i] += baseReward * attestingBalance / totalBalance
			}, func(i ValidatorIndex) {
				// Committee member did not participate, penalize them
				effectiveBalance := state.Validators[i].EffectiveBalance
				baseReward := effectiveBalance * BASE_REWARD_FACTOR / totalBalanceSqRoot / BASE_REWARDS_PER_EPOCH

				deltas.Penalties[i] += baseReward
			})
	}
	return deltas
}
