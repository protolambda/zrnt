package full

import (
	. "github.com/protolambda/zrnt/eth2/beacon/active"
	. "github.com/protolambda/zrnt/eth2/beacon/attestations"
	. "github.com/protolambda/zrnt/eth2/beacon/compact"
	. "github.com/protolambda/zrnt/eth2/beacon/crosslinks"
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
	. "github.com/protolambda/zrnt/eth2/beacon/shardrot"
	. "github.com/protolambda/zrnt/eth2/beacon/shuffling"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings/attslash"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings/propslash"
	. "github.com/protolambda/zrnt/eth2/beacon/transfers"
	. "github.com/protolambda/zrnt/eth2/beacon/transition"
	"github.com/protolambda/zrnt/eth2/phase0"
)

// Full feature set for phase 0
type FullFeatures struct {
	// All base features a state has
	*phase0.BeaconState

	ShardRotFeature
	*StartShardStatus

	ShufflingFeature
	*ShufflingStatus

	AttesterStatusFeature
	AttesterStatuses

	CrosslinkingFeature
	*CrosslinkingStatus

	ProposingFeature
	*EpochProposerIndices

	// Data computation to serve light clients
	ActiveRootsFeature       // roots of past active indices lists
	CompactCommitteesFeature // roots of past crosslink committees, in compact form (minimal validator data)

	// Rewarding process, optimized to use precomputed crosslink/shuffling/etc. data
	CrosslinkDeltasFeature   // rewards/penalties computation for crosslinking
	AttestationDeltasFeature // rewards/penalties computation for attestations

	SeedFeature

	JustificationFeature
	CrosslinksFeature
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
	TransferFeature
	VoluntaryExitFeature

	phase0.SlotProcessFeature
	phase0.EpochProcessFeature
	TransitionFeature
}

func (f *FullFeatures) Load(state *phase0.BeaconState) {
	// The con of heavy composition: it needs to be hooked up at the upper abstraction level
	// for cross references through interfaces to work.

	// add state
	f.BeaconState = state

	// hook up features
	f.ShufflingFeature.Meta = f

	f.AttesterStatusFeature.State = &f.AttestationsState
	f.AttesterStatusFeature.Meta = f
	f.AttesterStatusFeature.State = &f.AttestationsState
	f.CrosslinkingFeature.Meta = f
	f.CrosslinkingFeature.State = &f.AttestationsState

	f.ShardRotFeature.Meta = f
	f.ShardRotFeature.State = &f.ShardRotationState

	f.ActiveRootsFeature.Meta = f
	f.ActiveRootsFeature.State = &f.ActiveState

	f.CompactCommitteesFeature.Meta = f
	f.CompactCommitteesFeature.State = &f.CompactCommitteesState

	f.CrosslinkDeltasFeature.Meta = f
	f.CrosslinkDeltasFeature.State = &f.CrosslinksState
	f.AttestationDeltasFeature.Meta = f

	f.SeedFeature.Meta = f
	f.ProposingFeature.Meta = f

	// TODO: disabled for now, need to implement "meta.TargetStaking"
	//f.JustificationFeature.Meta = f
	f.JustificationFeature.State = &f.FinalityState
	f.CrosslinksFeature.Meta = f
	f.CrosslinksFeature.State = &f.CrosslinksState
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
	f.TransferFeature.Meta = f
	f.VoluntaryExitFeature.Meta = f

	f.SlotProcessFeature.Meta = f
	f.EpochProcessFeature.Meta = f
	f.TransitionFeature.Meta = f

	// pre-compute the data!
	// TODO: could re-use some pre-computed data from older states, worth benchmarking
	// TODO decide on some lookback time, or load it dynamically
	f.StartShardStatus = f.ShardRotFeature.LoadStartShardStatus(f.CurrentEpoch() - 20)
	f.ShufflingStatus = f.ShufflingFeature.LoadShufflingStatus()
	f.CrosslinkingStatus = f.CrosslinkingFeature.LoadCrosslinkingStatus()
	f.AttesterStatuses = f.AttesterStatusFeature.LoadAttesterStatuses()
	f.EpochProposerIndices = f.LoadBeaconProposerIndices()
}
