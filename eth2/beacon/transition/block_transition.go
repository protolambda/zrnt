package transition

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/beacon/block_processing"
)

type BlockProcessor func(state *BeaconState, block *BeaconBlock) error

var blockProcessors = []BlockProcessor{
	ProcessBlockHeader,
	ProcessBlockRandao,
	ProcessBlockEth1,
	// --- Transactions ---
	ProcessBlockProposerSlashings,
	ProcessBlockAttesterSlashings,
	ProcessBlockAttestations,
	ProcessBlockDeposits,
	ProcessBlockVoluntaryExits,
	ProcessBlockTransfers,
	// --------------------
}

// Applies all block-processing functions to the state, for the given block.
// Stops applying at first error.
func ProcessBlock(state *BeaconState, block *BeaconBlock) error {
	for _, p := range blockProcessors {
		if err := p(state, block); err != nil {
			return err
		}
	}
	return nil
}
