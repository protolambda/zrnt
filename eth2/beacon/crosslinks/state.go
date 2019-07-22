package crosslinks

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
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
	State *CrosslinksState
	Meta  interface {
		meta.Versioning
		meta.RegistrySize
		meta.Staking
		meta.EffectiveBalances
		meta.CrosslinkCommittees
		meta.CommitteeCount
		meta.CrosslinkTiming
		meta.WinningCrosslinks
	}
}

func (f *CrosslinksFeature) CrosslinkDeltas() *Deltas {
	deltas := NewDeltas(f.Meta.ValidatorCount())

	totalActiveBalance := f.Meta.GetTotalStakedBalance(f.Meta.CurrentEpoch())

	totalBalanceSqRoot := Gwei(math.IntegerSquareroot(uint64(totalActiveBalance)))

	epoch := f.Meta.PreviousEpoch()
	count := Shard(f.Meta.GetCommitteeCount(epoch))
	epochStartShard := f.Meta.GetStartShard(epoch)
	for offset := Shard(0); offset < count; offset++ {
		shard := (epochStartShard + offset) % SHARD_COUNT

		crosslinkCommittee := f.Meta.GetCrosslinkCommittee(epoch, shard)
		committee := make(ValidatorSet, 0, len(crosslinkCommittee))
		committee = append(committee, crosslinkCommittee...)
		sort.Sort(committee)

		_, attestingIndices := f.Meta.GetWinningCrosslinkAndAttesters(epoch, shard)
		attestingBalance := f.Meta.SumEffectiveBalanceOf(attestingIndices)
		totalBalance := f.Meta.SumEffectiveBalanceOf(committee)

		// reward/penalize using a zig-zag merge join.
		// ----------------------------------------------------
		committee.ZigZagJoin(attestingIndices,
			func(i ValidatorIndex) {
				// Committee member participated, reward them
				effectiveBalance := f.Meta.EffectiveBalance(i)
				baseReward := effectiveBalance * BASE_REWARD_FACTOR / totalBalanceSqRoot / BASE_REWARDS_PER_EPOCH

				deltas.Rewards[i] += baseReward * attestingBalance / totalBalance
			}, func(i ValidatorIndex) {
				// Committee member did not participate, penalize them
				effectiveBalance := f.Meta.EffectiveBalance(i)
				baseReward := effectiveBalance * BASE_REWARD_FACTOR / totalBalanceSqRoot / BASE_REWARDS_PER_EPOCH

				deltas.Penalties[i] += baseReward
			})
	}
	return deltas
}

func (f *CrosslinksFeature) ProcessEpochCrosslinks() {
	f.State.PreviousCrosslinks = f.State.CurrentCrosslinks
	currentEpoch := f.Meta.CurrentEpoch()
	previousEpoch := f.Meta.PreviousEpoch()
	for epoch := previousEpoch; epoch <= currentEpoch; epoch++ {
		count := f.Meta.GetCommitteeCount(epoch)
		startShard := f.Meta.GetStartShard(epoch)
		for offset := uint64(0); offset < count; offset++ {
			shard := (startShard + Shard(offset)) % SHARD_COUNT
			crosslinkCommittee := f.Meta.GetCrosslinkCommittee(epoch, shard)
			winningCrosslink, attestingIndices := f.Meta.GetWinningCrosslinkAndAttesters(epoch, shard)
			participatingBalance := f.Meta.SumEffectiveBalanceOf(attestingIndices)
			totalBalance := f.Meta.SumEffectiveBalanceOf(crosslinkCommittee)
			if 3*participatingBalance >= 2*totalBalance {
				f.State.CurrentCrosslinks[shard] = *winningCrosslink
			}
		}
	}
}
