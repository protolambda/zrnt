package meta

import (
	"github.com/protolambda/zrnt/eth2/beacon/validator"
	. "github.com/protolambda/zrnt/eth2/core"
)

type Exits interface {
	InitiateValidatorExit(currentEpoch Epoch, index ValidatorIndex)
}

type Balance interface {
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

type RegistrySize interface {
	IsValidIndex(index ValidatorIndex) bool
	ValidatorCount() uint64
}

type Pubkeys interface {
	Pubkey(index ValidatorIndex) BLSPubkey
	ValidatorIndex(pubkey BLSPubkey) (index ValidatorIndex, exists bool)
}

type EffectiveBalances interface {
	EffectiveBalance(index ValidatorIndex) Gwei
	SumEffectiveBalanceOf(indices []ValidatorIndex) (sum Gwei)
}

type EffectiveBalancesUpdate interface {
	UpdateEffectiveBalances()
}

type Finality interface {
	Finalized() Checkpoint
	CurrentJustified() Checkpoint
	PreviousJustified() Checkpoint
}

type Justification interface {
	Justify(checkpoint Checkpoint)
}

type EpochAttestations interface {
	RotateEpochAttestations()
}

type AttesterStatuses interface {
	GetAttesterStatus(index ValidatorIndex) AttesterStatus
}

type SlashedIndices interface {
	IsSlashed(i ValidatorIndex) bool
	FilterUnslashed(indices []ValidatorIndex) []ValidatorIndex
}

type CompactCommittees interface {
	Pubkeys
	EffectiveBalances
	SlashedIndices
	GetCompactCommitteesRoot(epoch Epoch) Root
}

type CompactCommitteesUpdate interface {
	UpdateCompactCommitteesRoot(epoch Epoch)
}

type Staking interface {
	// Staked = Active effective balance
	GetTotalStakedBalance(epoch Epoch) Gwei
}

type TargetStaking interface {
	GetTargetTotalStakedBalance(epoch Epoch) Gwei
}

type Slashing interface {
	GetIndicesToSlash(withdrawal Epoch) (out []ValidatorIndex)
}

type SlashingHistory interface {
	ResetSlashings(epoch Epoch)
}

type Slasher interface {
	SlashValidator(slashedIndex ValidatorIndex, whistleblowerIndex *ValidatorIndex)
}

type Validators interface {
	Validator(index ValidatorIndex) *validator.Validator
}

type Versioning interface {
	CurrentSlot() Slot
	CurrentEpoch() Epoch
	PreviousEpoch() Epoch
	GetDomain(dom BLSDomainType, messageEpoch Epoch) BLSDomain
}

type Eth1Voting interface {
	ResetEth1Votes()
}

type Deposits interface {
	DepIndex() DepositIndex
	DepCount() DepositIndex
	DepRoot() Root
}

type Onboarding interface {
	AddNewValidator(pubkey BLSPubkey, withdrawalCreds Root, balance Gwei)
}

type Depositing interface {
	IncrementDepositIndex()
}

type LatestHeader interface {
	// Signing root of latest_block_header
	GetLatestBlockRoot() Root
}

type LatestHeaderUpdate interface {
	UpdateLatestBlockRoot(stateRoot Root) Root
}

type History interface {
	GetBlockRootAtSlot(slot Slot) Root
	GetBlockRoot(epoch Epoch) Root
}

type HistoryUpdate interface {
	SetRecentRoots(slot Slot, blockRoot Root, stateRoot Root)
	UpdateStateRoot(root Root)
	UpdateHistoricalRoots()
}

type EpochSeed interface {
	// Retrieve the seed for beacon proposer indices.
	GetSeed(epoch Epoch) Root
}

type Proposers interface {
	GetBeaconProposerIndex(slot Slot) ValidatorIndex
}

type ActivationExit interface {
	GetChurnLimit(epoch Epoch) uint64
	ExitQueueEnd(epoch Epoch) Epoch
}

type ActivationQeueue interface {
	ProcessActivationQueue(activationEpoch Epoch, currentEpoch Epoch)
}

type ActiveValidatorCount interface {
	GetActiveValidatorCount(epoch Epoch) uint64
}

type ActiveIndices interface {
	GetActiveValidatorIndices(epoch Epoch) RegistryIndices
	ComputeActiveIndexRoot(epoch Epoch) Root
}

type ActiveIndexRoots interface {
	GetActiveIndexRoot(epoch Epoch) Root
}

type ActiveIndexRootsUpdate interface {
	UpdateActiveIndexRoot(epoch Epoch)
}

type CommitteeCount interface {
	// Amount of committees per epoch. Minimum is SLOTS_PER_EPOCH
	GetCommitteeCount(epoch Epoch) uint64
}

type CrosslinkTiming interface {
	GetStartShard(epoch Epoch) Shard
}

type ShardRotation interface {
	GetShardDelta(epoch Epoch) Shard
	RotateStartShard()
}

type CrosslinkCommittees interface {
	GetCrosslinkCommittee(epoch Epoch, shard Shard) []ValidatorIndex
}

type Crosslinks interface {
	GetCurrentCrosslinkRoots() *[SHARD_COUNT]Root
	GetPreviousCrosslinkRoots() *[SHARD_COUNT]Root
	GetCurrentCrosslink(shard Shard) *Crosslink
	GetPreviousCrosslink(shard Shard) *Crosslink
}

type WinningCrosslinks interface {
	GetWinningCrosslinkAndAttesters(epoch Epoch, shard Shard) (*Crosslink, ValidatorSet)
}

type Randao interface {
	PrepareRandao(epoch Epoch)
}

type Randomness interface {
	GetRandomMix(epoch Epoch) Root
}
