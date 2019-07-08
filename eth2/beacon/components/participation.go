package components

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bitfield"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"sort"
)

func (state *BeaconState) GetWinningCrosslinkAndAttestingIndices(shard Shard, epoch Epoch) (*Crosslink, ValidatorSet) {
	pendingAttestations := state.PreviousEpochAttestations
	if epoch == state.Epoch() {
		pendingAttestations = state.CurrentEpochAttestations
	}

	latestCrosslinkRoot := ssz.HashTreeRoot(&state.CurrentCrosslinks[shard], CrosslinkSSZ)

	// keyed by raw crosslink object. Not too big, and simplifies reduction to unique crosslinks
	crosslinkAttesters := make(map[*Crosslink]ValidatorSet)
	for _, att := range pendingAttestations {
		if att.Data.Crosslink.Shard == shard {
			if att.Data.Crosslink.ParentRoot == latestCrosslinkRoot ||
				latestCrosslinkRoot == ssz.HashTreeRoot(&att.Data.Crosslink, CrosslinkSSZ) {
				participants, _ := state.GetAttestingIndices(&att.Data, &att.AggregationBitfield)
				crosslinkAttesters[&att.Data.Crosslink] = append(crosslinkAttesters[&att.Data.Crosslink], participants...)
			}
		}
	}
	// handle when no attestations for shard available
	if len(crosslinkAttesters) == 0 {
		return &Crosslink{}, nil
	}
	for k, v := range crosslinkAttesters {
		v.Dedup()
		crosslinkAttesters[k] = state.Validators.FilterUnslashed(v)
	}

	// Now determine the best crosslink, by total weight (votes, weighted by balance)
	var winningLink *Crosslink = nil
	winningWeight := Gwei(0)
	for crosslink, attesters := range crosslinkAttesters {
		// effectively "get_attesting_balance": attesters consists of only de-duplicated unslashed validators.
		weight := state.Validators.GetTotalEffectiveBalanceOf(attesters)
		if winningLink == nil || weight > winningWeight {
			winningLink = crosslink
		}
		if winningLink != nil && weight == winningWeight {
			// break tie lexicographically
			for i := 0; i < 32; i++ {
				if crosslink.DataRoot[i] > winningLink.DataRoot[i] {
					winningLink = crosslink
					break
				}
			}
		}
	}

	// now retrieve all the attesters of this winning root
	winners := crosslinkAttesters[winningLink]

	return winningLink, winners
}

// Return the sorted attesting indices at for the attestation_data and bitfield
func (state *BeaconState) GetAttestingIndicesUnsorted(attestationData *AttestationData, bitfield *bitfield.Bitfield) ([]ValidatorIndex, error) {
	// Find the committee in the list with the desired shard
	crosslinkCommittee := state.GetCrosslinkCommittee(attestationData.TargetEpoch, attestationData.Crosslink.Shard)

	if len(crosslinkCommittee) == 0 {
		return nil, fmt.Errorf("cannot find crosslink committee at target epoch %d for shard %d", attestationData.TargetEpoch, attestationData.Crosslink.Shard)
	}
	if !bitfield.VerifySize(uint64(len(crosslinkCommittee))) {
		return nil, errors.New("bitfield has wrong size for corresponding crosslink committee")
	}

	// Find the participating attesters in the committee
	participants := make([]ValidatorIndex, 0, len(crosslinkCommittee))
	for i, vIndex := range crosslinkCommittee {
		if bitfield.GetBit(uint64(i)) == 1 {
			participants = append(participants, vIndex)
		}
	}
	return participants, nil
}

// Return the sorted attesting indices at for the attestation_data and bitfield
func (state *BeaconState) GetAttestingIndices(attestationData *AttestationData, bitfield *bitfield.Bitfield) (ValidatorSet, error) {
	participants, err := state.GetAttestingIndicesUnsorted(attestationData, bitfield)
	if err != nil {
		return nil, err
	}
	out := ValidatorSet(participants)
	sort.Sort(out)
	return out, nil
}
