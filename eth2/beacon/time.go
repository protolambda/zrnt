package beacon

import (
	. "github.com/protolambda/ztyp/view"
)

// Unix timestamp
type Timestamp Uint64View

func (t Timestamp) ToSlot(genesisTime Timestamp) Slot {
	return Slot((t - genesisTime) / SECONDS_PER_SLOT)
}

func AsTimestamp(v View, err error) (Timestamp, error) {
	i, err := AsUint64(v, err)
	return Timestamp(i), err
}

// Eth1 deposit ordering
type DepositIndex Uint64View

func AsDepositIndex(v View, err error) (DepositIndex, error) {
	i, err := AsUint64(v, err)
	return DepositIndex(i), err
}

const SlotType = Uint64Type

type Slot Uint64View

func (s Slot) ToEpoch() Epoch {
	return Epoch(s / SLOTS_PER_EPOCH)
}

func AsSlot(v View, err error) (Slot, error) {
	i, err := AsUint64(v, err)
	return Slot(i), err
}

const EpochType = Uint64Type

type Epoch Uint64View

func (e Epoch) GetStartSlot() Slot {
	return Slot(e) * SLOTS_PER_EPOCH
}

// Return the epoch at which an activation or exit triggered in epoch takes effect.
func (e Epoch) ComputeActivationExitEpoch() Epoch {
	return e + 1 + MAX_SEED_LOOKAHEAD
}

func (e Epoch) Previous() Epoch {
	if e == GENESIS_EPOCH {
		return GENESIS_EPOCH
	} else {
		return e - 1
	}
}

func AsEpoch(v View, err error) (Epoch, error) {
	i, err := AsUint64(v, err)
	return Epoch(i), err
}
