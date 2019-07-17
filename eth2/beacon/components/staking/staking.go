package staking

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/meta"
	. "github.com/protolambda/zrnt/eth2/core"
)

type StakingReq interface {
	VersioningMeta
	RegistrySizeMeta
}

func ProcessEpochRewardsAndPenalties(meta StakingReq) {
	if meta.Epoch() == GENESIS_EPOCH {
		return
	}
	sum := NewDeltas(meta.ValidatorCount())
	sum.Add(state.AttestationDeltas())
	sum.Add(state.CrosslinksDeltas())
	meta.ApplyDeltas(sum)
}
