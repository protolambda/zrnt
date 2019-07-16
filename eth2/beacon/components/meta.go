package components

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	. "github.com/protolambda/zrnt/eth2/core"
)

type ExitMeta interface {
	InitiateValidatorExit(index ValidatorIndex)
}

type BalanceMeta interface {
	IncreaseBalance(index ValidatorIndex, v Gwei)
	DecreaseBalance(index ValidatorIndex, v Gwei)
}

type CompactValidatorMeta interface {
	EffectiveBalance(index ValidatorIndex) Gwei
	Pubkey(index ValidatorIndex) BLSPubkey
	IsValidIndex(index ValidatorIndex) bool
}

type ValidatorMeta interface {
	Validator(index ValidatorIndex) *Validator
}

type VersioningMeta interface {
	Slot() Slot
	Epoch() Epoch
	GetDomain(dom BLSDomainType, messageEpoch Epoch) BLSDomain
}

type ProposingMeta interface {
	GetBeaconProposerIndex() ValidatorIndex
}

type ActivationExitMeta interface {
	GetChurnLimit(epoch Epoch) uint64
	ExitQueueEnd(epoch Epoch) Epoch
}

type FullRegistryMeta interface {
	ValidatorMeta
	GetActiveValidatorIndices() RegistryIndices
}

type ShardDeltaMeta interface {
	GetShardDelta(epoch Epoch) Shard
}

type ActiveIndexRootMeta interface {
	GetActiveIndexRoot(epoch Epoch) Root
	// TODO
	// indices := state.Validators.GetActiveValidatorIndices(epoch)
	// ssz.HashTreeRoot(indices, RegistryIndicesSSZ)
}

type CrosslinkMeta interface {
	GetCrosslinkCommittee(epoch Epoch, shard Shard) []ValidatorIndex
	GetStartShard(epoch Epoch) Shard
	GetCommitteeCount(epoch Epoch) uint64
}

type RandomnessMeta interface {
	GetRandomMix(epoch Epoch) Root
	// epoch + EPOCHS_PER_HISTORICAL_VECTOR - MIN_SEED_LOOKAHEAD) // Avoid underflow
}
