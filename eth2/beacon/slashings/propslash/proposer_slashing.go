package propslash

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon/header"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
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

type ProposerSlashing struct {
	ProposerIndex ValidatorIndex
	Header1       BeaconBlockHeader // First proposal
	Header2       BeaconBlockHeader // Second proposal
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
	// Verify that the epoch is the same
	if ps.Header1.Slot.ToEpoch() != ps.Header2.Slot.ToEpoch() {
		return errors.New("proposer slashing requires slashing headers to be in same epoch")
	}
	// But the headers are different
	if ps.Header1 == ps.Header2 {
		return errors.New("proposer slashing requires two different headers")
	}
	proposer := f.Meta.Validator(ps.ProposerIndex)
	// Check proposer is slashable
	if !proposer.IsSlashable(f.Meta.CurrentEpoch()) {
		return errors.New("proposer slashing requires proposer to be slashable")
	}
	// Signatures are valid
	if !bls.BlsVerify(proposer.Pubkey, ssz.SigningRoot(ps.Header1, BeaconBlockHeaderSSZ), ps.Header1.Signature,
		f.Meta.GetDomain(DOMAIN_BEACON_PROPOSER, ps.Header1.Slot.ToEpoch())) {
		return errors.New("proposer slashing header 1 has invalid BLS signature")
	}
	if !bls.BlsVerify(proposer.Pubkey, ssz.SigningRoot(ps.Header2, BeaconBlockHeaderSSZ), ps.Header2.Signature,
		f.Meta.GetDomain(DOMAIN_BEACON_PROPOSER, ps.Header2.Slot.ToEpoch())) {
		return errors.New("proposer slashing header 2 has invalid BLS signature")
	}
	f.Meta.SlashValidator(ps.ProposerIndex, nil)
	return nil
}
