package beacon

import (
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

// Unix timestamp
type Timestamp Uint64View

func (spec *Spec) TimeToSlot(t Timestamp, genesisTime Timestamp) Slot {
	if t < genesisTime {
		return 0
	}
	return Slot((t - genesisTime) / spec.SECONDS_PER_SLOT)
}

func (t Timestamp) HashTreeRoot(hFn tree.HashFn) Root {
	return Uint64View(t).HashTreeRoot(hFn)
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

func (i DepositIndex) HashTreeRoot(hFn tree.HashFn) Root {
	return Uint64View(i).HashTreeRoot(hFn)
}

const SlotType = Uint64Type

type Slot Uint64View

func (spec *Spec) SlotToEpoch(s Slot) Epoch {
	return Epoch(s / spec.SLOTS_PER_EPOCH)
}

func (s Slot) HashTreeRoot(hFn tree.HashFn) Root {
	return Uint64View(s).HashTreeRoot(hFn)
}

func AsSlot(v View, err error) (Slot, error) {
	i, err := AsUint64(v, err)
	return Slot(i), err
}

const EpochType = Uint64Type

type Epoch Uint64View

func (spec *Spec) EpochStartSlot(e Epoch) Slot {
	out := Slot(e) * spec.SLOTS_PER_EPOCH
	// check if it overflowed, saturate on max value if so.
	if e != spec.SlotToEpoch(out) {
		return ^Slot(0)
	} else {
		return out
	}
}

// Return the epoch at which an activation or exit triggered in epoch takes effect.
func (spec *Spec) ComputeActivationExitEpoch(e Epoch) Epoch {
	return e + 1 + spec.MAX_SEED_LOOKAHEAD
}

func (e Epoch) HashTreeRoot(hFn tree.HashFn) Root {
	return Uint64View(e).HashTreeRoot(hFn)
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
