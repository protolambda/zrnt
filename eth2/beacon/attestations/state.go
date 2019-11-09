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

var AttestationDataSSZ = zssz.GetSSZ((*AttestationData)(nil))

type PendingAttestation struct {
	AggregationBits CommitteeBits
	Data            AttestationData
	InclusionDelay  Slot
	ProposerIndex   ValidatorIndex
}

type EpochPendingAttestations []*PendingAttestation

func (*EpochPendingAttestations) Limit() uint64 {
	return MAX_ATTESTATIONS * uint64(SLOTS_PER_EPOCH)
}

type AttestationsState struct {
	PreviousEpochAttestations EpochPendingAttestations
	CurrentEpochAttestations  EpochPendingAttestations
}

// Rotate current/previous epoch attestations
func (state *AttestationsState) RotateEpochAttestations() {
	state.PreviousEpochAttestations = state.CurrentEpochAttestations
	state.CurrentEpochAttestations = nil
}
