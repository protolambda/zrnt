package rewardpenalty

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type RewardsAndPenaltiesEpochProcess interface {
	ProcessEpochRewardsAndPenalties()
}

type RewardsAndPenaltiesFeature struct {
	Meta interface {
		meta.Versioning
		meta.RegistrySize
		meta.BalanceDeltas
		meta.AttestationDeltas
	}
}

func (f *RewardsAndPenaltiesFeature) ProcessEpochRewardsAndPenalties() {
	if f.Meta.CurrentEpoch() == GENESIS_EPOCH {
		return
	}
	sum := NewDeltas(f.Meta.ValidatorCount())
	sum.Add(f.Meta.AttestationDeltas())
	f.Meta.ApplyDeltas(sum)
}
