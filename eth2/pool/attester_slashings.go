package pool

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/ztyp/tree"
	"sync"
)

type VersionedRoot struct {
	Root    beacon.Root
	Version beacon.Version
}

type AttesterSlashingPool struct {
	sync.RWMutex
	spec      *beacon.Spec
	slashings map[VersionedRoot]*beacon.AttesterSlashing
}

func NewAttesterSlashingPool(spec *beacon.Spec) *AttesterSlashingPool {
	return &AttesterSlashingPool{
		spec:      spec,
		slashings: make(map[VersionedRoot]*beacon.AttesterSlashing),
	}
}

// This does not filter slashings that are a subset of other slashings.
// The pool merely collects them. Make sure to protect against spam elsewhere as a caller.
func (asp *AttesterSlashingPool) AddAttesterSlashing(sl *beacon.AttesterSlashing, version beacon.Version) (exists bool) {
	root := sl.HashTreeRoot(asp.spec, tree.GetHashFn())
	asp.Lock()
	defer asp.Unlock()
	key := VersionedRoot{Root: root, Version: version}
	if _, ok := asp.slashings[key]; ok {
		return true
	}
	asp.slashings[key] = sl
	return false
}
