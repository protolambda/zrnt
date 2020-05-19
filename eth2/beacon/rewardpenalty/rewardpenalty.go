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
	rewAndPenalties := f.Meta.AttestationRewardsAndPenalties()
	sum.Add(rewAndPenalties.Source)
	sum.Add(rewAndPenalties.Target)
	sum.Add(rewAndPenalties.Head)
	sum.Add(rewAndPenalties.InclusionDelay)
	sum.Add(rewAndPenalties.Inactivity)
	f.Meta.ApplyDeltas(sum)
}
