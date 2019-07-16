package components

import (
	"encoding/binary"
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

type ShufflingData interface {
	GetCrosslinkCommittee(epoch Epoch, shard Shard) []ValidatorIndex
}

// Randomness and committees
type ShufflingState struct {
	LatestActiveIndexRoots [EPOCHS_PER_HISTORICAL_VECTOR]Root
}

// Epoch is expected to be between (current_epoch - EPOCHS_PER_HISTORICAL_VECTOR + ACTIVATION_EXIT_DELAY, current_epoch + ACTIVATION_EXIT_DELAY].
func (state *ShufflingState) GetActiveIndexRoot(epoch Epoch) Root {
	return state.LatestActiveIndexRoots[epoch%EPOCHS_PER_HISTORICAL_VECTOR]
}

func (state *BeaconState) UpdateActiveIndexRoot(epoch Epoch) {
	position := epoch % EPOCHS_PER_HISTORICAL_VECTOR
	indices := state.Validators.GetActiveValidatorIndices(epoch)
	state.LatestActiveIndexRoots[position] = ssz.HashTreeRoot(indices, RegistryIndicesSSZ)
}

// Generate a seed for the given epoch
func (state *BeaconState) GetSeed(epoch Epoch) Root {
	buf := make([]byte, 32*3)
	mix := state.GetRandomMix(epoch + EPOCHS_PER_HISTORICAL_VECTOR - MIN_SEED_LOOKAHEAD) // Avoid underflow
	copy(buf[0:32], mix[:])
	activeIndexRoot := state.GetActiveIndexRoot(epoch)
	copy(buf[32:64], activeIndexRoot[:])
	binary.LittleEndian.PutUint64(buf[64:], uint64(epoch))
	return Hash(buf)
}
