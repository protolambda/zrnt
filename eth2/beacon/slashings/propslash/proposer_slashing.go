package propslash

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon/header"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type ProposerSlashingProcessor interface {
	ProcessProposerSlashings(ops []ProposerSlashing) error
	ProcessProposerSlashing(ps *ProposerSlashing) error
}

type PropSlashFeature struct {
	Meta interface {
		meta.Versioning
		meta.RegistrySize
		meta.Validators
		meta.Proposers
		meta.Balance
		meta.Exits
		meta.Slasher
	}
}

var ProposerSlashingSSZ = zssz.GetSSZ((*ProposerSlashing)(nil))

type ProposerSlashing struct {
	ProposerIndex ValidatorIndex
	SignedHeader1 SignedBeaconBlockHeader // First proposal
	SignedHeader2 SignedBeaconBlockHeader // Second proposal
}

func (f *PropSlashFeature) ProcessProposerSlashings(ops []ProposerSlashing) error {
	for i := range ops {
		if err := f.ProcessProposerSlashing(&ops[i]); err != nil {
			return err
		}
	}
	return nil
}

func (f *PropSlashFeature) ProcessProposerSlashing(ps *ProposerSlashing) error {
	if !f.Meta.IsValidIndex(ps.ProposerIndex) {
		return errors.New("invalid proposer index")
	}
	// Verify slots match
	if ps.SignedHeader1.Message.Slot != ps.SignedHeader2.Message.Slot {
		return errors.New("proposer slashing requires slashing headers to have the same slot")
	}
	// But the headers are different
	if ps.SignedHeader1.Message == ps.SignedHeader2.Message {
		return errors.New("proposer slashing requires two different headers")
	}
	proposer := f.Meta.Validator(ps.ProposerIndex)
	// Check proposer is slashable
	if !proposer.IsSlashable(f.Meta.CurrentEpoch()) {
		return errors.New("proposer slashing requires proposer to be slashable")
	}
	// Signatures are valid
	if !bls.BlsVerify(proposer.Pubkey, ssz.HashTreeRoot(ps.SignedHeader1.Message, BeaconBlockHeaderSSZ),
		ps.SignedHeader1.Signature,
		f.Meta.GetDomain(DOMAIN_BEACON_PROPOSER, ps.SignedHeader1.Message.Slot.ToEpoch())) {
		return errors.New("proposer slashing header 1 has invalid BLS signature")
	}
	if !bls.BlsVerify(proposer.Pubkey,
		ssz.HashTreeRoot(ps.SignedHeader2.Message, BeaconBlockHeaderSSZ),
		ps.SignedHeader2.Signature,
		f.Meta.GetDomain(DOMAIN_BEACON_PROPOSER, ps.SignedHeader2.Message.Slot.ToEpoch())) {
		return errors.New("proposer slashing header 2 has invalid BLS signature")
	}
	f.Meta.SlashValidator(ps.ProposerIndex, nil)
	return nil
}
