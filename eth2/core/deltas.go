package core

import "github.com/protolambda/zssz"

type GweiList []Gwei

func (_ *GweiList) Limit() uint64 {
	return VALIDATOR_REGISTRY_LIMIT
}

type Deltas struct {
	Rewards   GweiList
	Penalties GweiList
}

var DeltasSSZ = zssz.GetSSZ((*Deltas)(nil))

func NewDeltas(validatorCount uint64) *Deltas {
	return &Deltas{
		Rewards:   make(GweiList, validatorCount, validatorCount),
		Penalties: make(GweiList, validatorCount, validatorCount),
	}
}

func (deltas *Deltas) Add(other *Deltas) {
	for i := 0; i < len(deltas.Rewards); i++ {
		deltas.Rewards[i] += other.Rewards[i]
	}
	for i := 0; i < len(deltas.Penalties); i++ {
		deltas.Penalties[i] += other.Penalties[i]
	}
}

type RewardsAndPenalties struct {
	Source         *Deltas
	Target         *Deltas
	Head           *Deltas
	InclusionDelay *Deltas
	Inactivity     *Deltas
}

func NewRewardsAndPenalties(validatorCount uint64) *RewardsAndPenalties {
	return &RewardsAndPenalties{
		Source:         NewDeltas(validatorCount),
		Target:         NewDeltas(validatorCount),
		Head:           NewDeltas(validatorCount),
		InclusionDelay: NewDeltas(validatorCount),
		Inactivity:     NewDeltas(validatorCount),
	}
}
