package pool

import (
	"context"
	"fmt"
	"sync"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
)

type VoluntaryExitPool struct {
	sync.RWMutex
	spec  *common.Spec
	exits map[common.ValidatorIndex]*phase0.SignedVoluntaryExit
}

func NewVoluntaryExitPool(spec *common.Spec) *VoluntaryExitPool {
	return &VoluntaryExitPool{
		spec:  spec,
		exits: make(map[common.ValidatorIndex]*phase0.SignedVoluntaryExit),
	}
}

func (vep *VoluntaryExitPool) AddVoluntaryExit(ctx context.Context, exit *phase0.SignedVoluntaryExit) error {
	vep.Lock()
	defer vep.Unlock()
	key := exit.Message.ValidatorIndex
	if _, ok := vep.exits[key]; ok {
		return fmt.Errorf("already have exit for validator %d", key)
	}
	vep.exits[key] = exit
	return nil
}

func (vep *VoluntaryExitPool) All() []*phase0.SignedVoluntaryExit {
	vep.RLock()
	defer vep.RUnlock()
	out := make([]*phase0.SignedVoluntaryExit, 0, len(vep.exits))
	for _, a := range vep.exits {
		out = append(out, a)
	}
	return out
}

// Pack n exits, removes the exits from the pool. A ranking function is used to pick the best exits.
// Exits with negative rank function outputs will not be packed.
func (vep *VoluntaryExitPool) Pack(rank func(sl *phase0.SignedVoluntaryExit) int, n uint) []*phase0.SignedVoluntaryExit {
	// TODO
	return nil
}
