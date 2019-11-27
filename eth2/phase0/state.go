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
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

// Beacon state
var BeaconState = &ContainerType{
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
	{"eth1_deposit_index", Uint64TypeType},
	// Registry
	{"validators", RegistryValidatorsType},
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

type BeaconStateView struct {
	*ContainerView
}

func (state *BeaconState) StateRoot() Root {
	return ssz.HashTreeRoot(state, BeaconStateSSZ)
}

func (state *BeaconState) IncrementSlot() {
	state.Slot++
}
