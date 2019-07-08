package components

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type Fork struct {
	// Previous fork version
	PreviousVersion ForkVersion
	// Current fork version
	CurrentVersion ForkVersion
	// Fork epoch number
	Epoch Epoch
}

type VersioningState struct {
	GenesisTime Timestamp
	Slot Slot
	Fork Fork
}

// Get current epoch
func (state *VersioningState) Epoch() Epoch {
	return state.Slot.ToEpoch()
}

// Return previous epoch.
func (state *VersioningState) PreviousEpoch() Epoch {
	currentEpoch := state.Epoch()
	if currentEpoch == GENESIS_EPOCH {
		return GENESIS_EPOCH
	} else {
		return currentEpoch - 1
	}
}

// Return the signature domain (fork version concatenated with domain type) of a message.
func (state *VersioningState) GetDomain(dom BLSDomainType, messageEpoch Epoch) BLSDomain {
	v := state.Fork.CurrentVersion
	if messageEpoch < state.Fork.Epoch {
		v = state.Fork.PreviousVersion
	}
	// combine fork version with domain type.
	return BLSDomain((uint64(v.ToUint32()) << 32) | uint64(dom))
}
