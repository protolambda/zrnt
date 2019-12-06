package rewardpenalty

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type RewardsAndPenaltiesEpochProcess interface {
	ProcessEpochRewardsAndPenalties() error
}

type RewardsAndPenaltiesFeature struct {
	Meta interface {
		meta.Versioning
		meta.RegistrySize
		meta.BalanceDeltas
		meta.AttestationDeltas
	}
}

func (f *RewardsAndPenaltiesFeature) ProcessEpochRewardsAndPenalties() error {
	currentEpoch, err := f.Meta.CurrentEpoch()
	if err != nil {
		return err
	}
	if currentEpoch == GENESIS_EPOCH {
		return nil
	}
	valCount, err := f.Meta.ValidatorCount()
	if err != nil {
		return err
	}
	sum := NewDeltas(valCount)
	attDeltas, err := f.Meta.AttestationDeltas()
	if err != nil {
		return err
	}
	sum.Add(attDeltas)
	return f.Meta.ApplyDeltas(sum)
}
