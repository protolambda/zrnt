package pool

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"sync"
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

func (vep *VoluntaryExitPool) AddVoluntaryExit(exit *phase0.SignedVoluntaryExit) (exists bool) {
	vep.Lock()
	defer vep.Unlock()
	key := exit.Message.ValidatorIndex
	if _, ok := vep.exits[key]; ok {
		return true
	}
	vep.exits[key] = exit
	return false
}
