package registry

import . "github.com/protolambda/zrnt/eth2/core"

type Deltas struct {
	// element for each validator in registry
	Rewards []Gwei
	// element for each validator in registry
	Penalties []Gwei
}

func NewDeltas(validatorCount uint64) *Deltas {
	return &Deltas{
		Rewards:   make([]Gwei, validatorCount, validatorCount),
		Penalties: make([]Gwei, validatorCount, validatorCount),
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
