package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"sort"
)

type CrosslinkingFeature struct {
	State *AttestationsState
	Meta interface {
		meta.VersioningMeta
		meta.CrosslinkMeta
		meta.EffectiveBalanceMeta
		meta.CrosslinkCommitteeMeta
		meta.SlashedMeta
	}
}

type CrosslinkingStatus struct {
	Previous *CrosslinkingEpoch
	Current  *CrosslinkingEpoch
}

func (f *CrosslinkingFeature) LoadCrosslinkingStatus() *CrosslinkingStatus {
	return &CrosslinkingStatus{
		Previous: f.LoadCrosslinkEpoch(f.Meta.PreviousEpoch()),
		Current:  f.LoadCrosslinkEpoch(f.Meta.CurrentEpoch()),
	}
}

type LinkWinner struct {
	Crosslink *Crosslink   // nil when there are no crosslinks for the shard.
	Attesters ValidatorSet // nil-slice when there are no attestations for the shard.
}

type CrosslinkingEpoch struct {
	WinningLinks [SHARD_COUNT]LinkWinner
}

type weightedLink struct {
	weight    Gwei
	link      *Crosslink
	attesters []ValidatorIndex
}

func (f *CrosslinkingFeature) LoadCrosslinkEpoch(epoch Epoch) *CrosslinkingEpoch {
	var crosslinkRoots *[SHARD_COUNT]Root
	var attestations []*PendingAttestation

	if epoch == f.Meta.CurrentEpoch() {
		crosslinkRoots = f.Meta.GetPreviousCrosslinkRoots()
		attestations = f.State.PreviousEpochAttestations
	} else {
		crosslinkRoots = f.Meta.GetCurrentCrosslinkRoots()
		attestations = f.State.CurrentEpochAttestations
	}

	// Keyed by raw crosslink object. Not too big, and simplifies reduction to unique crosslinks
	// For shards with no attestations available, the value will be a nil slice.
	crosslinkAttesters := make(map[*Crosslink]CommitteeBits)
	for _, att := range attestations {
		shard := att.Data.Crosslink.Shard
		if att.Data.Crosslink.ParentRoot == crosslinkRoots[shard] ||
			ssz.HashTreeRoot(&att.Data.Crosslink, crosslinkSSZ) == crosslinkRoots[shard] {

			bits, ok := crosslinkAttesters[&att.Data.Crosslink]
			if !ok {
				// initialize new bitlist. We can ignore the leading bit, it will be ORed anyway.
				bits = make(CommitteeBits, len(att.AggregationBits))
				crosslinkAttesters[&att.Data.Crosslink] = bits
			}

			// Mark attesters
			bits.Or(att.AggregationBits)
		}
	}

	winningCrosslinks := [SHARD_COUNT]weightedLink{}
	participants := make([]ValidatorIndex, 0, MAX_VALIDATORS_PER_COMMITTEE)
	for k, v := range crosslinkAttesters {
		shard := k.Shard
		committee := f.Meta.GetCrosslinkCommittee(epoch, shard)
		participants = participants[:0]                      // reset old slice (re-used in for loop)
		participants = append(participants, committee...)    // add committee indices
		participants = v.FilterParticipants(participants)    // only keep the participants
		participants = f.Meta.FilterUnslashed(participants)  // and only those who are not slashed
		weight := f.Meta.SumEffectiveBalanceOf(participants) // and get their weight

		currentWinner := &winningCrosslinks[shard]
		isNewWinner := currentWinner.link == nil
		isNewWinner = isNewWinner || (weight > currentWinner.weight)
		if !isNewWinner && weight == currentWinner.weight {
			// break tie lexicographically
			for i := 0; i < 32; i++ {
				if k.DataRoot[i] > currentWinner.link.DataRoot[i] {
					isNewWinner = true
					break
				}
			}
		}
		if isNewWinner {
			// overwrite winning link
			currentWinner.weight = weight
			currentWinner.link = k
			if currentWinner.attesters == nil {
				currentWinner.attesters = make([]ValidatorIndex, 0, len(participants)<<2) // bit of extra capacity
			}
			// re-use previously allocated indices slice (append will re-allocate if more participants than previously)
			currentWinner.attesters = currentWinner.attesters[:0]
			currentWinner.attesters = append(currentWinner.attesters, participants...)
		}
	}

	crep := new(CrosslinkingEpoch)
	for shard, winner := range winningCrosslinks {
		out := &crep.WinningLinks[shard]
		out.Crosslink = winner.link
		out.Attesters = winner.attesters
		if out.Attesters != nil {
			sort.Sort(out.Attesters) // validator sets must be sorted
		}
	}
	return crep
}

