package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type AttestationData struct {
	// LMD GHOST vote
	BeaconBlockRoot Root

	// FFG vote
	Source Checkpoint
	Target Checkpoint

	// Crosslink vote
	Crosslink Crosslink
}

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

type AttestationSlotMeta interface {
	meta.CrosslinkTiming
	meta.CommitteeCount
}

func (attData *AttestationData) GetAttestationSlot(m AttestationSlotMeta) Slot {
	epoch := attData.Target.Epoch
	committeeCount := Slot(m.GetCommitteeCount(epoch))
	offset := Slot((attData.Crosslink.Shard + SHARD_COUNT - m.GetStartShard(epoch)) % SHARD_COUNT)
	return epoch.GetStartSlot() + (offset / (committeeCount / SLOTS_PER_EPOCH))
}
