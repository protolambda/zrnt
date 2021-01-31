package pool

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"sync"
)

type VoluntaryExitPool struct {
	sync.RWMutex
	spec  *beacon.Spec
	exits map[beacon.ValidatorIndex]*beacon.SignedVoluntaryExit
}

func NewVoluntaryExitPool(spec *beacon.Spec) *VoluntaryExitPool {
	return &VoluntaryExitPool{
		spec:  spec,
		exits: make(map[beacon.ValidatorIndex]*beacon.SignedVoluntaryExit),
	}
}

func (vep *VoluntaryExitPool) AddVoluntaryExit(exit *beacon.SignedVoluntaryExit) (exists bool) {
	vep.Lock()
	defer vep.Unlock()
	key := exit.Message.ValidatorIndex
	if _, ok := vep.exits[key]; ok {
		return true
	}
	vep.exits[key] = exit
	return false
}
