package pool

import (
	"context"
	"fmt"
	"sync"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/tree"
)

type AttesterSlashingPool struct {
	sync.RWMutex
	spec      *common.Spec
	slashings map[common.Root]*phase0.AttesterSlashing
}

func NewAttesterSlashingPool(spec *common.Spec) *AttesterSlashingPool {
	return &AttesterSlashingPool{
		spec:      spec,
		slashings: make(map[common.Root]*phase0.AttesterSlashing),
	}
}

// This does not filter slashings that are a subset of other slashings.
// The pool merely collects them. Make sure to protect against spam elsewhere as a caller.
func (asp *AttesterSlashingPool) AddAttesterSlashing(ctx context.Context, sl *phase0.AttesterSlashing) error {
	root := sl.HashTreeRoot(asp.spec, tree.GetHashFn())
	asp.Lock()
	defer asp.Unlock()
	if _, ok := asp.slashings[root]; ok {
		return fmt.Errorf("already have an attester slashing for message %s", root)
	}
	asp.slashings[root] = sl
	return nil
}

func (asp *AttesterSlashingPool) All() []*phase0.AttesterSlashing {
	asp.RLock()
	defer asp.RUnlock()
	out := make([]*phase0.AttesterSlashing, 0, len(asp.slashings))
	for _, a := range asp.slashings {
		out = append(out, a)
	}
	return out
}

// Pack n slashings, removes the slashings from the pool. A reward estimator is used to pick the best slashings.
// Slashings with negative rewards will not be packed.
func (asp *AttesterSlashingPool) Pack(estReward func(sl *phase0.AttesterSlashing) int, n uint) []*phase0.AttesterSlashing {
	// TODO
	return nil
}
