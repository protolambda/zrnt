package beacon

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
	v := fork.GetVersion(epoch)
	return BLSDomain(v[0] << 24 | v[1] << 16 | v[2] << 8 | v[3]) + dom
}
