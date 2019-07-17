package compact

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/meta"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type CompactCommitteesReq interface {
	CrosslinkCommitteeMeta
	ValidatorMeta
}

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

func (state *CompactCommitteesState) GetCompactCommittees(meta CompactCommitteesReq, epoch Epoch) *CompactCommittees {
	compacts := CompactCommittees{}
	for shard := Shard(0); shard < SHARD_COUNT; shard++ {
		committee := meta.GetCrosslinkCommittee(epoch, shard)
		compact := &compacts[shard]
		compact.Pubkeys = make(CommitteePubkeys, 0, len(committee))
		compact.CompactValidators = make(CommitteeCompactValidators, 0, len(committee))
		for _, index := range committee {
			v := meta.Validator(index)
			compact.Pubkeys = append(compact.Pubkeys, v.Pubkey)
			compactValidator := MakeCompactValidator(index, v.Slashed, v.EffectiveBalance)
			compact.CompactValidators = append(compact.CompactValidators, compactValidator)
		}
	}
	return &compacts
}

func (state *CompactCommitteesState) GetCompactCommittesRoot(meta CompactCommitteesReq, epoch Epoch) Root {
	return ssz.HashTreeRoot(state.GetCompactCommittees(meta, epoch), CompactCommitteesSSZ)
}

func (state *CompactCommitteesState) UpdateCompactCommitteesRoot(meta CompactCommitteesReq, epoch Epoch) {
	position := epoch % EPOCHS_PER_HISTORICAL_VECTOR
	state.CompactCommitteesRoots[position] = state.GetCompactCommittesRoot(meta, epoch)
}
