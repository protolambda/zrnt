package beacon

import . "github.com/protolambda/zrnt/eth2/meta"

type UltraLightMeta interface {
	VersioningMeta
}

type LightMeta interface {
	CompactValidatorMeta
	FinalityMeta
	// TODO: define light client props
}

// TODO: clean up, can combine interfaces by categorization
type FullMeta interface {
	ExitMeta
	BalanceMeta
	RegistrySizeMeta
	PubkeyMeta
	EffectiveBalanceMeta
	FinalityMeta
	AttesterStatusMeta
	SlashedMeta
	StakingMeta
	TargetStakingMeta
	SlashingMeta
	SlasherMeta
	ValidatorMeta
	VersioningMeta
	Eth1Meta
	OnboardMeta
	DepositMeta
	HeaderMeta
	UpdateHeaderMeta
	HistoryMeta
	HistoryUpdateMeta
	SeedMeta
	ProposingMeta
	ActivationExitMeta
	ActiveValidatorCountMeta
	ActiveIndicesMeta
	ActiveIndexRootMeta
	CommitteeCountMeta
	CrosslinkTimingMeta
	ShardRotMeta
	CrosslinkCommitteeMeta
	CrosslinkMeta
	WinningCrosslinkMeta
	RandomnessMeta
}
