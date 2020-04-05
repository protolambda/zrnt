package beacon

import (
	"errors"

	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

var ProposerSlashingSSZ = zssz.GetSSZ((*ProposerSlashing)(nil))

type ProposerSlashing struct {
	SignedHeader1       SignedBeaconBlockHeader
	SignedHeader2       SignedBeaconBlockHeader
}

// Beacon operations
var ProposerSlashingType =  ContainerType("ProposerSlashing", []FieldDef{
	{"header_1", SignedBeaconBlockHeaderType},
	{"header_2", SignedBeaconBlockHeaderType},
})

func (state *BeaconStateView) ProcessProposerSlashings(ops []ProposerSlashing) error {
	for i := range ops {
		if err := state.ProcessProposerSlashing(&ops[i]); err != nil {
			return err
		}
	}
	return nil
}

func (state *BeaconStateView) ProcessProposerSlashing(ps *ProposerSlashing) error {
	// Verify header slots match
	if ps.SignedHeader1.Message.Slot != ps.SignedHeader2.Message.Slot {
		return errors.New("proposer slashing requires slashing headers to have the same slot")
	}
	// Verify header proposer indices match
	if ps.SignedHeader1.Message.ProposerIndex != ps.SignedHeader2.Message.ProposerIndex {
		return errors.New("proposer slashing headers proposer-indices do not match")
	}
	// Verify header proposer index is valid
	if valid, err := input.IsValidIndex(ps.SignedHeader1.Message.ProposerIndex); err != nil {
		return err
	} else if !valid {
		return errors.New("invalid proposer index")
	}
	// Verify the headers are different
	if ps.SignedHeader1.Message == ps.SignedHeader2.Message {
		return errors.New("proposer slashing requires two different headers")
	}
	currentEpoch, err := input.CurrentEpoch()
	if err != nil {
		return err
	}
	// Verify the proposer is slashable
	if slashable, err := input.IsSlashable(ps.SignedHeader1.Message.ProposerIndex, currentEpoch); err != nil {
		return err
	} else if !slashable {
		return errors.New("proposer slashing requires proposer to be slashable")
	}
	domain, err := state.GetDomain(DOMAIN_BEACON_PROPOSER, ps.SignedHeader1.Message.Slot.ToEpoch())
	if err != nil {
		return err
	}
	pubkey, err := input.Pubkey(ps.SignedHeader1.Message.ProposerIndex)
	if err != nil {
		return err
	}
	// Verify signatures
	if !bls.Verify(
		pubkey,
		ComputeSigningRoot(ps.SignedHeader1.Message.HashTreeRoot(),	domain),
		ps.SignedHeader1.Signature) {
		return errors.New("proposer slashing header 1 has invalid BLS signature")
	}
	if !bls.Verify(
		pubkey,
		ComputeSigningRoot(ps.SignedHeader2.Message.HashTreeRoot(),	domain),
		ps.SignedHeader2.Signature) {
		return errors.New("proposer slashing header 2 has invalid BLS signature")
	}
	input.SlashValidator(ps.SignedHeader1.Message.ProposerIndex, nil)
	return nil
}
