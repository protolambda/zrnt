package shuffling

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/math"
)

// Randomness and committees
type ShufflingState struct {
	LatestActiveIndexRoots [EPOCHS_PER_HISTORICAL_VECTOR]Root
}

// CurrentEpoch is expected to be between (current_epoch - EPOCHS_PER_HISTORICAL_VECTOR + ACTIVATION_EXIT_DELAY, current_epoch + ACTIVATION_EXIT_DELAY].
func (state *ShufflingState) GetActiveIndexRoot(epoch Epoch) Root {
	return state.LatestActiveIndexRoots[epoch%EPOCHS_PER_HISTORICAL_VECTOR]
}

func (state *ShufflingState) UpdateActiveIndexRoot(meta ActiveIndexRootMeta, epoch Epoch) {
	position := epoch % EPOCHS_PER_HISTORICAL_VECTOR
	state.LatestActiveIndexRoots[position] = meta.GetActiveIndexRoot(epoch)
}

type CommitteeCountCalc struct {
	ActiveCountMeta ActiveValidatorCountMeta
}

// Return the number of committees in one epoch.
func (calc CommitteeCountCalc) GetCommitteeCount(epoch Epoch) uint64 {
	activeValidatorCount := calc.ActiveCountMeta.GetActiveValidatorCount(epoch)
	committeesPerSlot := math.MaxU64(1,
		math.MinU64(
			uint64(SHARD_COUNT)/uint64(SLOTS_PER_EPOCH),
			activeValidatorCount/uint64(SLOTS_PER_EPOCH)/TARGET_COMMITTEE_SIZE,
		))
	return committeesPerSlot * uint64(SLOTS_PER_EPOCH)
}
