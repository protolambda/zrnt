package beacon

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

// Return the epoch at which an activation or exit triggered in epoch takes effect.
func (e Epoch) GetDelayedActivationExitEpoch() Epoch {
	return e + 1 + ACTIVATION_EXIT_DELAY
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

func Max(a Gwei, b Gwei) Gwei {
	if a > b {
		return a
	}
	return b
}

func Min(a Gwei, b Gwei) Gwei {
	if a < b {
		return a
	}
	return b
}

// Get the domain number that represents the fork meta and signature domain.
func GetDomain(fork Fork, epoch Epoch, dom BLSDomain) BLSDomain {
	// combine fork version with domain.
	// TODO: spec is unclear about input size expectations.
	// TODO And is "+" different than packing into 64 bits here? I.e. ((32 bits fork version << 32) | (dom 32 bits))
	return BLSDomain(fork.GetVersion(epoch)<<32) + dom
}
