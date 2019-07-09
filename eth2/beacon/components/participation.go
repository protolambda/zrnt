package components

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bitfield"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"sort"
)

// Return the sorted attesting indices at for the attestation_data and bitfield
func (state *BeaconState) GetAttestingIndicesUnsorted(attestationData *AttestationData, bitfield *bitfield.Bitfield) ([]ValidatorIndex, error) {
	// Find the committee in the list with the desired shard
	crosslinkCommittee := state.GetCrosslinkCommittee(attestationData.Target.Epoch, attestationData.Crosslink.Shard)

	if len(crosslinkCommittee) == 0 {
		return nil, fmt.Errorf("cannot find crosslink committee at target epoch %d for shard %d", attestationData.Target.Epoch, attestationData.Crosslink.Shard)
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
