package meta

import (
	"github.com/protolambda/zrnt/eth2/beacon/validator"
	. "github.com/protolambda/zrnt/eth2/core"
)

type ExitMeta interface {
	InitiateValidatorExit(currentEpoch Epoch, index ValidatorIndex)
}

type BalanceMeta interface {
	GetBalance(index ValidatorIndex) Gwei
	IncreaseBalance(index ValidatorIndex, v Gwei)
	DecreaseBalance(index ValidatorIndex, v Gwei)
}

type BalanceDeltas interface {
	ApplyDeltas(deltas *Deltas)
}

type AttestationDeltas interface {
	AttestationDeltas() *Deltas
}

type CrosslinkDeltas interface {
	CrosslinkDeltas() *Deltas
}

type RegistrySizeMeta interface {
	IsValidIndex(index ValidatorIndex) bool
	ValidatorCount() uint64
}

type PubkeyMeta interface {
	Pubkey(index ValidatorIndex) BLSPubkey
	ValidatorIndex(pubkey BLSPubkey) (index ValidatorIndex, exists bool)
}

type EffectiveBalanceMeta interface {
	EffectiveBalance(index ValidatorIndex) Gwei
	SumEffectiveBalanceOf(indices []ValidatorIndex) (sum Gwei)
}

type FinalityMeta interface {
	Finalized() Checkpoint
	CurrentJustified() Checkpoint
	PreviousJustified() Checkpoint
}

type JustificationMeta interface {
	Justify(checkpoint Checkpoint)
}

type AttesterStatusMeta interface {
	GetAttesterStatus(index ValidatorIndex) AttesterStatus
}

type SlashedMeta interface {
	IsSlashed(i ValidatorIndex) bool
	FilterUnslashed(indices []ValidatorIndex) []ValidatorIndex
}

type CompactValidatorMeta interface {
	PubkeyMeta
	EffectiveBalanceMeta
	SlashedMeta
	GetCompactCommitteesRoot(epoch Epoch) Root
}

type StakingMeta interface {
	// Staked = Active effective balance
	GetTotalStakedBalance(epoch Epoch) Gwei
}

type TargetStakingMeta interface {
	GetTargetTotalStakedBalance(epoch Epoch) Gwei
}

type SlashingMeta interface {
	GetIndicesToSlash(withdrawal Epoch) (out []ValidatorIndex)
}

type SlasherMeta interface {
	SlashValidator(slashedIndex ValidatorIndex, whistleblowerIndex *ValidatorIndex)
}

type ValidatorMeta interface {
	Validator(index ValidatorIndex) *validator.Validator
}

type VersioningMeta interface {
	CurrentSlot() Slot
	CurrentEpoch() Epoch
	PreviousEpoch() Epoch
	GetDomain(dom BLSDomainType, messageEpoch Epoch) BLSDomain
}

type Eth1Meta interface {
	DepIndex() DepositIndex
	DepCount() DepositIndex
	DepRoot() Root
}

type OnboardMeta interface {
	AddNewValidator(pubkey BLSPubkey, withdrawalCreds Root, balance Gwei)
}

type DepositMeta interface {
	IncrementDepositIndex()
}

type HeaderMeta interface {
	// Signing root of latest_block_header
	GetLatestBlockRoot() Root
}

type UpdateHeaderMeta interface {
	UpdateLatestBlockRoot(stateRoot Root) Root
}

type HistoryMeta interface {
	GetBlockRootAtSlot(slot Slot) Root
	GetBlockRoot(epoch Epoch) Root
}

type HistoryUpdateMeta interface {
	SetRecentRoots(slot Slot, blockRoot Root, stateRoot Root)
	UpdateStateRoot(root Root)
}

type SeedMeta interface {
	// Retrieve the seed for beacon proposer indices.
	GetSeed(epoch Epoch) Root
}

type ProposingMeta interface {
	GetBeaconProposerIndex(slot Slot) ValidatorIndex
}

type ActivationExitMeta interface {
	GetChurnLimit(epoch Epoch) uint64
	ExitQueueEnd(epoch Epoch) Epoch
}

type ActivationMeta interface {
	ProcessActivationQueue(activationEpoch Epoch, currentEpoch Epoch)
}

type ActiveValidatorCountMeta interface {
	GetActiveValidatorCount(epoch Epoch) uint64
}

type ActiveIndicesMeta interface {
	GetActiveValidatorIndices(epoch Epoch) RegistryIndices
	ComputeActiveIndexRoot(epoch Epoch) Root
}

type ActiveIndexRootMeta interface {
	GetActiveIndexRoot(epoch Epoch) Root
}

type CommitteeCountMeta interface {
	// Amount of committees per epoch. Minimum is SLOTS_PER_EPOCH
	GetCommitteeCount(epoch Epoch) uint64
}

type CrosslinkTimingMeta interface {
	GetStartShard(epoch Epoch) Shard
}

type ShardRotMeta interface {
	GetShardDelta(epoch Epoch) Shard
}

type CrosslinkCommitteeMeta interface {
	GetCrosslinkCommittee(epoch Epoch, shard Shard) []ValidatorIndex
}

type CrosslinkMeta interface {
	GetCurrentCrosslinkRoots() *[SHARD_COUNT]Root
	GetPreviousCrosslinkRoots() *[SHARD_COUNT]Root
	GetCurrentCrosslink(shard Shard) *Crosslink
	GetPreviousCrosslink(shard Shard) *Crosslink
}

type WinningCrosslinkMeta interface {
	GetWinningCrosslinkAndAttesters(epoch Epoch, shard Shard) (*Crosslink, ValidatorSet)
}

type RandomnessMeta interface {
	GetRandomMix(epoch Epoch) Root
	// epoch + EPOCHS_PER_HISTORICAL_VECTOR - MIN_SEED_LOOKAHEAD) // TODO Avoid underflow
}

// TODO: split up?
type FinalUpdates interface {
	ResetEth1Votes()
	UpdateEffectiveBalances()
	RotateStartShard()
	UpdateActiveIndexRoot(epoch Epoch)
	UpdateCompactCommitteesRoot(epoch Epoch)
	ResetSlashings(epoch Epoch)
	PrepareRandao(epoch Epoch)
	UpdateHistoricalRoots()
	RotateEpochAttestations()
}
