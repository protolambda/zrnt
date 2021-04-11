package states

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type DBStats struct {
	Count     int64
	LastWrite common.Root
}

type DB interface {
	// Store a state.
	// Returns exists=true if the state exists (previously), false otherwise. If error, it may not be accurate.
	Store(ctx context.Context, state common.BeaconState) (exists bool, err error)
	// Get a state. The state is a view of a shared immutable backing.
	// The view is save to mutate (it forks away from the original backing)
	Get(ctx context.Context, root common.Root) (state common.BeaconState, exists bool, err error)
	// Remove removes a state from the DB. Removing a state that does not exist is safe.
	// Returns exists=true if the state exists (previously), false otherwise. If error, it may not be accurate.
	Remove(root common.Root) (exists bool, err error)
	// Stats shows some database statistics such as latest write key and entry count.
	Stats() DBStats
	// List all known state roots
	List() []common.Root
	// Get Path
	Path() string
	// Spec of states
	Spec() *common.Spec
}
