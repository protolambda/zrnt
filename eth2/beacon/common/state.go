package common

import (
	"context"

	"github.com/protolambda/ztyp/tree"
	"github.com/protolambda/ztyp/view"
)

type BatchRoots interface {
	GetRoot(slot Slot) (Root, error)
	SetRoot(slot Slot, r Root) error
	HashTreeRoot(fn tree.HashFn) Root
}
type Validator interface {
	Pubkey() (BLSPubkey, error)
	WithdrawalCredentials() (out Root, err error)
	SetWithdrawalCredentials(out Root) (err error)
	EffectiveBalance() (Gwei, error)
	SetEffectiveBalance(b Gwei) error
	Slashed() (bool, error)
	MakeSlashed() error
	ActivationEligibilityEpoch() (Epoch, error)
	SetActivationEligibilityEpoch(epoch Epoch) error
	ActivationEpoch() (Epoch, error)
	SetActivationEpoch(epoch Epoch) error
	ExitEpoch() (Epoch, error)
	SetExitEpoch(ep Epoch) error
	PrincipalBalance() (Gwei, error)
	SetPrincipalBalance(pb Gwei) error
	// Flatten the validator data into destination struct
	// For intensive validator registry work, it is more efficient to iterate the registry once,
	// unpack validators into a flat structure, and work with the flattened data.
	Flatten(dst *FlatValidator) error
}

type ValidatorRegistry interface {
	ValidatorCount() (uint64, error)
	Validator(index ValidatorIndex) (Validator, error)
	Iter() (next func() (val Validator, ok bool, err error))
	IsValidIndex(index ValidatorIndex) (valid bool, err error)
	HashTreeRoot(fn tree.HashFn) Root
}

type BalancesRegistry interface {
	GetBalance(index ValidatorIndex) (Gwei, error)
	SetBalance(index ValidatorIndex, bal Gwei) error
	AppendBalance(bal Gwei) error
	Iter() (next func() (bal Gwei, ok bool, err error))
	AllBalances() ([]Gwei, error)
	Length() (uint64, error)
}

type RandaoMixes interface {
	GetRandomMix(epoch Epoch) (Root, error)
	SetRandomMix(epoch Epoch, mix Root) error
}

// ForkSettings are values that changed throughout forks, without change in surrounding logic.
// To select the right settings based on configuration, ForkSettings(spec) is called on
// the fork-specific implementation of the BeaconState interface.
type ForkSettings struct {
	MinSlashingPenaltyQuotient uint64
	InactivityPenaltyQuotient  uint64
	CalcProposerShare          func(whistleblowerReward Gwei) Gwei
}

type BeaconState interface {
	view.View

	GenesisTime() (Timestamp, error)
	SetGenesisTime(t Timestamp) error
	GenesisValidatorsRoot() (Root, error)
	SetGenesisValidatorsRoot(r Root) error

	Slot() (Slot, error)
	SetSlot(slot Slot) error
	Fork() (Fork, error)
	SetFork(f Fork) error
	// Returns a copy of the latest header in the state
	LatestBlockHeader() (*BeaconBlockHeader, error)
	SetLatestBlockHeader(v *BeaconBlockHeader) error
	BlockRoots() (BatchRoots, error)
	StateRoots() (BatchRoots, error)
	RewardAdjustmentFactor() (Number, error)
	SetRewardAdjustmentFactor(v Number) error

	Validators() (ValidatorRegistry, error)

	Balances() (BalancesRegistry, error)
	SetBalances(balances []Gwei) error
	Reserves() (Number, error)
	SetReserves(v Number) error

	AddValidator(spec *Spec, pub BLSPubkey, withdrawalCreds Root, balance Gwei) error

	RandaoMixes() (RandaoMixes, error)
	SeedRandao(spec *Spec, seed Root) error

	JustificationBits() (JustificationBits, error)
	SetJustificationBits(bits JustificationBits) error

	PreviousJustifiedCheckpoint() (Checkpoint, error)
	SetPreviousJustifiedCheckpoint(c Checkpoint) error
	CurrentJustifiedCheckpoint() (Checkpoint, error)
	SetCurrentJustifiedCheckpoint(c Checkpoint) error
	FinalizedCheckpoint() (Checkpoint, error)
	SetFinalizedCheckpoint(c Checkpoint) error

	HashTreeRoot(fn tree.HashFn) Root
	CopyState() (BeaconState, error)

	ForkSettings(spec *Spec) *ForkSettings

	// ProcessEpoch applies an epoch-transition to the state.
	ProcessEpoch(ctx context.Context, spec *Spec, epc *EpochsContext) error
	// ProcessBlock applies a block to the state.
	// Excludes slot processing and signature validation. Just applies the block as-is. Error if mismatching slot.
	ProcessBlock(ctx context.Context, spec *Spec, epc *EpochsContext, benv *BeaconBlockEnvelope) error
}

type UpgradeableBeaconState interface {
	BeaconState

	// Called whenever the state may need to upgrade to a next fork, changes the BeaconState interface contents if so.
	UpgradeMaybe(ctx context.Context, spec *Spec, epc *EpochsContext) error
}
