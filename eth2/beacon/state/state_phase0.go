package state

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	"github.com/protolambda/zssz"
)

var BeaconStateSSZ = zssz.GetSSZ((*BeaconState)(nil))

type BeaconState struct {
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
