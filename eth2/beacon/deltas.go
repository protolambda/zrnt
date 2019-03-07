package beacon

type Deltas struct {
	// element for each validator in registry
	Rewards []Gwei
	// element for each validator in registry
	Penalties []Gwei
}

func NewDeltas(validatorCount uint64) *Deltas {
	return &Deltas{
		Rewards: make([]Gwei, validatorCount, validatorCount),
		Penalties: make([]Gwei, validatorCount, validatorCount),
	}
}

func (deltas *Deltas) Add(other *Deltas) {
	if len(deltas.Penalties) != len(other.Penalties) ||
		len(deltas.Rewards) != len(other.Rewards) {
		panic("cannot add other deltas, lengths do not match")
	}
	for i := 0; i < len(deltas.Rewards); i++ {
		deltas.Rewards[i] += other.Rewards[i]
	}
	for i := 0; i < len(deltas.Penalties); i++ {
		deltas.Penalties[i] += other.Penalties[i]
	}
}

type DeltasCalculator func(state *BeaconState) *Deltas

