package proposer_slashing

import (
	"errors"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func ProcessBlockProposerSlashings(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.ProposerSlashings) > beacon.MAX_PROPOSER_SLASHINGS {
		return errors.New("too many proposer slashings")
	}
	for _, ps := range block.Body.ProposerSlashings {
		if err := ProcessProposerSlashing(state, &ps); err != nil {
			return err
		}
	}
	return nil
}

func ProcessProposerSlashing(state *beacon.BeaconState, ps *beacon.ProposerSlashing) error {
	if !state.ValidatorRegistry.IsValidatorIndex(ps.ProposerIndex) {
		return errors.New("invalid proposer index")
	}
	proposer := &state.ValidatorRegistry[ps.ProposerIndex]
	if !(ps.Header1.Slot == ps.Header2.Slot &&
		ps.Header1.BlockBodyRoot != ps.Header2.BlockBodyRoot && proposer.Slashed == false &&
		bls.BlsVerify(proposer.Pubkey, ssz.SignedRoot(ps.Header1), ps.Header1.Signature, beacon.GetDomain(state.Fork, ps.Header1.Slot.ToEpoch(), beacon.DOMAIN_BEACON_BLOCK)) &&
		bls.BlsVerify(proposer.Pubkey, ssz.SignedRoot(ps.Header2), ps.Header2.Signature, beacon.GetDomain(state.Fork, ps.Header2.Slot.ToEpoch(), beacon.DOMAIN_BEACON_BLOCK))) {
		return errors.New("proposer slashing is invalid")
	}
	if err := state.SlashValidator(ps.ProposerIndex); err != nil {
		return err
	}
}
