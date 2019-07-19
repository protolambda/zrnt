package beacon

import (
	. "github.com/protolambda/zrnt/eth2/beacon/attestations"
	. "github.com/protolambda/zrnt/eth2/beacon/compact"
	. "github.com/protolambda/zrnt/eth2/beacon/crosslinks"
	. "github.com/protolambda/zrnt/eth2/beacon/eth1"
	. "github.com/protolambda/zrnt/eth2/beacon/finality"
	. "github.com/protolambda/zrnt/eth2/beacon/header"
	. "github.com/protolambda/zrnt/eth2/beacon/history"
	. "github.com/protolambda/zrnt/eth2/beacon/randao"
	. "github.com/protolambda/zrnt/eth2/beacon/registry"
	. "github.com/protolambda/zrnt/eth2/beacon/shardrot"
	. "github.com/protolambda/zrnt/eth2/beacon/shuffling"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings"
	. "github.com/protolambda/zrnt/eth2/beacon/versioning"
	"github.com/protolambda/zssz"
)

var BeaconStateSSZ = zssz.GetSSZ((*BeaconState)(nil))

type BeaconState struct {
	VersioningState
	BlockHeaderState
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

// Run and attach precomputed status items
//func (status *Status) Load(state *BeaconState) {
//	status.ShufflingStatus.Load(state)
//	status.CrosslinkingStatus.Load(state, &status.ShufflingStatus)
//	status.ValidationStatus.Load(state, &status.ShufflingStatus)
//}
