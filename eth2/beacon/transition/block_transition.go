package transition

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/attestations"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/block_header"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/eth1"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/exits/voluntary_exits"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/randao"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/slashing/attester_slashing"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/slashing/proposer_slashing"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/validator_balances/deposits"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/processes/validator_balances/transfers"
)

type BlockProcessor func(staten *beacon.BeaconState, block *beacon.BeaconBlock) error

var blockProcessors = []BlockProcessor{
	block_header.ProcessBlockHeader,
	randao.ProcessBlockRandao,
	eth1.ProcessBlockEth1,
	proposer_slashing.ProcessBlockProposerSlashings,
	attester_slashing.ProcessBlockAttesterSlashings,
	attestations.ProcessBlockAttestations,
	deposits.ProcessBlockDeposits,
	voluntary_exits.ProcessBlockVoluntaryExits,
	transfers.ProcessBlockTransfers,
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
