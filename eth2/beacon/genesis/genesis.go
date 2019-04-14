package genesis

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	. "github.com/protolambda/zrnt/eth2/util/data_types"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func GetGenesisBeaconState(validatorDeposits []beacon.Deposit, time beacon.Timestamp, eth1Data beacon.Eth1Data) *beacon.BeaconState {
	state := &beacon.BeaconState{
		Slot:        beacon.GENESIS_SLOT,
		GenesisTime: time,
		Fork: beacon.Fork{
			// genesis fork versions are 0
			PreviousVersion: beacon.Int32ToForkVersion(beacon.GENESIS_FORK_VERSION),
			CurrentVersion:  beacon.Int32ToForkVersion(beacon.GENESIS_FORK_VERSION),
			Epoch:           beacon.GENESIS_EPOCH,
		},
		// Validator registry
		ValidatorRegistry: make(beacon.ValidatorRegistry, 0),
		Balances: make([]beacon.Gwei, 0),
		ValidatorRegistryUpdateEpoch: beacon.GENESIS_EPOCH,
		// Randomness and committees
		LatestStartShard: beacon.GENESIS_START_SHARD,
		// Finality
		PreviousEpochAttestations: make([]beacon.PendingAttestation, 0),
		CurrentEpochAttestations: make([]beacon.PendingAttestation, 0),
		PreviousJustifiedEpoch: beacon.GENESIS_EPOCH - 1,
		CurrentJustifiedEpoch:  beacon.GENESIS_EPOCH,
		FinalizedEpoch:         beacon.GENESIS_EPOCH,
		// Recent state
		LatestBlockHeader: beacon.GetEmptyBlock().GetTemporaryBlockHeader(),
		HistoricalRoots: make([]Root, 0),
		// Ethereum 1.0 chain data
		LatestEth1Data: eth1Data,
	}
	// Initialize crosslinks
	for i := beacon.Shard(0); i < beacon.SHARD_COUNT; i++ {
		state.LatestCrosslinks[i] = beacon.Crosslink{Epoch: beacon.GENESIS_EPOCH}
	}
	// Process genesis deposits
	for _, dep := range validatorDeposits {
		// in the rare case someone tries to create a genesis block using invalid data, panic.
		if err := block_processing.ProcessDeposit(state, &dep); err != nil {
			panic(err)
		}
	}
	for i := range state.ValidatorRegistry {
		if state.GetEffectiveBalance(beacon.ValidatorIndex(i)) >= beacon.MAX_DEPOSIT_AMOUNT {
			state.ActivateValidator(beacon.ValidatorIndex(i), true)
		}
	}
	genesisActiveIndexRoot := ssz.HashTreeRoot(
		state.ValidatorRegistry.GetActiveValidatorIndices(beacon.GENESIS_EPOCH))
	for i := beacon.Epoch(0); i < beacon.LATEST_ACTIVE_INDEX_ROOTS_LENGTH; i++ {
		state.LatestActiveIndexRoots[i] = genesisActiveIndexRoot
	}
	return state
}
