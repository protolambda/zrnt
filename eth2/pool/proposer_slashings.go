package pool

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"sync"
)

type ProposerSlashingPool struct {
	sync.RWMutex
	spec      *beacon.Spec
	slashings map[beacon.ValidatorIndex]*beacon.ProposerSlashing
}

func NewProposerSlashingPool(spec *beacon.Spec) *ProposerSlashingPool {
	return &ProposerSlashingPool{
		spec:      spec,
		slashings: make(map[beacon.ValidatorIndex]*beacon.ProposerSlashing),
	}
}

func (psp *ProposerSlashingPool) AddProposerSlashing(sl *beacon.ProposerSlashing) (exists bool) {
	psp.Lock()
	defer psp.Unlock()
	key := sl.SignedHeader1.Message.ProposerIndex
	if _, ok := psp.slashings[key]; ok {
		return true
	}
	psp.slashings[key] = sl
	return false
}
