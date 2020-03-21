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
	*BeaconStateProps

	ShufflingFeature
	*ShufflingStatus

	AttesterStatusFeature

	ProposingFeature
	*ProposersData

	// Rewarding process, optimized to use precomputed crosslink/shuffling/etc. data
	AttestationDeltasFeature // rewards/penalties computation for attestations

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

func (f *FullFeaturedState) LoadPrecomputedData() error {
	// TODO: could re-use some pre-computed data from older states, worth benchmarking
	var err error
	f.ShufflingStatus, err = f.ShufflingFeature.LoadShufflingStatus()
	if err != nil {
		return err
	}
	f.ProposersData, err = f.LoadBeaconProposersData()
	if err != nil {
		return err
	}
}

func (f *FullFeaturedState) RotateEpochData() error {
	// TODO: rotate data where possible (e.g. shuffling) instead of plain overwriting
	return f.LoadPrecomputedData()
}

func (f *FullFeaturedState) StartEpoch() error {
	return f.RotateEpochData()
}

func (f *FullFeaturedState) CurrentProposer() (BLSPubkey, error) {
	slot, err := f.CurrentSlot()
	if err != nil {
		return BLSPubkey{}, err
	}
	index, err := f.GetBeaconProposerIndex(slot)
	if err != nil {
		return BLSPubkey{}, err
	}
	return f.Pubkey(index)
}

func NewFullFeaturedState(state *BeaconStateView) *FullFeaturedState {
	// The con of heavy composition: it needs to be hooked up at the upper abstraction level
	// for cross references through interfaces to work.
	f := new(FullFeaturedState)

	// add state
	f.BeaconStateProps = state.Props()

	// hook up features
	f.ShufflingFeature.Meta = f

	f.AttesterStatusFeature.State = &f.AttestationsProps
	f.AttesterStatusFeature.Meta = f

	f.AttestationDeltasFeature.Meta = f

	f.SeedFeature.Meta = f
	f.ProposingFeature.Meta = f

	f.JustificationFeature.Meta = f
	f.JustificationFeature.State = f.FinalityProps
	f.RewardsAndPenaltiesFeature.Meta = f
	f.RegistryUpdatesFeature.Meta = f
	f.RegistryUpdatesFeature.State = &f.RegistryState
	f.SlashingFeature.Meta = f
	f.SlashingFeature.State = f.SlashingsProp
	f.FinalUpdateFeature.Meta = f

	f.RandaoFeature.Meta = f
	f.RandaoFeature.State = f.RandaoMixesProp

	f.BlockHeaderFeature.Meta = f
	f.BlockHeaderFeature.State = f.LatestBlockHeaderProp

	f.AttestationFeature.Meta = f
	f.AttestationFeature.State = f.AttestationsProps
	f.AttestSlashFeature.Meta = f
	f.PropSlashFeature.Meta = f
	f.DepositFeature.Meta = f
	f.VoluntaryExitFeature.Meta = f

	f.SlotProcessFeature.Meta = f
	f.EpochProcessFeature.Meta = f
	f.TransitionFeature.Meta = f

	return f
}
