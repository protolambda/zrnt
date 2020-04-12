package beacon

import (
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

var BeaconStateSSZ = zssz.GetSSZ((*BeaconState)(nil))

type BeaconState struct {
	// Versioning
	GenesisTime           Timestamp
	GenesisValidatorsRoot Root
	Slot                  Slot
	Fork                  Fork
	// History
	LatestBlockHeader BeaconBlockHeader
	HistoricalBatch   // embedded BlockRoots and StateRoots
	HistoricalRoots   HistoricalRoots
	// Eth1
	Eth1Data      Eth1Data
	Eth1DataVotes Eth1DataVotes
	DepositIndex  DepositIndex
	// Registry
	Validators ValidatorRegistry
	Balances   Balances
	// Randomness
	RandaoMixes [EPOCHS_PER_HISTORICAL_VECTOR]Root
	// Slashings
	Slashings [EPOCHS_PER_SLASHINGS_VECTOR]Gwei
	// Attestations
	PreviousEpochAttestations PendingAttestations
	CurrentEpochAttestations  PendingAttestations
	// Finality
	JustificationBits           JustificationBits
	PreviousJustifiedCheckpoint Checkpoint
	CurrentJustifiedCheckpoint  Checkpoint
	FinalizedCheckpoint         Checkpoint
}

// Beacon state
var BeaconStateType = ContainerType("BeaconState", []FieldDef{
	// Versioning
	{"genesis_time", Uint64Type},
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
	{"current_epoch_attestations", PendingAttestationsType},
	// Finality
	{"justification_bits", JustificationBitsType},     // Bit set for every recent justified epoch
	{"previous_justified_checkpoint", CheckpointType}, // Previous epoch snapshot
	{"current_justified_checkpoint", CheckpointType},
	{"finalized_checkpoint", CheckpointType},
})

func AsBeaconStateView(v View, err error) (*BeaconStateView, error) {
	c, err := AsContainer(v, err)
	return &BeaconStateView{c}, err
}

type BeaconStateView struct {
	*ContainerView
}

func (state *BeaconStateView) GenesisTime() (Timestamp, error) {
	return AsTimestamp(state.Get(0))
}

func (state *BeaconStateView) Slot() (Slot, error) {
	return AsSlot(state.Get(1))
}

func (state *BeaconStateView) SetSlot(slot Slot) error {
	return state.Set(1, Uint64View(slot))
}

func (state *BeaconStateView) Fork() (*ForkView, error) {
	return AsFork(state.Get(2))
}

func (state *BeaconStateView) LatestBlockHeader() (*BeaconBlockHeaderView, error) {
	return AsBeaconBlockHeader(state.Get(3))
}

func (state *BeaconStateView) SetLatestBlockHeader(v *BeaconBlockHeaderView) error {
	return state.Set(3, v)
}

func (state *BeaconStateView) BlockRoots() (*BatchRootsView, error) {
	return AsBatchRoots(state.Get(4))
}

func (state *BeaconStateView) StateRoots() (*BatchRootsView, error) {
	return AsBatchRoots(state.Get(5))
}

func (state *BeaconStateView) HistoricalRoots() (*HistoricalRootsView, error) {
	return AsHistoricalRoots(state.Get(6))
}

func (state *BeaconStateView) Eth1Data() (*Eth1DataView, error) {
	return AsEth1Data(state.Get(7))
}
func (state *BeaconStateView) SetEth1Data(v *Eth1DataView) error {
	return state.Set(7, v)
}

func (state *BeaconStateView) Eth1DataVotes() (*Eth1DataVotesView, error) {
	return AsEth1DataVotes(state.Get(8))
}

func (state *BeaconStateView) DepositIndex() (DepositIndex, error) {
	return AsDepositIndex(state.Get(9))
}

func (state *BeaconStateView) IncrementDepositIndex() error {
	depIndex, err := state.DepositIndex()
	if err != nil {
		return err
	}
	return state.Set(9, Uint64View(depIndex + 1))
}

func (state *BeaconStateView) Validators() (*ValidatorsRegistryView, error) {
	return AsValidatorsRegistry(state.Get(10))
}

func (state *BeaconStateView) Balances() (*RegistryBalancesView, error) {
	return AsRegistryBalances(state.Get(11))
}

func (state *BeaconStateView) RandaoMixes() (*RandaoMixesView, error) {
	return AsRandaoMixes(state.Get(12))
}

func (state *BeaconStateView) Slashings() (*SlashingsView, error) {
	return AsSlashings(state.Get(13))
}

func (state *BeaconStateView) PreviousEpochAttestations() (*PendingAttestationsView, error) {
	return AsPendingAttestations(state.Get(14))
}

func (state *BeaconStateView) CurrentEpochAttestations() (*PendingAttestationsView, error) {
	return AsPendingAttestations(state.Get(15))
}

func (state *BeaconStateView) JustificationBits() (*JustificationBitsView, error) {
	return AsJustificationBits(state.Get(16))
}

func (state *BeaconStateView) PreviousJustifiedCheckpoint() (*CheckpointView, error) {
	return AsCheckPoint(state.Get(17))
}

func (state *BeaconStateView) CurrentJustifiedCheckpoint() (*CheckpointView, error) {
	return AsCheckPoint(state.Get(18))
}

func (state *BeaconStateView) FinalizedCheckpoint() (*CheckpointView, error) {
	return AsCheckPoint(state.Get(19))
}

func (state *BeaconStateView) IsValidIndex(index ValidatorIndex) (bool, error) {
	vals, err := state.Validators()
	if err != nil {
		return false, err
	}
	count, err := vals.Length()
	if err != nil {
		return false, err
	}
	return uint64(index) < count, nil
}
