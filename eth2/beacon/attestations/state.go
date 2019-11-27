package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zssz"
)

type AttestationData struct {
	Slot  Slot
	Index CommitteeIndex

	// LMD GHOST vote
	BeaconBlockRoot Root

	// FFG vote
	Source Checkpoint
	Target Checkpoint
}

var AttestationDataType = &ContainerType{
	{"slot", SlotType},
	{"index", CommitteeIndexType},
	// LMD GHOST vote
	{"beacon_block_root", RootType},
	// FFG vote
	{"source", CheckpointType},
	{"target", CheckpointType},
}

var AttestationDataSSZ = zssz.GetSSZ((*AttestationData)(nil))

type PendingAttestation struct {
	AggregationBits CommitteeBits
	Data            AttestationData
	InclusionDelay  Slot
	ProposerIndex   ValidatorIndex
}

var PendingAttestationType = &ContainerType{
	{"aggregation_bits", CommitteeBitsType},
	{"data", AttestationDataType},
	{"inclusion_delay", SlotType},
	{"proposer_index", ValidatorIndexType},
}


type EpochPendingAttestations []*PendingAttestation

func (*EpochPendingAttestations) Limit() uint64 {
	return MAX_ATTESTATIONS * uint64(SLOTS_PER_EPOCH)
}

var PendingAttestationsType = ListType(PendingAttestationType, MAX_ATTESTATIONS*SLOTS_PER_EPOCH)

type AttestationsState struct {
	PreviousEpochAttestations EpochPendingAttestations
	CurrentEpochAttestations  EpochPendingAttestations
}

// Rotate current/previous epoch attestations
func (state *AttestationsState) RotateEpochAttestations() {
	state.PreviousEpochAttestations = state.CurrentEpochAttestations
	state.CurrentEpochAttestations = nil
}
