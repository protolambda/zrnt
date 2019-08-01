package active

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
)

type ActiveRootsFeature struct {
	State *ActiveState
	Meta  interface {
		ActiveValidatorCount
		ActiveIndices
	}
}

// Randomness and committees
type ActiveState struct {
	ActiveIndexRoots [EPOCHS_PER_HISTORICAL_VECTOR]Root
}

// CurrentEpoch is expected to be between (current_epoch - EPOCHS_PER_HISTORICAL_VECTOR + ACTIVATION_EXIT_DELAY, current_epoch + ACTIVATION_EXIT_DELAY].
func (state *ActiveState) GetActiveIndexRoot(epoch Epoch) Root {
	return state.ActiveIndexRoots[epoch%EPOCHS_PER_HISTORICAL_VECTOR]
}

func (f *ActiveRootsFeature) UpdateActiveIndexRoot(epoch Epoch) {
	position := epoch % EPOCHS_PER_HISTORICAL_VECTOR
	f.State.ActiveIndexRoots[position] = f.Meta.ComputeActiveIndexRoot(epoch)
}
