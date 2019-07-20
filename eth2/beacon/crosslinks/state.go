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

func (state *CrosslinksState) GetCurrentCrosslinkRoots() (out [SHARD_COUNT]Root) {
	for shard := Shard(0); shard < SHARD_COUNT; shard++ {
		out[shard] = ssz.HashTreeRoot(&state.CurrentCrosslinks[shard], CrosslinkSSZ)
	}
	return
}

func (state *CrosslinksState) GetPreviousCrosslinkRoots() (out [SHARD_COUNT]Root) {
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

type CrosslinkDeltasReq interface {
	VersioningMeta
	RegistrySizeMeta
	StakingMeta
	EffectiveBalanceMeta
	CrosslinkCommitteeMeta
	CommitteeCountMeta
	CrosslinkTimingMeta
	WinningCrosslinkMeta
}

func (state *CrosslinksState) CrosslinksDeltas(meta CrosslinkDeltasReq) *Deltas {
	deltas := NewDeltas(meta.ValidatorCount())

	totalActiveBalance := meta.GetTotalStakedBalance(meta.Epoch())

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
		attestingBalance := meta.SumEffectiveBalanceOf(attestingIndices)
		totalBalance := meta.SumEffectiveBalanceOf(committee)

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

type CrosslinkingReq interface {
	VersioningMeta
	CrosslinkCommitteeMeta
	CommitteeCountMeta
	CrosslinkTimingMeta
	WinningCrosslinkMeta
	EffectiveBalanceMeta
	StakingMeta
}

func (state *CrosslinksState) ProcessEpochCrosslinks(meta CrosslinkingReq) {
	state.PreviousCrosslinks = state.CurrentCrosslinks
	currentEpoch := meta.Epoch()
	previousEpoch := meta.PreviousEpoch()
	for epoch := previousEpoch; epoch <= currentEpoch; epoch++ {
		count := meta.GetCommitteeCount(epoch)
		startShard := meta.GetStartShard(epoch)
		for offset := uint64(0); offset < count; offset++ {
			shard := (startShard + Shard(offset)) % SHARD_COUNT
			crosslinkCommittee := meta.GetCrosslinkCommittee(epoch, shard)
			winningCrosslink, attestingIndices := meta.GetWinningCrosslinkAndAttesters(epoch, shard)
			participatingBalance := meta.SumEffectiveBalanceOf(attestingIndices)
			totalBalance := meta.SumEffectiveBalanceOf(crosslinkCommittee)
			if 3*participatingBalance >= 2*totalBalance {
				state.CurrentCrosslinks[shard] = *winningCrosslink
			}
		}
	}
}
