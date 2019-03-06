package transition

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/transition/block_processing"
)

// Applies all block-processing functions to the state, for the given block.
// Stops applying at first error.
func ApplyBlock(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	// Verify header
	if err := block_processing.ProcessBlockHeader(state, block); err != nil {
		return nil
	}

	//// RANDAO
	if err := block_processing.ProcessRandao(state, block); err != nil {
		return err
	}

	// Eth1 data
	if err := block_processing.ProcessEth1(state, block); err != nil {
		return err
	}

	// Transactions
	// START ------------------------------

	// Proposer slashings
	if err := block_processing.ProcessProposerSlashings(state, block); err != nil {
		return err
	}

	// Attester slashings
	if err := block_processing.ProcessAttesterSlashings(state, block); err != nil {
		return err
	}


	// Attestations
	if err := block_processing.ProcessAttestations(state, block); err != nil {
		return err
	}


	// Deposits
	if err := block_processing.ProcessDeposits(state, block); err != nil {
		return err
	}

	// Voluntary exits
	if err := block_processing.ProcessVoluntaryExits(state, block); err != nil {
		return err
	}

	// Transfers
	if err := block_processing.ProcessTransfers(state, block); err != nil {
		return err
	}

	// END ------------------------------

	return nil
}