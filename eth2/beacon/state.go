package beacon

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/attestations"
	. "github.com/protolambda/zrnt/eth2/beacon/components/compact"
	. "github.com/protolambda/zrnt/eth2/beacon/components/crosslinks"
	. "github.com/protolambda/zrnt/eth2/beacon/components/eth1"
	. "github.com/protolambda/zrnt/eth2/beacon/components/finality"
	. "github.com/protolambda/zrnt/eth2/beacon/components/header"
	. "github.com/protolambda/zrnt/eth2/beacon/components/history"
	. "github.com/protolambda/zrnt/eth2/beacon/components/randao"
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	. "github.com/protolambda/zrnt/eth2/beacon/components/shardrot"
	. "github.com/protolambda/zrnt/eth2/beacon/components/shuffling"
	. "github.com/protolambda/zrnt/eth2/beacon/components/slashings"
	. "github.com/protolambda/zrnt/eth2/beacon/components/versioning"
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
