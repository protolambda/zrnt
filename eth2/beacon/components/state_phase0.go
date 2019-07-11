package components

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	"github.com/protolambda/zssz"
)

// TODO call .Load(state) on each of the data types to precompute
type BeaconData struct {
	ShufflingData
	CrosslinkingData
	ValidationData
}

var BeaconStateSSZ = zssz.GetSSZ((*BeaconState)(nil))

type BeaconState struct {
	PrecomputedData BeaconData `ssz:"omit"`
	VersioningState
	HistoryState
	Eth1State
	RegistryState
	ShardRotationState
	RandaoState
	ShufflingState
	CompactCommitteesState
	SlashingsState
	AttestationsState
	CrosslinksState
	FinalityState
}
