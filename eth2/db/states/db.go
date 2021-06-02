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
	Store(ctx context.Context, state common.BeaconState) error
	// Get a state. The state is a view of a shared immutable backing.
	// The view is save to mutate (it forks away from the original backing)
	// Returns nil if the state does not exist.
	Get(ctx context.Context, root common.Root) (state common.BeaconState, err error)
	// Remove removes a state from the DB. Removing a state that does not exist is safe.
	Remove(root common.Root) error
}
