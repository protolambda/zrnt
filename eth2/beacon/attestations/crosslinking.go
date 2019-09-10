package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"sort"
)

type CrosslinkingFeature struct {
	State *AttestationsState
	Meta  interface {
		meta.Versioning
		meta.Crosslinks
		meta.EffectiveBalances
		meta.CrosslinkCommittees
		meta.SlashedIndices
	}
}

type LinkWinner struct {
	Crosslink *Crosslink   // nil when there are no crosslinks for the shard.
	Attesters ValidatorSet // nil-slice when there are no attestations for the shard.
}

type CrosslinkingEpoch struct {
	Epoch        Epoch
	WinningLinks [SHARD_COUNT]LinkWinner
}

func (ce *CrosslinkingEpoch) GetWinningCrosslinkAndAttesters(shard Shard) (*Crosslink, ValidatorSet) {
	winner := ce.WinningLinks[shard]
	return winner.Crosslink, winner.Attesters
}

type weightedLink struct {
	attestationIndex uint64
	weight           Gwei
	link             *Crosslink
	attesters        []ValidatorIndex
}

type orderedCrosslinkAttesters struct {
	attestationIndex uint64
	crosslink        *Crosslink
	committee        CommitteeBits
}

func (f *CrosslinkingFeature) LoadEpochCrosslinkWinners(epoch Epoch) meta.EpochCrosslinkWinners {
	var attestations []*PendingAttestation
	if epoch == f.Meta.CurrentEpoch() {
		attestations = f.State.CurrentEpochAttestations
	} else {
		attestations = f.State.PreviousEpochAttestations
	}

	crosslinkRoots := f.Meta.GetCurrentCrosslinkRoots()

	// Keyed by raw crosslink object. Not too big, and simplifies reduction to unique crosslinks
	// For shards with no attestations available, the value will be a nil slice.
	crosslinkAttesters := make(map[Root]orderedCrosslinkAttesters)
	for i := range attestations {
		att := attestations[i]
		shard := att.Data.Crosslink.Shard
		cr := ssz.HashTreeRoot(&att.Data.Crosslink, crosslinkSSZ)
		if att.Data.Crosslink.ParentRoot == crosslinkRoots[shard] || cr == crosslinkRoots[shard] {

			oca, ok := crosslinkAttesters[cr]
			if !ok {
				// initialize new bitlist. We can ignore the leading bit, it will be ORed anyway.
				oca.committee = make(CommitteeBits, len(att.AggregationBits))
				oca.crosslink = &att.Data.Crosslink
				oca.attestationIndex = uint64(i)
			}

			// Mark attesters
			oca.committee.Or(att.AggregationBits)
			crosslinkAttesters[cr] = oca
		}
	}

	winningCrosslinks := [SHARD_COUNT]weightedLink{}
	participants := make([]ValidatorIndex, 0, MAX_VALIDATORS_PER_COMMITTEE)
	for _, v := range crosslinkAttesters {
		shard := v.crosslink.Shard
		committee := f.Meta.GetCrosslinkCommittee(epoch, shard)
		participants = participants[:0]                             // reset old slice (re-used in for loop)
		participants = append(participants, committee...)           // add committee indices
		participants = v.committee.FilterParticipants(participants) // only keep the participants
		participants = f.Meta.FilterUnslashed(participants)         // and only those who are not slashed
		weight := f.Meta.SumEffectiveBalanceOf(participants)        // and get their weight

		currentWinner := &winningCrosslinks[shard]
		isNewWinner := currentWinner.link == nil
		isNewWinner = isNewWinner || (weight > currentWinner.weight)
		if !isNewWinner && weight == currentWinner.weight {
			// if no lexicographical tie can be made, it is the attestation order that determines the winner
			if v.crosslink.DataRoot == currentWinner.link.DataRoot {
				if v.attestationIndex < currentWinner.attestationIndex {
					isNewWinner = true
					break
				}
			} else {
				// break tie lexicographically
				for i := 0; i < 32; i++ {
					if v.crosslink.DataRoot[i] > currentWinner.link.DataRoot[i] {
						isNewWinner = true
						break
					}
				}
			}
		}
		if isNewWinner {
			// overwrite winning link
			currentWinner.weight = weight
			currentWinner.attestationIndex = v.attestationIndex
			currentWinner.link = v.crosslink
			if currentWinner.attesters == nil {
				currentWinner.attesters = make([]ValidatorIndex, 0, len(participants)<<2) // bit of extra capacity
			}
			// re-use previously allocated indices slice (append will re-allocate if more participants than previously)
			currentWinner.attesters = currentWinner.attesters[:0]
			currentWinner.attesters = append(currentWinner.attesters, participants...)
		}
	}

	crep := &CrosslinkingEpoch{Epoch: epoch}
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
