package core

import "github.com/protolambda/zssz"

var AttestationDataSSZ = zssz.GetSSZ((*AttestationData)(nil))

type AttestationData struct {
	Slot  Slot
	Index CommitteeIndex

	// LMD GHOST vote
	BeaconBlockRoot Root

	// FFG vote
	Source Checkpoint
	Target Checkpoint
}

type PendingAttestation struct {
	AggregationBits CommitteeBits
	Data            AttestationData
	InclusionDelay  Slot
	ProposerIndex   ValidatorIndex
}
