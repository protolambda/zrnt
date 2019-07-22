package compact

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type CommitteePubkeys []BLSPubkey

func (_ *CommitteePubkeys) Limit() uint64 {
	return MAX_VALIDATORS_PER_COMMITTEE
}

type CommitteeCompactValidators []CompactValidator

func (_ *CommitteeCompactValidators) Limit() uint64 {
	return MAX_VALIDATORS_PER_COMMITTEE
}

type CompactCommittee struct {
	Pubkeys           CommitteePubkeys
	CompactValidators CommitteeCompactValidators
}

type CompactCommittees [SHARD_COUNT]CompactCommittee

var CompactCommitteesSSZ = zssz.GetSSZ((*CompactCommittees)(nil))

// Randomness and committees
type CompactCommitteesState struct {
	CompactCommitteesRoots [EPOCHS_PER_HISTORICAL_VECTOR]Root
}

func (state *CompactCommitteesState) GetCompactCommitteesRoot(epoch Epoch) Root {
	return state.CompactCommitteesRoots[epoch % EPOCHS_PER_HISTORICAL_VECTOR]
}

type CompactCommitteesFeature struct {
	State *CompactCommitteesState
	Meta interface {
		meta.CrosslinkCommitteeMeta
		meta.ValidatorMeta
	}
}

func (f *CompactCommitteesFeature) GetCompactCommittees(epoch Epoch) *CompactCommittees {
	compacts := CompactCommittees{}
	for shard := Shard(0); shard < SHARD_COUNT; shard++ {
		committee := f.Meta.GetCrosslinkCommittee(epoch, shard)
		compact := &compacts[shard]
		compact.Pubkeys = make(CommitteePubkeys, 0, len(committee))
		compact.CompactValidators = make(CommitteeCompactValidators, 0, len(committee))
		for _, index := range committee {
			v := f.Meta.Validator(index)
			compact.Pubkeys = append(compact.Pubkeys, v.Pubkey)
			compactValidator := MakeCompactValidator(index, v.Slashed, v.EffectiveBalance)
			compact.CompactValidators = append(compact.CompactValidators, compactValidator)
		}
	}
	return &compacts
}

func (cf *CompactCommitteesFeature) ComputeCompactCommitteesRoot(epoch Epoch) Root {
	return ssz.HashTreeRoot(cf.GetCompactCommittees(epoch), CompactCommitteesSSZ)
}

func (cf *CompactCommitteesFeature) UpdateCompactCommitteesRoot(epoch Epoch) {
	position := epoch % EPOCHS_PER_HISTORICAL_VECTOR
	cf.State.CompactCommitteesRoots[position] = cf.ComputeCompactCommitteesRoot(epoch)
}
