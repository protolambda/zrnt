package beacon

import (
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func GetGenesisBeaconState(validatorDeposits []Deposit, time Timestamp, eth1Data Eth1Data) *BeaconState {
	state := &BeaconState{
		Slot:        GENESIS_SLOT,
		GenesisTime: time,
		Fork: Fork{
			// genesis fork versions are 0
			PreviousVersion: Int32ToForkVersion(GENESIS_FORK_VERSION),
			CurrentVersion:  Int32ToForkVersion(GENESIS_FORK_VERSION),
			Epoch:           GENESIS_EPOCH,
		},
		// Validator registry
		ValidatorRegistry: make(ValidatorRegistry, 0),
		Balances: make([]Gwei, 0),
		ValidatorRegistryUpdateEpoch: GENESIS_EPOCH,
		// Randomness and committees
		LatestStartShard: GENESIS_START_SHARD,
		// Finality
		PreviousEpochAttestations: make([]PendingAttestation, 0),
		CurrentEpochAttestations: make([]PendingAttestation, 0),
		PreviousJustifiedEpoch: GENESIS_EPOCH - 1,
		CurrentJustifiedEpoch:  GENESIS_EPOCH,
		FinalizedEpoch:         GENESIS_EPOCH,
		// Recent state
		LatestBlockHeader: GetEmptyBlock().GetTemporaryBlockHeader(),
		HistoricalRoots: make([]Root, 0),
		// Ethereum 1.0 chain data
		LatestEth1Data: eth1Data,
	}
	// Initialize crosslinks
	for i := Shard(0); i < SHARD_COUNT; i++ {
		state.LatestCrosslinks[i] = Crosslink{Epoch: GENESIS_EPOCH}
	}
	// Process genesis deposits
	for _, dep := range validatorDeposits {
		// in the rare case someone tries to create a genesis block using invalid data, panic.
		if err := block_processing.ProcessDeposit(state, &dep); err != nil {
			panic(err)
		}
	}
	for i := range state.ValidatorRegistry {
		if state.GetEffectiveBalance(ValidatorIndex(i)) >= MAX_DEPOSIT_AMOUNT {
			state.ActivateValidator(ValidatorIndex(i), true)
		}
	}
	genesisActiveIndexRoot := ssz.HashTreeRoot(
		state.ValidatorRegistry.GetActiveValidatorIndices(GENESIS_EPOCH))
	for i := Epoch(0); i < LATEST_ACTIVE_INDEX_ROOTS_LENGTH; i++ {
		state.LatestActiveIndexRoots[i] = genesisActiveIndexRoot
	}
	return state
}
