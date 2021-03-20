package common

import "github.com/protolambda/ztyp/tree"

type BatchRoots interface {
	GetRoot(slot Slot) (Root, error)
	SetRoot(slot Slot, r Root) error
	HashTreeRoot(fn tree.HashFn) Root
}

type HistoricalRoots interface {
	Append(root Root) error
}

type Eth1DataVotes interface {
	Reset() error
	Length() (uint64, error)
	Count(dat Eth1Data) (uint64, error)
	Append(dat Eth1Data) error
}

type Validator interface {
	Pubkey() (BLSPubkey, error)
	WithdrawalCredentials() (out Root, err error)
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
	WithdrawableEpoch() (Epoch, error)
	SetWithdrawableEpoch(epoch Epoch) error
}

type ValidatorRegistry interface {
	ValidatorCount() (uint64, error)
	Validator(index ValidatorIndex) (Validator, error)
	Iter() (next func() (val Validator, ok bool, err error))
	IsValidIndex(index ValidatorIndex) (valid bool, err error)
}

type Balances interface {
	GetBalance(index ValidatorIndex) (Gwei, error)
	SetBalance(index ValidatorIndex, bal Gwei) error
	Iter() (next func() (bal Gwei, ok bool, err error))
}

type RandaoMixes interface {
	GetRandomMix(epoch Epoch) (Root, error)
	SetRandomMix(epoch Epoch, mix Root) error
}

type Slashings interface {
	GetSlashingsValue(epoch Epoch) (Gwei, error)
	ResetSlashings(epoch Epoch) error
	AddSlashing(epoch Epoch, add Gwei) error
	Total() (sum Gwei, err error)
}

type BeaconState interface {
	GenesisTime() (Timestamp, error)
	SetGenesisTime(t Timestamp) error
	GenesisValidatorsRoot() (Root, error)
	SetGenesisValidatorsRoot(r Root) error

	Slot() (Slot, error)
	SetSlot(slot Slot) error
	Fork() (Fork, error)
	SetFork(f Fork) error
	LatestBlockHeader() (*BeaconBlockHeader, error)
	SetLatestBlockHeader(v *BeaconBlockHeader) error
	BlockRoots() (BatchRoots, error)
	StateRoots() (BatchRoots, error)
	HistoricalRoots() (HistoricalRoots, error)
	Eth1Data() (Eth1Data, error)
	SetEth1Data(v Eth1Data) error
	Eth1DataVotes() (Eth1DataVotes, error)
	DepositIndex() (DepositIndex, error)
	IncrementDepositIndex() error

	Validators() (ValidatorRegistry, error)
	Balances() (Balances, error)

	AddValidator(spec *Spec, pub BLSPubkey, withdrawalCreds Root, balance Gwei) error

	RandaoMixes() (RandaoMixes, error)
	SeedRandao(spec *Spec, seed Root) error

	Slashings() (Slashings, error)

	JustificationBits() (JustificationBits, error)
	SetJustificationBits(bits JustificationBits) error

	PreviousJustifiedCheckpoint() (Checkpoint, error)
	SetPreviousJustifiedCheckpoint(c Checkpoint) error
	CurrentJustifiedCheckpoint() (Checkpoint, error)
	SetCurrentJustifiedCheckpoint(c Checkpoint) error
	FinalizedCheckpoint() (Checkpoint, error)
	SetFinalizedCheckpoint(c Checkpoint) error

	HashTreeRoot(fn tree.HashFn) Root
}
