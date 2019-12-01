package phase0

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
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/ztyp/props"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

// Beacon state
var BeaconStateType = &ContainerType{
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
	{"current_epoch_attestations", PendingAttestationType},
	// Finality
	{"justification_bits", JustificationBitsType},     // Bit set for every recent justified epoch
	{"previous_justified_checkpoint", CheckpointType}, // Previous epoch snapshot
	{"current_justified_checkpoint", CheckpointType},
	{"finalized_checkpoint", CheckpointType},
}

// TODO: can also create an explicit read-only props view (to not rely on tree forking on modifications)

type BeaconStateMutProps struct {
	GenesisTimeProp
	CurrentSlotMutProp
	ForkProp
	// TODO remaining props
	SlashingsProp
	LatestHeaderProp
	RandaoMixesProp

	// history
	HistoryProps
	FinalityProps
}

type BeaconStateView struct {
	*ContainerView
}

// MutProps returns a mutable view of the BeaconState
func (state *BeaconStateView) MutProps() *BeaconStateMutProps {
	return &BeaconStateMutProps{
		GenesisTimeProp: GenesisTimeProp(PropReader(state, 0)),
		CurrentSlotMutProp: CurrentSlotMutProp{
			CurrentSlotReadProp: CurrentSlotReadProp(PropReader(state, 1)),
			SlotWriteProp: SlotWriteProp(PropWriter(state, 1)),
		},
		ForkProp:        ForkProp(PropReader(state, 2)),
		// TODO remaining props
		SlashingsProp:      SlashingsProp(PropReader(state, 123)),
		LatestHeaderProp: LatestHeaderProp(PropReader(state, 123)),
		RandaoMixesProp:      RandaoMixesProp(PropReader(state, 123)),
		HistoryProps: HistoryProps{
			BlockRootsProp:      BlockRootsProp(PropReader(state, 123)),
			StateRootsProp:      StateRootsProp(PropReader(state, 123)),
			HistoricalRootsProp: HistoricalRootsProp(PropReader(state, 123)),
		},
		FinalityProps: FinalityProps{
			JustificationBits:           JustificationBitsProp(PropReader(state, 123)),
			PreviousJustifiedCheckpoint: CheckpointProp(PropReader(state, 123)),
			CurrentJustifiedCheckpoint:  CheckpointProp(PropReader(state, 123)),
			FinalizedCheckpoint:         CheckpointProp(PropReader(state, 123)),
		},
	}
}

func (state *BeaconStateView) StateRoot() Root {
	return state.ViewRoot(tree.Hash)
}
