package versioning

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type Fork struct {
	// Previous fork version
	PreviousVersion Version
	// Current fork version
	CurrentVersion Version
	// Fork epoch number
	Epoch Epoch
}

type VersioningState struct {
	GenesisTime           Timestamp
	GenesisValidatorsRoot Root
	Slot                  Slot
	Fork                  Fork
}

// Get current slot
func (state *VersioningState) CurrentSlot() Slot {
	return state.Slot
}

// Get current epoch
func (state *VersioningState) CurrentEpoch() Epoch {
	return state.Slot.ToEpoch()
}

// Return previous epoch.
func (state *VersioningState) PreviousEpoch() Epoch {
	return state.CurrentEpoch().Previous()
}

func (state *VersioningState) CurrentVersion() Version {
	return state.Fork.CurrentVersion
}

func (state *VersioningState) GenesisValRoot() Root {
	return state.GenesisValidatorsRoot
}

// Return the signature domain (fork version concatenated with domain type) of a message.
func (state *VersioningState) GetDomain(dom BLSDomainType, messageEpoch Epoch) BLSDomain {
	v := state.Fork.CurrentVersion
	if messageEpoch < state.Fork.Epoch {
		v = state.Fork.PreviousVersion
	}
	// combine fork version with domain type.
	return ComputeDomain(dom, v, state.GenesisValidatorsRoot)
}
