package meta

import (
	"github.com/protolambda/zrnt/eth2/beacon/validator"
	. "github.com/protolambda/zrnt/eth2/core"
)

type Exits interface {
	InitiateValidatorExit(currentEpoch Epoch, index ValidatorIndex) error
}

type Balance interface {
	GetBalance(index ValidatorIndex) (Gwei, error)
	IncreaseBalance(index ValidatorIndex, v Gwei) error
	DecreaseBalance(index ValidatorIndex, v Gwei) error
}

type BalanceDeltas interface {
	ApplyDeltas(deltas *Deltas) error
}

type AttestationDeltas interface {
	AttestationDeltas() (*Deltas, error)
}

type RegistrySize interface {
	IsValidIndex(index ValidatorIndex) (bool, error)
	ValidatorCount() (uint64, error)
}

type Pubkeys interface {
	Pubkey(index ValidatorIndex) (BLSPubkey, error)
	ValidatorIndex(pubkey BLSPubkey) (index ValidatorIndex, exists bool, err error)
}

type EffectiveBalances interface {
	EffectiveBalance(index ValidatorIndex) (Gwei, error)
	SumEffectiveBalanceOf(indices []ValidatorIndex) (sum Gwei, err error)
}

type EffectiveBalancesUpdate interface {
	UpdateEffectiveBalances() error
}

type Finality interface {
	Finalized() (Checkpoint, error)
	CurrentJustified() (Checkpoint, error)
	PreviousJustified() (Checkpoint, error)
}

type Justification interface {
	Justify(checkpoint Checkpoint) error
}

type EpochAttestations interface {
	RotateEpochAttestations() error
}

type AttesterStatuses interface {
	GetAttesterStatuses() ([]AttesterStatus, error)
}

type SlashedIndices interface {
	IsSlashed(i ValidatorIndex) (bool, error)
	FilterUnslashed(indices []ValidatorIndex) ([]ValidatorIndex, error)
}

type CompactCommittees interface {
	Pubkeys
	EffectiveBalances
	SlashedIndices
	GetCompactCommitteesRoot(epoch Epoch) Root
}

type Staking interface {
	// Staked = Active effective balance
	GetTotalStake() Gwei
	GetAttestersStake(statuses []AttesterStatus, mask AttesterFlag) Gwei
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
	GetSeed(epoch Epoch, domainType BLSDomainType) Root
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

type ValidatorEpochData interface {
	WithdrawableEpoch(index ValidatorIndex) Epoch
}

type ActiveIndices interface {
	IsActive(index ValidatorIndex, epoch Epoch) bool
	GetActiveValidatorIndices(epoch Epoch) RegistryIndices
	ComputeActiveIndexRoot(epoch Epoch) Root
}

type CommitteeCount interface {
	GetCommitteeCountAtSlot(slot Slot) uint64
}

type BeaconCommittees interface {
	GetBeaconCommittee(slot Slot, index CommitteeIndex) []ValidatorIndex
}

type Randao interface {
	PrepareRandao(epoch Epoch)
}

type Randomness interface {
	GetRandomMix(epoch Epoch) Root
}
