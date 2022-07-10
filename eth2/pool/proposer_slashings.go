package pool

import (
	"context"
	"fmt"
	"sync"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
)

type ProposerSlashingPool struct {
	sync.RWMutex
	spec      *common.Spec
	slashings map[common.ValidatorIndex]*phase0.ProposerSlashing
}

func NewProposerSlashingPool(spec *common.Spec) *ProposerSlashingPool {
	return &ProposerSlashingPool{
		spec:      spec,
		slashings: make(map[common.ValidatorIndex]*phase0.ProposerSlashing),
	}
}

func (psp *ProposerSlashingPool) AddProposerSlashing(ctx context.Context, sl *phase0.ProposerSlashing) error {
	psp.Lock()
	defer psp.Unlock()
	// maybe use pubkey instead?
	key := sl.SignedHeader1.Message.ProposerIndex
	if _, ok := psp.slashings[key]; ok {
		return fmt.Errorf("proposer %d is already getting slashed", key)
	}
	psp.slashings[key] = sl
	return nil
}

func (psp *ProposerSlashingPool) All() []*phase0.ProposerSlashing {
	psp.RLock()
	defer psp.RUnlock()
	out := make([]*phase0.ProposerSlashing, 0, len(psp.slashings))
	for _, a := range psp.slashings {
		out = append(out, a)
	}
	return out
}

// Pack n slashings, removes the slashings from the pool. A reward estimator is used to pick the best slashings.
// Slashings with negative rewards will not be packed.
func (psp *ProposerSlashingPool) Pack(estReward func(sl *phase0.ProposerSlashing) int, n uint) []*phase0.ProposerSlashing {
	// TODO
	return nil
}
