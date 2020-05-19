package phase0

import (
	. "github.com/protolambda/zrnt/eth2/beacon/attestations"
	. "github.com/protolambda/zrnt/eth2/beacon/deposits"
	. "github.com/protolambda/zrnt/eth2/beacon/exits"
	. "github.com/protolambda/zrnt/eth2/beacon/finality"
	. "github.com/protolambda/zrnt/eth2/beacon/finalupdates"
	. "github.com/protolambda/zrnt/eth2/beacon/header"
	. "github.com/protolambda/zrnt/eth2/beacon/proposing"
	. "github.com/protolambda/zrnt/eth2/beacon/randao"
	. "github.com/protolambda/zrnt/eth2/beacon/registry"
	. "github.com/protolambda/zrnt/eth2/beacon/rewardpenalty"
	. "github.com/protolambda/zrnt/eth2/beacon/seeding"
	. "github.com/protolambda/zrnt/eth2/beacon/shuffling"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings/attslash"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings/propslash"
	. "github.com/protolambda/zrnt/eth2/beacon/transition"
	. "github.com/protolambda/zrnt/eth2/core"
)

// Full feature set for phase 0
type FullFeaturedState struct {
	// All base features a state has
	*BeaconState

	ShufflingFeature
	*ShufflingStatus

	AttesterStatusFeature

	ProposingFeature
	*ProposersData

	// Rewarding process, optimized to use precomputed crosslink/shuffling/etc. data
	AttestationRewardsAndPenaltiesFeature // rewards/penalties computation for attestations

	SeedFeature

	JustificationFeature
	RewardsAndPenaltiesFeature
	RegistryUpdatesFeature
	SlashingFeature
	FinalUpdateFeature

	RandaoFeature
	BlockHeaderFeature

	// Process block operations
	AttestationFeature
	AttestSlashFeature
	PropSlashFeature
	DepositFeature
	VoluntaryExitFeature

	SlotProcessFeature
	EpochProcessFeature
	TransitionFeature
}

func (f *FullFeaturedState) LoadPrecomputedData() {
	// TODO: could re-use some pre-computed data from older states, worth benchmarking
	f.ShufflingStatus = f.ShufflingFeature.LoadShufflingStatus()
	f.ProposersData = f.LoadBeaconProposersData()
}

func (f *FullFeaturedState) RotateEpochData() {
	// TODO: rotate data where possible (e.g. shuffling) instead of plain overwriting
	f.LoadPrecomputedData()
}

func (f *FullFeaturedState) StartEpoch() {
	f.RotateEpochData()
}

func (f *FullFeaturedState) CurrentProposer() BLSPubkey {
	return f.Pubkey(f.GetBeaconProposerIndex(f.CurrentSlot()))
}

func NewFullFeaturedState(state *BeaconState) *FullFeaturedState {
	// The con of heavy composition: it needs to be hooked up at the upper abstraction level
	// for cross references through interfaces to work.
	f := new(FullFeaturedState)

	// add state
	f.BeaconState = state

	// hook up features
	f.ShufflingFeature.Meta = f

	f.AttesterStatusFeature.State = &f.AttestationsState
	f.AttesterStatusFeature.Meta = f

	f.AttestationRewardsAndPenaltiesFeature.Meta = f

	f.SeedFeature.Meta = f
	f.ProposingFeature.Meta = f

	// TODO: disabled for now, need to implement "meta.TargetStaking"
	f.JustificationFeature.Meta = f
	f.JustificationFeature.State = &f.FinalityState
	f.RewardsAndPenaltiesFeature.Meta = f
	f.RegistryUpdatesFeature.Meta = f
	f.RegistryUpdatesFeature.State = &f.RegistryState
	f.SlashingFeature.Meta = f
	f.SlashingFeature.State = &f.SlashingsState
	f.FinalUpdateFeature.Meta = f

	f.RandaoFeature.Meta = f
	f.RandaoFeature.State = &f.RandaoState

	f.BlockHeaderFeature.Meta = f
	f.BlockHeaderFeature.State = &f.BlockHeaderState

	f.AttestationFeature.Meta = f
	f.AttestationFeature.State = &f.AttestationsState
	f.AttestSlashFeature.Meta = f
	f.PropSlashFeature.Meta = f
	f.DepositFeature.Meta = f
	f.VoluntaryExitFeature.Meta = f

	f.SlotProcessFeature.Meta = f
	f.EpochProcessFeature.Meta = f
	f.TransitionFeature.Meta = f

	return f
}
