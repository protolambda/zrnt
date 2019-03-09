package transition

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
)

type BlockProcessor func(state *beacon.BeaconState, block *beacon.BeaconBlock) error

var blockProcessors = []BlockProcessor{
	block_processing.ProcessBlockHeader,
	block_processing.ProcessBlockRandao,
	block_processing.ProcessBlockEth1,
	block_processing.ProcessBlockProposerSlashings,
	block_processing.ProcessBlockAttesterSlashings,
	block_processing.ProcessBlockAttestations,
	block_processing.ProcessBlockDeposits,
	block_processing.ProcessBlockVoluntaryExits,
	block_processing.ProcessBlockTransfers,
}

// Applies all block-processing functions to the state, for the given block.
// Stops applying at first error.
func ApplyBlock(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	for _, p := range blockProcessors {
		if err := p(state, block); err != nil {
			return err
		}
	}
	return nil
}
