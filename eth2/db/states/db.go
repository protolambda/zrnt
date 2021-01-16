package states

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon"
)

type DBStats struct {
	Count     int64
	LastWrite beacon.Root
}

type DB interface {
	// Store a state.
	// Returns exists=true if the state exists (previously), false otherwise. If error, it may not be accurate.
	Store(ctx context.Context, state *beacon.BeaconStateView) (exists bool, err error)
	// Get a state. The state is a view of a shared immutable backing.
	// The view is save to mutate (it forks away from the original backing)
	Get(ctx context.Context, root beacon.Root) (state *beacon.BeaconStateView, exists bool, err error)
	// Remove removes a state from the DB. Removing a state that does not exist is safe.
	// Returns exists=true if the state exists (previously), false otherwise. If error, it may not be accurate.
	Remove(root beacon.Root) (exists bool, err error)
	// Stats shows some database statistics such as latest write key and entry count.
	Stats() DBStats
	// List all known state roots
	List() []beacon.Root
	// Get Path
	Path() string
	// Spec of states
	Spec() *beacon.Spec
}
