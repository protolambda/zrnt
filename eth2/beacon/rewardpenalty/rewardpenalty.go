package rewardpenalty

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type RewardsAndPenaltiesEpochProcess interface {
	ProcessEpochRewardsAndPenalties(input RewardsAndPenaltiesInput) error
}

type RewardsAndPenaltiesInput interface {
	meta.Versioning
	meta.RegistrySize
	meta.BalanceDeltas
	meta.AttestationDeltas
}

func ProcessEpochRewardsAndPenalties(input RewardsAndPenaltiesInput) error {
	currentEpoch, err := input.CurrentEpoch()
	if err != nil {
		return err
	}
	if currentEpoch == GENESIS_EPOCH {
		return nil
	}
	valCount, err := input.ValidatorCount()
	if err != nil {
		return err
	}
	sum := NewDeltas(valCount)
	attDeltas, err := input.AttestationDeltas()
	if err != nil {
		return err
	}
	sum.Add(attDeltas)
	return input.ApplyDeltas(sum)
}
