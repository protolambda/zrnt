package beacon

import (
	. "github.com/protolambda/zrnt/eth2/beacon/attestations"
	. "github.com/protolambda/zrnt/eth2/beacon/eth1"
	. "github.com/protolambda/zrnt/eth2/beacon/finality"
	. "github.com/protolambda/zrnt/eth2/beacon/header"
	. "github.com/protolambda/zrnt/eth2/beacon/history"
	. "github.com/protolambda/zrnt/eth2/beacon/randao"
	. "github.com/protolambda/zrnt/eth2/beacon/registry"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings"
	. "github.com/protolambda/zrnt/eth2/beacon/versioning"

	. "github.com/protolambda/ztyp/props"
	"github.com/protolambda/ztyp/tree"
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

func (state *BeaconStateView) Fork() (*ForkView, error) {
	return AsFork(state.Get(2))
}

// MutProps returns a mutable view of the BeaconState
func (state *BeaconStateView) Props() *BeaconStateProps {
	return &BeaconStateProps{
		VersioningProps: VersioningProps{
			GenesisTimeProp: GenesisTimeProp(PropReader(state, 0)),
			GenesisValidatorsRootProp: GenesisValidatorsRootProp(PropReader(state, 1)),
			CurrentSlotMutProp: CurrentSlotMutProp{
				CurrentSlotReadProp: CurrentSlotReadProp(PropReader(state, 2)),
				SlotWriteProp: SlotWriteProp(PropWriter(state, 3)),// TODO
			},
			ForkProp:        ForkProp(PropReader(state, 4)),
		},
		LatestBlockHeaderProp: LatestBlockHeaderProp(PropReader(state, 5)),
		HistoryProps: HistoryProps{
			BlockRootsProp:      BlockRootsProp(PropReader(state, 6)),
			StateRootsProp:      StateRootsProp(PropReader(state, 7)),
			HistoricalRootsProp: HistoricalRootsProp(PropReader(state, 8)),
		},
		// TODO remaining props
		RandaoMixesProp:      RandaoMixesProp(PropReader(state, 9)),
		SlashingsProp:      SlashingsProp(PropReader(state, 10)),
		AttestationsProps:  AttestationsProps{
			PreviousEpochAttestations: EpochPendingAttestationsProp(PropReader(state, 11)),
			CurrentEpochAttestations:  EpochPendingAttestationsProp(PropReader(state, 12)),
		},
		FinalityProps: FinalityProps{
			JustificationBits:           JustificationBitsProp(PropReader(state, 13)),
			PreviousJustifiedCheckpoint: CheckpointProp(PropReader(state, 14)),
			CurrentJustifiedCheckpoint:  CheckpointProp(PropReader(state, 15)),
			FinalizedCheckpoint:         CheckpointProp(PropReader(state, 16)),
		},
	}
}


func (state *BeaconStateView) StateRoot() Root {
	return state.HashTreeRoot(tree.GetHashFn())
}
