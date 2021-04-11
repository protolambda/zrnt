package pool

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"sync"
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

func (psp *ProposerSlashingPool) AddProposerSlashing(sl *phase0.ProposerSlashing) (exists bool) {
	psp.Lock()
	defer psp.Unlock()
	key := sl.SignedHeader1.Message.ProposerIndex
	if _, ok := psp.slashings[key]; ok {
		return true
	}
	psp.slashings[key] = sl
	return false
}
