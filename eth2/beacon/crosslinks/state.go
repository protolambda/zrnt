package crosslinks

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	"sort"
)

type CrosslinksState struct {
	CurrentCrosslinks  [SHARD_COUNT]Crosslink
	PreviousCrosslinks [SHARD_COUNT]Crosslink
}

var CrosslinkSSZ = zssz.GetSSZ((*Crosslink)(nil))

func (state *CrosslinksState) GetCurrentCrosslinkRoots() (out *[SHARD_COUNT]Root) {
	out = new([SHARD_COUNT]Root)
	for shard := Shard(0); shard < SHARD_COUNT; shard++ {
		out[shard] = ssz.HashTreeRoot(&state.CurrentCrosslinks[shard], CrosslinkSSZ)
	}
	return
}

func (state *CrosslinksState) GetPreviousCrosslinkRoots() (out *[SHARD_COUNT]Root) {
	out = new([SHARD_COUNT]Root)
	for shard := Shard(0); shard < SHARD_COUNT; shard++ {
		out[shard] = ssz.HashTreeRoot(&state.PreviousCrosslinks[shard], CrosslinkSSZ)
	}
	return
}

func (state *CrosslinksState) GetCurrentCrosslink(shard Shard) *Crosslink {
	return &state.CurrentCrosslinks[shard]
}

func (state *CrosslinksState) GetPreviousCrosslink(shard Shard) *Crosslink {
	return &state.PreviousCrosslinks[shard]
}

type CrosslinksFeature struct {
	*CrosslinksState
	Meta interface {
		VersioningMeta
		RegistrySizeMeta
		StakingMeta
		EffectiveBalanceMeta
		CrosslinkCommitteeMeta
		CommitteeCountMeta
		CrosslinkTimingMeta
		WinningCrosslinkMeta
	}
}

func (state *CrosslinksFeature) CrosslinkDeltas() *Deltas {
	deltas := NewDeltas(state.Meta.ValidatorCount())

	totalActiveBalance := state.Meta.GetTotalStakedBalance(state.Meta.CurrentEpoch())

	totalBalanceSqRoot := Gwei(math.IntegerSquareroot(uint64(totalActiveBalance)))

	epoch := state.Meta.PreviousEpoch()
	count := Shard(state.Meta.GetCommitteeCount(epoch))
	epochStartShard := state.Meta.GetStartShard(epoch)
	for offset := Shard(0); offset < count; offset++ {
		shard := (epochStartShard + offset) % SHARD_COUNT

		crosslinkCommittee := state.Meta.GetCrosslinkCommittee(epoch, shard)
		committee := make(ValidatorSet, 0, len(crosslinkCommittee))
		committee = append(committee, crosslinkCommittee...)
		sort.Sort(committee)

		_, attestingIndices := state.Meta.GetWinningCrosslinkAndAttesters(epoch, shard)
		attestingBalance := state.Meta.SumEffectiveBalanceOf(attestingIndices)
		totalBalance := state.Meta.SumEffectiveBalanceOf(committee)

		// reward/penalize using a zig-zag merge join.
		// ----------------------------------------------------
		committee.ZigZagJoin(attestingIndices,
			func(i ValidatorIndex) {
				// Committee member participated, reward them
				effectiveBalance := state.Meta.EffectiveBalance(i)
				baseReward := effectiveBalance * BASE_REWARD_FACTOR / totalBalanceSqRoot / BASE_REWARDS_PER_EPOCH

				deltas.Rewards[i] += baseReward * attestingBalance / totalBalance
			}, func(i ValidatorIndex) {
				// Committee member did not participate, penalize them
				effectiveBalance := state.Meta.EffectiveBalance(i)
				baseReward := effectiveBalance * BASE_REWARD_FACTOR / totalBalanceSqRoot / BASE_REWARDS_PER_EPOCH

				deltas.Penalties[i] += baseReward
			})
	}
	return deltas
}

func (state *CrosslinksFeature) ProcessEpochCrosslinks() {
	state.PreviousCrosslinks = state.CurrentCrosslinks
	currentEpoch := state.Meta.CurrentEpoch()
	previousEpoch := state.Meta.PreviousEpoch()
	for epoch := previousEpoch; epoch <= currentEpoch; epoch++ {
		count := state.Meta.GetCommitteeCount(epoch)
		startShard := state.Meta.GetStartShard(epoch)
		for offset := uint64(0); offset < count; offset++ {
			shard := (startShard + Shard(offset)) % SHARD_COUNT
			crosslinkCommittee := state.Meta.GetCrosslinkCommittee(epoch, shard)
			winningCrosslink, attestingIndices := state.Meta.GetWinningCrosslinkAndAttesters(epoch, shard)
			participatingBalance := state.Meta.SumEffectiveBalanceOf(attestingIndices)
			totalBalance := state.Meta.SumEffectiveBalanceOf(crosslinkCommittee)
			if 3*participatingBalance >= 2*totalBalance {
				state.CurrentCrosslinks[shard] = *winningCrosslink
			}
		}
	}
}
