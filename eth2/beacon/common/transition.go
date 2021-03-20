package common

import (
	"context"
	"github.com/protolambda/ztyp/tree"
)

func ProcessSlot(ctx context.Context, _ *Spec, state BeaconState) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// The state root could take long, but absolute worst case is around a 1.5 seconds.
	// With any caching, this is more like < 50 ms. So no context use.
	// Cache state root
	previousStateRoot := state.HashTreeRoot(tree.GetHashFn())

	stateRootsBatch, err := state.StateRoots()
	if err != nil {
		return err
	}
	slot, err := state.Slot()
	if err != nil {
		return err
	}
	if err := stateRootsBatch.SetRoot(slot, previousStateRoot); err != nil {
		return err
	}

	latestHeader, err := state.LatestBlockHeader()
	if err != nil {
		return err
	}
	if latestHeader.StateRoot == (Root{}) {
		latestHeader.StateRoot = previousStateRoot
		if err := state.SetLatestBlockHeader(latestHeader); err != nil {
			return err
		}
	}
	previousBlockRoot := latestHeader.HashTreeRoot(tree.GetHashFn())

	// Cache latest known block and state root
	blockRootsBatch, err := state.BlockRoots()
	if err != nil {
		return err
	}
	if err := blockRootsBatch.SetRoot(slot, previousBlockRoot); err != nil {
		return err
	}

	return nil
}
