package phase0

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type SlotProcessFeature struct {
	Meta interface{
		StateRoot() Root
		meta.Versioning
		meta.LatestHeaderUpdate
		meta.HistoryUpdate
	}
}

func (f *SlotProcessFeature) ProcessSlot() {
	// Cache latest known state root (for previous slot)
	latestStateRoot := f.Meta.StateRoot()

	previousBlockRoot := f.Meta.UpdateLatestBlockRoot(latestStateRoot)

	// Cache latest known block and state root
	f.Meta.SetRecentRoots(f.Meta.CurrentSlot(), previousBlockRoot, latestStateRoot)
}
