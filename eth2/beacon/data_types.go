package beacon

import "github.com/protolambda/zrnt/eth2/util/hex"

// Beacon misc.
// ----------------------

type Shard uint64
type ValidatorIndex uint64
type DepositIndex uint64

// Beacon timing
// ----------------------

type Timestamp uint64

type Slot uint64

func (s Slot) ToEpoch() Epoch {
	return Epoch(s / SLOTS_PER_EPOCH)
}

type Epoch uint64

func (e Epoch) GetStartSlot() Slot {
	return Slot(e) * SLOTS_PER_EPOCH
}

// Return the epoch at which an activation or exit triggered in epoch takes effect.
func (e Epoch) GetDelayedActivationExitEpoch() Epoch {
	return e + 1 + ACTIVATION_EXIT_DELAY
}

// Value
// ----------------------

type Gwei uint64

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

// BLS specific
// ----------------------
type BLSDomain uint64

// Get the domain number that represents the fork meta and signature domain.
func GetDomain(fork Fork, epoch Epoch, dom BLSDomain) BLSDomain {
	// combine fork version with domain.
	v := fork.GetVersion(epoch)
	return BLSDomain(v[0]<<24|v[1]<<16|v[2]<<8|v[3]) + dom
}


// byte arrays
// ----------------------

type Root [32]byte
type Bytes32 [32]byte
type BLSPubkey [48]byte
type BLSSignature [96]byte

func (v Root) MarshalJSON() ([]byte, error)         { return hex.EncodeHex(v[:]) }
func (v Bytes32) MarshalJSON() ([]byte, error)      { return hex.EncodeHex(v[:]) }
func (v BLSPubkey) MarshalJSON() ([]byte, error)    { return hex.EncodeHex(v[:]) }
func (v BLSSignature) MarshalJSON() ([]byte, error) { return hex.EncodeHex(v[:]) }

func (v Root) UnmarshalJSON(data []byte) error         { return hex.DecodeHex(data[1:len(data)-1], v[:]) }
func (v Bytes32) UnmarshalJSON(data []byte) error      { return hex.DecodeHex(data[1:len(data)-1], v[:]) }
func (v BLSPubkey) UnmarshalJSON(data []byte) error    { return hex.DecodeHex(data[1:len(data)-1], v[:]) }
func (v BLSSignature) UnmarshalJSON(data []byte) error { return hex.DecodeHex(data[1:len(data)-1], v[:]) }

func (v Root) String() string         { return hex.EncodeHexStr(v[:]) }
func (v Bytes32) String() string      { return hex.EncodeHexStr(v[:]) }
func (v BLSPubkey) String() string    { return hex.EncodeHexStr(v[:]) }
func (v BLSSignature) String() string { return hex.EncodeHexStr(v[:]) }


// Collection of validators, should always be sorted.
type ValidatorSet []ValidatorIndex

func (vs ValidatorSet) Len() int {
	return len(vs)
}

func (vs ValidatorSet) Less(i int, j int) bool {
	return vs[i] < vs[j]
}

func (vs ValidatorSet) Swap(i int, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}

// Joins two validator sets:
// reports all indices of source that are in the target (call onIn), and those that are not (call onOut)
func (vs ValidatorSet) ZigZagJoin(target ValidatorSet, onIn func(i ValidatorIndex), onOut func(i ValidatorIndex)) {
	// index for source set side of the zig-zag
	i := 0
	// index for target set side of the zig-zag
	j := 0
	var iV, jV ValidatorIndex
	updateI := func() {
		// if out of bounds, just update to an impossibly high index
		if i < len(vs) {
			iV = vs[i]
		} else {
			iV = ValidatorIndexMarker
		}
	}
	updateJ := func() {
		// if out of bounds, just update to an impossibly high index
		if i < len(vs) {
			jV = target[j]
		} else {
			jV = ValidatorIndexMarker
		}
	}
	updateI()
	updateJ()
	for {
		// at some point all items in vs have been processed.
		if i >= len(vs) {
			break
		}
		if iV == jV {
			if onIn != nil {
				onIn(iV)
			}
			// go to next
			i++
			updateI()
			j++
			updateJ()
		} else if iV < jV {
			// if the index is lower than the current item in the target, it's not in the target.
			for x := i; x < j; x++ {
				if onOut != nil {
					onOut(iV)
				}
				// go to next
				i++
				updateI()
			}
		} else if iV > jV {
			// if the index is higher than the current item in the target, go to the next item.
			j++
			updateJ()
		}
	}
}
