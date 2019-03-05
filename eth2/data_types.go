package eth2

import (
	"reflect"
	"unsafe"
)

type Slot uint64
type Epoch uint64
type Shard uint64
type Gwei uint64
type Timestamp uint64
type ValidatorIndex uint64
type DepositIndex uint64
type BLSDomain uint64

// byte arrays
type Root [32]byte
type Bytes32 [32]byte
type BLSPubkey [48]byte
type BLSSignature [96]byte

type ValueFunction func(index ValidatorIndex) Gwei


func (s Slot) ToEpoch() Epoch {
	return Epoch(s / SLOTS_PER_EPOCH)
}

func (e Epoch) GetStartSlot() Slot {
	return Slot(e) * SLOTS_PER_EPOCH
}

type ValidatorIndexList []ValidatorIndex

func (raw ValidatorIndexList) RawIndexSlice() []uint64 {
	// Get the slice header
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&raw))

	// The length and capacity of the slice are different.
	header.Len /= 8
	header.Cap /= 8

	// Convert slice header to an []uint64
	data := *(*[]uint64)(unsafe.Pointer(&header))
	return data
}
