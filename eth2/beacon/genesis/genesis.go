package genesis

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/validator_balances/deposits"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func GetGenesisBeaconState(validatorDeposits []beacon.Deposit, time beacon.Timestamp, eth1Data beacon.Eth1Data) *beacon.BeaconState {
	state := &beacon.BeaconState{
		Slot:        beacon.GENESIS_SLOT,
		GenesisTime: time,
		Fork: beacon.Fork{
			PreviousVersion: beacon.GENESIS_FORK_VERSION,
			CurrentVersion:  beacon.GENESIS_FORK_VERSION,
			Epoch:           beacon.GENESIS_EPOCH,
		},
		// Validator registry
		ValidatorRegistryUpdateEpoch: beacon.GENESIS_EPOCH,
		// Randomness and committees
		PreviousShufflingStartShard: beacon.GENESIS_START_SHARD,
		CurrentShufflingStartShard:  beacon.GENESIS_START_SHARD,
		CurrentShufflingEpoch:       beacon.GENESIS_EPOCH,
		PreviousShufflingEpoch:      beacon.GENESIS_EPOCH,
		// Finality
		PreviousJustifiedEpoch: beacon.GENESIS_EPOCH,
		JustifiedEpoch:         beacon.GENESIS_EPOCH,
		FinalizedEpoch:         beacon.GENESIS_EPOCH,
		// Recent state
		LatestBlockHeader: beacon.GetEmptyBlock().GetTemporaryBlockHeader(),
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
		if err := deposits.ProcessDeposit(state, &dep); err != nil {
			panic(err)
		}
	}
	for i := range state.ValidatorRegistry {
		if state.ValidatorBalances.GetEffectiveBalance(beacon.ValidatorIndex(i)) >= beacon.MAX_DEPOSIT_AMOUNT {
			state.ActivateValidator(beacon.ValidatorIndex(i), true)
		}
	}
	genesisActiveIndexRoot := ssz.HashTreeRoot(
		state.ValidatorRegistry.GetActiveValidatorIndices(beacon.GENESIS_EPOCH))
	for i := beacon.Epoch(0); i < beacon.LATEST_ACTIVE_INDEX_ROOTS_LENGTH; i++ {
		state.LatestActiveIndexRoots[i] = genesisActiveIndexRoot
	}
	state.CurrentShufflingSeed = state.GenerateSeed(beacon.GENESIS_EPOCH)
	return state
}
