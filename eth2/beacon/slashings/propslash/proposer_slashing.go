package propslash

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon/header"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

type ProposerSlashingProcessor interface {
	ProcessProposerSlashings(ops []ProposerSlashing) error
	ProcessProposerSlashing(ps *ProposerSlashing) error
}

type PropSlashFeature struct {
	Meta interface {
		meta.Pubkeys
		meta.SigDomain
		meta.SlashableCheck
		meta.Versioning
		meta.RegistrySize
		meta.Proposers
		meta.Balance
		meta.Exits
		meta.Slasher
	}
}

var ProposerSlashingSSZ = zssz.GetSSZ((*ProposerSlashing)(nil))

type ProposerSlashing struct {
	ProposerIndex ValidatorIndex
	SignedHeader1       SignedBeaconBlockHeader
	SignedHeader2       SignedBeaconBlockHeader
}

// Beacon operations
var ProposerSlashingType = &ContainerType{
	{"proposer_index", ValidatorIndexType},
	{"header_1", SignedBeaconBlockHeaderType},
	{"header_2", SignedBeaconBlockHeaderType},
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
	if valid, err := f.Meta.IsValidIndex(ps.ProposerIndex); err != nil {
		return err
	} else if !valid {
		return errors.New("invalid proposer index")
	}
	// Verify slots match
	if ps.SignedHeader1.Message.Slot != ps.SignedHeader2.Message.Slot {
		return errors.New("proposer slashing requires slashing headers to have the same slot")
	}
	root1 := ps.SignedHeader1.Message.HashTreeRoot()
	root2 := ps.SignedHeader2.Message.HashTreeRoot()
	// But the headers are different
	if root1 == root2 {
		return errors.New("proposer slashing requires two different headers")
	}
	currentEpoch, err := f.Meta.CurrentEpoch()
	if err != nil {
		return err
	}
	// Check proposer is slashable
	if slashable, err := f.Meta.IsSlashable(ps.ProposerIndex, currentEpoch); err != nil {
		return err
	} else if !slashable {
		return errors.New("proposer slashing requires proposer to be slashable")
	}
	pubkey, err := f.Meta.Pubkey(ps.ProposerIndex)
	if err != nil {
		return err
	}
	domain, err := f.Meta.GetDomain(DOMAIN_BEACON_PROPOSER, ps.SignedHeader1.Message.Slot.ToEpoch())
	if err != nil {
		return err
	}
	// Signatures are valid
	if !bls.BlsVerify(pubkey, root1, ps.SignedHeader1.Signature, domain) {
		return errors.New("proposer slashing header 1 has invalid BLS signature")
	}
	if !bls.BlsVerify(pubkey, root2, ps.SignedHeader2.Signature, domain) {
		return errors.New("proposer slashing header 2 has invalid BLS signature")
	}
	return f.Meta.SlashValidator(ps.ProposerIndex, nil)
}
