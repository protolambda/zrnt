package block_processing

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessBlockProposerSlashings(state *BeaconState, block *BeaconBlock) error {
	if len(block.Body.ProposerSlashings) > MAX_PROPOSER_SLASHINGS {
		return errors.New("too many proposer slashings")
	}
	for _, ps := range block.Body.ProposerSlashings {
		if err := ProcessProposerSlashing(state, &ps); err != nil {
			return err
		}
	}
	return nil
}

func ProcessProposerSlashing(state *BeaconState, ps *ProposerSlashing) error {
	if !state.ValidatorRegistry.IsValidatorIndex(ps.ProposerIndex) {
		return errors.New("invalid proposer index")
	}
	proposer := state.ValidatorRegistry[ps.ProposerIndex]
	// Verify that the epoch is the same
	if ps.Header1.Slot.ToEpoch() != ps.Header2.Slot.ToEpoch() {
		return errors.New("proposer slashing requires slashing headers to be in same epoch")
	}
	// But the headers are different
	if ps.Header1 == ps.Header2 {
		return errors.New("proposer slashing requires two different headers")
	}
	// Check proposer is slashable
	if !proposer.IsSlashable(state.Epoch()) {
		return errors.New("proposer slashing requires proposer to be slashable")
	}
	// Signatures are valid
	if !(
		bls.BlsVerify(proposer.Pubkey, ssz.SigningRoot(ps.Header1), ps.Header1.Signature, GetDomain(state.Fork, ps.Header1.Slot.ToEpoch(), DOMAIN_BEACON_BLOCK)) &&
		bls.BlsVerify(proposer.Pubkey, ssz.SigningRoot(ps.Header2), ps.Header2.Signature, GetDomain(state.Fork, ps.Header2.Slot.ToEpoch(), DOMAIN_BEACON_BLOCK))) {
		return errors.New("proposer slashing has header with invalid BLS signature")
	}
	return state.SlashValidator(ps.ProposerIndex)
}
