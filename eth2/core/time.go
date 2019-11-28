package core

import (
	"fmt"
	. "github.com/protolambda/ztyp/view"
)

type Uint64ReadProp ReadPropFn

func (p Uint64ReadProp) Uint64() (Uint64View, error) {
	v, err := p()
	if err != nil {
		return 0, err
	}
	n, ok := v.(Uint64View)
	if ok {
		return 0, fmt.Errorf("not a uint64 view: %v", v)
	}
	return n, nil
}

type Uint64WriteProp WritePropFn

func (p Uint64WriteProp) SetUint64(v Uint64View) error {
	return p(v)
}

// Unix timestamp
type Timestamp Uint64View

func (t Timestamp) ToSlot(genesisTime Timestamp) Slot {
	return Slot((t - genesisTime) / SECONDS_PER_SLOT)
}

type TimestampReadProp Uint64ReadProp

func (p TimestampReadProp) Timestamp() (Timestamp, error) {
	v, err := (Uint64ReadProp)(p).Uint64()
	return Timestamp(v), err
}

// Eth1 deposit ordering
type DepositIndex Uint64View

type DepositIndexReadProp Uint64ReadProp

func (p DepositIndexReadProp) DepositIndex() (DepositIndex, error) {
	v, err := (Uint64ReadProp)(p).Uint64()
	return DepositIndex(v), err
}

const SlotType = Uint64Type

type Slot Uint64View

func (s Slot) ToEpoch() Epoch {
	return Epoch(s / SLOTS_PER_EPOCH)
}

type SlotReadProp Uint64ReadProp

func (p SlotReadProp) Slot() (Slot, error) {
	v, err := (Uint64ReadProp)(p).Uint64()
	return Slot(v), err
}

type SlotWriteProp Uint64WriteProp

func (p SlotWriteProp) SetSlot(v Slot) error {
	return (Uint64WriteProp)(p).SetUint64(Uint64View(v))
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

type EpochReadProp Uint64ReadProp

func (p EpochReadProp) Epoch() (Epoch, error) {
	v, err := (Uint64ReadProp)(p).Uint64()
	return Epoch(v), err
}
