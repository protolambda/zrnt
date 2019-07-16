package core

// Unix timestamp
type Timestamp uint64

func (t Timestamp) ToSlot(genesisTime Timestamp) Slot {
	return Slot((t - genesisTime) / SECONDS_PER_SLOT)
}

// Eth1 deposit ordering
type DepositIndex uint64

// Current slot
type Slot uint64

func (s Slot) ToEpoch() Epoch {
	return Epoch(s / SLOTS_PER_EPOCH)
}

type Epoch uint64

func (e Epoch) GetStartSlot() Slot {
	return Slot(e) * SLOTS_PER_EPOCH
}

// Return the epoch at which an activation or exit triggered in epoch takes effect.
func (e Epoch) ComputeActivationExitEpoch() Epoch {
	return e + 1 + ACTIVATION_EXIT_DELAY
}

func (e Epoch) Previous() Epoch {
	if e == GENESIS_EPOCH {
		return GENESIS_EPOCH
	} else {
		return e - 1
	}
}
