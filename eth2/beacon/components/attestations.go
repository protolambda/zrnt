package components

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bitfield"
	"github.com/protolambda/zrnt/eth2/util/shuffling"
	"github.com/protolambda/zssz"
)

var AttestationDataAndCustodyBitSSZ = zssz.GetSSZ((*AttestationDataAndCustodyBit)(nil))

type AttestationDataAndCustodyBit struct {
	// Attestation data
	Data AttestationData
	// Custody bit
	CustodyBit bool
}

type AttestationData struct {
	// Root of the signed beacon block
	BeaconBlockRoot Root

	// FFG vote
	SourceEpoch Epoch
	SourceRoot  Root
	TargetEpoch Epoch
	TargetRoot  Root

	// Crosslink vote
	Crosslink Crosslink
}

type PendingAttestation struct {
	// Attester aggregation bitfield
	AggregationBitfield bitfield.Bitfield
	// Attestation data
	Data AttestationData
	// Inclusion delay
	InclusionDelay Slot
	// Proposer index
	ProposerIndex ValidatorIndex
}

type AttestationsState struct {
	PreviousEpochAttestations []*PendingAttestation
	CurrentEpochAttestations  []*PendingAttestation
}

// Optimized compared to spec: takes pre-shuffled active indices as input, to not shuffle per-committee.
func computeCommittee(shuffled []ValidatorIndex, index uint64, count uint64) []ValidatorIndex {
	// Return the index'th shuffled committee out of the total committees data (shuffled active indices)
	startOffset := (uint64(len(shuffled)) * index) / count
	endOffset := (uint64(len(shuffled)) * (index + 1)) / count
	return shuffled[startOffset:endOffset]
}

func (state *BeaconState) GetCrosslinkCommittee(epoch Epoch, shard Shard) []ValidatorIndex {
	currentEpoch := state.Epoch()
	previousEpoch := state.PreviousEpoch()
	nextEpoch := currentEpoch + 1

	if !(previousEpoch <= epoch && epoch <= nextEpoch) {
		panic("could not retrieve crosslink committee for out of range slot")
	}

	seed := state.GenerateSeed(epoch)
	activeIndices := state.Validators.GetActiveValidatorIndices(epoch)
	// Active validators, shuffled in-place.
	// TODO: cache shuffling
	shuffling.UnshuffleList(activeIndices, seed)
	index := uint64((shard + SHARD_COUNT - state.GetEpochStartShard(epoch)) % SHARD_COUNT)
	count := state.Validators.GetEpochCommitteeCount(epoch)
	return computeCommittee(activeIndices, index, count)
}

func (state *BeaconState) GetAttesters(attestations []*PendingAttestation, filter func(att *AttestationData) bool) ValidatorSet {
	out := make(ValidatorSet, 0)
	for _, att := range attestations {
		// If the attestation is for the boundary:
		if filter(&att.Data) {
			participants, _ := state.GetAttestingIndicesUnsorted(&att.Data, &att.AggregationBitfield)
			out = append(out, participants...)
		}
	}
	out.Dedup()
	return out
}

func (state *BeaconState) GetAttestationSlot(attData *AttestationData) Slot {
	epoch := attData.TargetEpoch
	committeeCount := Slot(state.Validators.GetEpochCommitteeCount(epoch))
	offset := Slot((attData.Crosslink.Shard + SHARD_COUNT - state.GetEpochStartShard(epoch)) % SHARD_COUNT)
	return epoch.GetStartSlot() + (offset / (committeeCount / SLOTS_PER_EPOCH))
}
