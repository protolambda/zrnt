package components

import (
	"encoding/binary"
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type ShufflingData interface {
	GetCrosslinkCommittee(epoch Epoch, shard Shard) []ValidatorIndex
}

// Randomness and committees
type ShufflingState struct {
	LatestActiveIndexRoots [EPOCHS_PER_HISTORICAL_VECTOR]Root
	CompactCommitteesRoots [EPOCHS_PER_HISTORICAL_VECTOR]Root
}

// Epoch is expected to be between (current_epoch - EPOCHS_PER_HISTORICAL_VECTOR + ACTIVATION_EXIT_DELAY, current_epoch + ACTIVATION_EXIT_DELAY].
func (state *ShufflingState) GetActiveIndexRoot(epoch Epoch) Root {
	return state.LatestActiveIndexRoots[epoch%EPOCHS_PER_HISTORICAL_VECTOR]
}

func (state *BeaconState) UpdateActiveIndexRoot(epoch Epoch) {
	position := epoch % EPOCHS_PER_HISTORICAL_VECTOR
	indices := state.Validators.GetActiveValidatorIndices(epoch)
	state.LatestActiveIndexRoots[position] = ssz.HashTreeRoot(indices, ValidatorIndexListSSZ)
}

type CommitteePubkeys []BLSPubkey

func (_ *CommitteePubkeys) Limit() uint32 {
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

func (_ *CommitteeCompactValidators) Limit() uint32 {
	return MAX_VALIDATORS_PER_COMMITTEE
}

type CompactCommittee struct {
	Pubkeys CommitteePubkeys
	CompactValidators CommitteeCompactValidators
}

type CompactCommittees [SHARD_COUNT]CompactCommittee

var CompactCommitteesSSZ = zssz.GetSSZ((*CompactCommittees)(nil))

func (state *BeaconState) GetCompactCommittesRoot(epoch Epoch) Root {
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
	return ssz.HashTreeRoot(&compacts, CompactCommitteesSSZ)
}

func (state *BeaconState) UpdateCompactCommitteesRoot(epoch Epoch) {
	position := epoch % EPOCHS_PER_HISTORICAL_VECTOR
	state.CompactCommitteesRoots[position] = state.GetCompactCommittesRoot(epoch)
}

// Generate a seed for the given epoch
func (state *BeaconState) GenerateSeed(epoch Epoch) Root {
	buf := make([]byte, 32*3)
	mix := state.GetRandomMix(epoch + EPOCHS_PER_HISTORICAL_VECTOR - MIN_SEED_LOOKAHEAD)
	copy(buf[0:32], mix[:])
	// get_active_index_root in spec, but only used once, and the assertion is unnecessary, since epoch input is always trusted
	activeIndexRoot := state.GetActiveIndexRoot(epoch)
	copy(buf[32:64], activeIndexRoot[:])
	binary.LittleEndian.PutUint64(buf[64:], uint64(epoch))
	return Hash(buf)
}
