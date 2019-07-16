package components

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

// Randomness and committees
type CompactCommitteesState struct {
	CompactCommitteesRoots [EPOCHS_PER_HISTORICAL_VECTOR]Root
}

type CommitteePubkeys []BLSPubkey

func (_ *CommitteePubkeys) Limit() uint64 {
	return MAX_VALIDATORS_PER_COMMITTEE
}

type CompactValidator uint64

func (cv CompactValidator) Index() ValidatorIndex {
	return ValidatorIndex(cv >> 16)
}

func (cv CompactValidator) Slashed() bool {
	return (cv >> 15) & 1 == 1
}

func (cv CompactValidator) EffectiveBalance() Gwei {
	return Gwei(cv & ((1 << 15) - 1))
}

type CommitteeCompactValidators []uint64

func (_ *CommitteeCompactValidators) Limit() uint64 {
	return MAX_VALIDATORS_PER_COMMITTEE
}

type CompactCommittee struct {
	Pubkeys CommitteePubkeys
	CompactValidators CommitteeCompactValidators
}

type CompactCommittees [SHARD_COUNT]CompactCommittee

var CompactCommitteesSSZ = zssz.GetSSZ((*CompactCommittees)(nil))

func (state *BeaconState) GetCompactCommittees(epoch Epoch) *CompactCommittees {
	compacts := CompactCommittees{}
	for shard := Shard(0); shard < SHARD_COUNT; shard++ {
		committee := state.PrecomputedData.GetCrosslinkCommittee(epoch, shard)
		compact := &compacts[shard]
		compact.Pubkeys = make(CommitteePubkeys, 0, len(committee))
		compact.CompactValidators = make(CommitteeCompactValidators, 0, len(committee))
		for _, index := range committee {
			v := state.Validators[index]
			compact.Pubkeys = append(compact.Pubkeys, v.Pubkey)
			compactData := uint64(index) << 16
			if v.Slashed {
				compactData |= 1 << 15
			}
			compactData |= uint64(v.EffectiveBalance / EFFECTIVE_BALANCE_INCREMENT)
			compact.CompactValidators = append(compact.CompactValidators, compactData)
		}
	}
	return &compacts
}

func (state *BeaconState) GetCompactCommittesRoot(epoch Epoch) Root {
	return ssz.HashTreeRoot(state.GetCompactCommittees(epoch), CompactCommitteesSSZ)
}

func (state *BeaconState) UpdateCompactCommitteesRoot(epoch Epoch) {
	position := epoch % EPOCHS_PER_HISTORICAL_VECTOR
	state.CompactCommitteesRoots[position] = state.GetCompactCommittesRoot(epoch)
}
