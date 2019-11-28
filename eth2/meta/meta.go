package meta

import (
	"github.com/protolambda/zrnt/eth2/beacon/validator"
	. "github.com/protolambda/zrnt/eth2/core"
)

type Exits interface {
	InitiateValidatorExit(currentEpoch Epoch, index ValidatorIndex) error
}

type ExitEpoch interface {
	ExitEpoch(index ValidatorIndex) (Epoch, error)
}
type ActivationEpoch interface {
	ActivationEpoch(index ValidatorIndex) (Epoch, error)
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

type SlashableCheck interface {
	IsSlashable(i ValidatorIndex, epoch Epoch) (bool, error)
}

type CompactCommittees interface {
	Pubkeys
	EffectiveBalances
	SlashedIndices
	GetCompactCommitteesRoot(epoch Epoch) Root
}

type Staking interface {
	// Staked = Active effective balance
	GetTotalStake() (Gwei, error)
	GetAttestersStake(statuses []AttesterStatus, mask AttesterFlag) (Gwei, error)
}

type Slashing interface {
	SlashAndDelayWithdraw(index ValidatorIndex, withdrawalEpoch Epoch)
	GetIndicesToSlash(withdrawal Epoch) (out []ValidatorIndex, err error)
}

type SlashingHistory interface {
	ResetSlashings(epoch Epoch) error
}

type Slasher interface {
	SlashValidator(slashedIndex ValidatorIndex, whistleblowerIndex *ValidatorIndex) error
}

type Validators interface {
	Validator(index ValidatorIndex) (*validator.Validator, error)
}

type Versioning interface {
	CurrentSlot() (Slot, error)
	CurrentEpoch() (Epoch, error)
	PreviousEpoch() (Epoch, error)
}

type SigDomain interface {
	GetDomain(dom BLSDomainType, messageEpoch Epoch) (BLSDomain, error)
}

type Eth1Voting interface {
	ResetEth1Votes() error
}

type Deposits interface {
	DepIndex() (DepositIndex, error)
	DepCount() (DepositIndex, error)
	DepRoot() (Root, error)
}

type Onboarding interface {
	AddNewValidator(pubkey BLSPubkey, withdrawalCreds Root, balance Gwei) error
}

type Depositing interface {
	IncrementDepositIndex() error
}

type LatestHeader interface {
	// Signing root of latest_block_header
	GetLatestBlockRoot() (Root, error)
}

type LatestHeaderUpdate interface {
	UpdateLatestBlockStateRoot(stateRoot Root) error
}

type History interface {
	GetBlockRootAtSlot(slot Slot) (Root, error)
	GetBlockRoot(epoch Epoch) (Root, error)
}

type HistoryUpdate interface {
	SetRecentRoots(slot Slot, blockRoot Root, stateRoot Root) error
	UpdateStateRoot(root Root) error
	UpdateHistoricalRoots() error
}

type EpochSeed interface {
	// Retrieve the seed for beacon proposer indices.
	GetSeed(epoch Epoch, domainType BLSDomainType) (Root, error)
}

type Proposers interface {
	GetBeaconProposerIndex(slot Slot) (ValidatorIndex, error)
}

type ActivationExit interface {
	GetChurnLimit(epoch Epoch) (uint64, error)
	ExitQueueEnd(epoch Epoch) (Epoch, error)
}

type ActivationQeueue interface {
	ProcessActivationQueue(activationEpoch Epoch, currentEpoch Epoch) error
}

type ActiveValidatorCount interface {
	GetActiveValidatorCount(epoch Epoch) (uint64, error)
}

type ValidatorEpochData interface {
	WithdrawableEpoch(index ValidatorIndex) (Epoch, error)
}

type ActiveIndices interface {
	IsActive(index ValidatorIndex, epoch Epoch) (bool, error)
	GetActiveValidatorIndices(epoch Epoch) (RegistryIndices, error)
	ComputeActiveIndexRoot(epoch Epoch) (Root, error)
}

type CommitteeCount interface {
	GetCommitteeCountAtSlot(slot Slot) (uint64, error)
}

type BeaconCommittees interface {
	GetBeaconCommittee(slot Slot, index CommitteeIndex) ([]ValidatorIndex, error)
}

type Randao interface {
	PrepareRandao(epoch Epoch) error
}

type Randomness interface {
	GetRandomMix(epoch Epoch) (Root, error)
}
