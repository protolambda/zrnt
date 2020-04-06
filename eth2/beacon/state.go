package beacon

import (
	. "github.com/protolambda/ztyp/view"
)

// Beacon state
var BeaconStateType = ContainerType("BeaconState", []FieldDef{
	// Versioning
	{"genesis_time", Uint64Type},
	{"genesis_validators_root", RootType},
	{"slot", SlotType},
	{"fork", ForkType},
	// History
	{"latest_block_header", BeaconBlockHeaderType},
	{"block_roots", BatchRootsType},
	{"state_roots", BatchRootsType},
	{"historical_roots", HistoricalRootsType},
	// Eth1
	{"eth1_data", Eth1DataType},
	{"eth1_data_votes", Eth1DataVotesType},
	{"eth1_deposit_index", Uint64Type},
	// Registry
	{"validators", ValidatorsRegistryType},
	{"balances", RegistryBalancesType},
	// Randomness
	{"randao_mixes", RandaoMixesType},
	// Slashings
	{"slashings", SlashingsType}, // Per-epoch sums of slashed effective balances
	// Attestations
	{"previous_epoch_attestations", PendingAttestationsType},
	{"current_epoch_attestations", PendingAttestationType},
	// Finality
	{"justification_bits", JustificationBitsType},     // Bit set for every recent justified epoch
	{"previous_justified_checkpoint", CheckpointType}, // Previous epoch snapshot
	{"current_justified_checkpoint", CheckpointType},
	{"finalized_checkpoint", CheckpointType},
})

type BeaconStateView struct {
	*ContainerView
}

func (state *BeaconStateView) GenesisTime() (Timestamp, error) {
	return AsTimestamp(state.Get(0))
}

func (state *BeaconStateView) GenesisValidatorsRoot() (Root, error) {
	return AsRoot(state.Get(1))
}

func (state *BeaconStateView) Slot() (Slot, error) {
	return AsSlot(state.Get(2))
}

func (state *BeaconStateView) SetSlot(slot Slot) error {
	return state.Set(2, Uint64View(slot))
}

func (state *BeaconStateView) Fork() (*ForkView, error) {
	return AsFork(state.Get(3))
}

func (state *BeaconStateView) LatestBlockHeader() (*BeaconBlockHeaderView, error) {
	return AsBeaconBlockHeader(state.Get(4))
}

func (state *BeaconStateView) SetLatestBlockHeader(v *BeaconBlockHeaderView) error {
	return state.Set(4, v)
}

func (state *BeaconStateView) BlockRoots() (*BatchRootsView, error) {
	return AsBatchRoots(state.Get(5))
}

func (state *BeaconStateView) StateRoots() (*BatchRootsView, error) {
	return AsBatchRoots(state.Get(6))
}

func (state *BeaconStateView) HistoricalRoots() (*HistoricalRootsView, error) {
	return AsHistoricalRoots(state.Get(7))
}

func (state *BeaconStateView) Eth1Data() (*Eth1DataView, error) {
	return AsEth1Data(state.Get(8))
}
func (state *BeaconStateView) SetEth1Data(v *Eth1DataView) error {
	return state.Set(8, v)
}

func (state *BeaconStateView) Eth1DataVotes() (*Eth1DataVotesView, error) {
	return AsEth1DataVotes(state.Get(9))
}

func (state *BeaconStateView) DepositIndex() (DepositIndex, error) {
	return AsDepositIndex(state.Get(10))
}

func (state *BeaconStateView) IncrementDepositIndex() error {
	depIndex, err := state.DepositIndex()
	if err != nil {
		return err
	}
	return state.Set(10, Uint64View(depIndex))
}

func (state *BeaconStateView) Validators() (*ValidatorsRegistryView, error) {
	return AsValidatorsRegistry(state.Get(11))
}

func (state *BeaconStateView) Balances() (*RegistryBalancesView, error) {
	return AsRegistryBalances(state.Get(12))
}

func (state *BeaconStateView) RandaoMixes() (*RandaoMixesView, error) {
	return AsRandaoMixes(state.Get(13))
}

func (state *BeaconStateView) Slashings() (*SlashingsView, error) {
	return AsSlashings(state.Get(14))
}

func (state *BeaconStateView) PreviousEpochAttestations() (*PendingAttestationsView, error) {
	return AsPendingAttestations(state.Get(15))
}

func (state *BeaconStateView) CurrentEpochAttestations() (*PendingAttestationsView, error) {
	return AsPendingAttestations(state.Get(16))
}

func (state *BeaconStateView) JustificationBits() (*JustificationBitsView, error) {
	return AsJustificationBits(state.Get(17))
}

func (state *BeaconStateView) PreviousJustifiedCheckpoint() (*CheckpointView, error) {
	return AsCheckPoint(state.Get(18))
}

func (state *BeaconStateView) CurrentJustifiedCheckpoint() (*CheckpointView, error) {
	return AsCheckPoint(state.Get(19))
}

func (state *BeaconStateView) FinalizedCheckpoint() (*CheckpointView, error) {
	return AsCheckPoint(state.Get(20))
}
