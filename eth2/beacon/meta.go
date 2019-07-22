package beacon

import . "github.com/protolambda/zrnt/eth2/meta"

type UltraLightMeta interface {
	Versioning
}

type LightMeta interface {
	CompactCommittees
	Finality
	// TODO: define light client props
}

// TODO: clean up, can combine interfaces by categorization
type FullMeta interface {
	Exits
	Balance
	RegistrySize
	Pubkeys
	EffectiveBalances
	Finality
	AttesterStatuses
	SlashedIndices
	Staking
	TargetStaking
	Slashing
	Slasher
	Validators
	Versioning
	Deposits
	Onboarding
	Depositing
	LatestHeader
	LatestHeaderUpdate
	History
	HistoryUpdate
	EpochSeed
	Proposers
	ActivationExit
	ActiveValidatorCount
	ActiveIndices
	ActiveIndexRoots
	CommitteeCount
	CrosslinkTiming
	ShardRotation
	CrosslinkCommittees
	Crosslinks
	WinningCrosslinks
	Randomness
}
