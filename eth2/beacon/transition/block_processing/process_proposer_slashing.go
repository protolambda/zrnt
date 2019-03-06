package block_processing

import (
	"errors"
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/transition"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func ProcessProposerSlashings(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.Proposer_slashings) > eth2.MAX_PROPOSER_SLASHINGS {
		return errors.New("too many proposer slashings")
	}
	for _, ps := range block.Body.Proposer_slashings {
		if err := ProcessProposerSlashing(state, &ps); err != nil {
			return err
		}
	}
	return nil
}

func ProcessProposerSlashing(state *beacon.BeaconState, ps *beacon.ProposerSlashing) error {
	if !transition.Is_validator_index(state, ps.Proposer_index) {
		return errors.New("invalid proposer index")
	}
	proposer := &state.Validator_registry[ps.Proposer_index]
	if !(ps.Header_1.Slot == ps.Header_2.Slot &&
		ps.Header_1.BlockBodyRoot != ps.Header_2.BlockBodyRoot && proposer.Slashed == false &&
		bls.Bls_verify(proposer.Pubkey, ssz.Signed_root(ps.Header_1), ps.Header_1.Signature, transition.Get_domain(state.Fork, ps.Header_1.Slot.ToEpoch(), eth2.DOMAIN_BEACON_BLOCK)) &&
		bls.Bls_verify(proposer.Pubkey, ssz.Signed_root(ps.Header_2), ps.Header_2.Signature, transition.Get_domain(state.Fork, ps.Header_2.Slot.ToEpoch(), eth2.DOMAIN_BEACON_BLOCK))) {
		return errors.New("proposer slashing is invalid")
	}
	if err := transition.Slash_validator(state, ps.Proposer_index); err != nil {
		return err
	}
}
