package operations

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

type ProposerSlashings []ProposerSlashing

func (_ *ProposerSlashings) Limit() uint32 {
	return MAX_PROPOSER_SLASHINGS
}

func (ops ProposerSlashings) Process(state *BeaconState) error {
	for _, op := range ops {
		if err := op.Process(state); err != nil {
			return err
		}
	}
	return nil
}

type ProposerSlashing struct {
	ProposerIndex ValidatorIndex
	Header1       BeaconBlockHeader // First proposal
	Header2       BeaconBlockHeader // Second proposal
}

func (ps *ProposerSlashing) Process(state *BeaconState) error {
	if !state.Validators.IsValidatorIndex(ps.ProposerIndex) {
		return errors.New("invalid proposer index")
	}
	proposer := state.Validators[ps.ProposerIndex]
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
	if !bls.BlsVerify(proposer.Pubkey, ssz.SigningRoot(ps.Header1, BeaconBlockHeaderSSZ), ps.Header1.Signature,
		state.GetDomain(DOMAIN_BEACON_PROPOSER, ps.Header1.Slot.ToEpoch())) {
		return errors.New("proposer slashing header 1 has invalid BLS signature")
	}
	if !bls.BlsVerify(proposer.Pubkey, ssz.SigningRoot(ps.Header2, BeaconBlockHeaderSSZ), ps.Header2.Signature,
		state.GetDomain(DOMAIN_BEACON_PROPOSER, ps.Header2.Slot.ToEpoch())) {
		return errors.New("proposer slashing header 2 has invalid BLS signature")
	}
	state.SlashValidator(ps.ProposerIndex, nil)
	return nil
}
