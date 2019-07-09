package components

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
	SlashingsState
	AttestationsState
	CrosslinksState
	FinalityState
}
