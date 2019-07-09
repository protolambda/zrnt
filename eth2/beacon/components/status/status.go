package status

import . "github.com/protolambda/zrnt/eth2/beacon/components"

type LoadedStatus interface {
	Load(state *BeaconState)
}

type Status struct {
	ShufflingStatus
	CrosslinkingStatus
	ValidationStatus
}

func (status *Status) Load(state *BeaconState) {
	status.ShufflingStatus.Load(state)
	status.CrosslinkingStatus.Load(state, &status.ShufflingStatus)
	// TODO use shuffling status to optimize validation status
	status.ValidationStatus.Load(state, &status.ShufflingStatus)
}
