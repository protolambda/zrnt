package propslash

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon/header"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	. "github.com/protolambda/ztyp/props"
	"github.com/protolambda/ztyp/tree"
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

type ProposerSlashing struct { *ContainerView }

func (ps *ProposerSlashing) ProposerIndex() (ValidatorIndex, error) {
	return ValidatorIndexProp(PropReader(ps, 0)).ValidatorIndex()
}

func (ps *ProposerSlashing) SignedHeader1() (*SignedBeaconBlockHeader, error) {
	return SignedBeaconBlockHeaderReadProp(PropReader(ps, 1)).SignedBeaconBlockHeader()
}
func (ps *ProposerSlashing) SignedHeader2() (*SignedBeaconBlockHeader, error) {
	return SignedBeaconBlockHeaderReadProp(PropReader(ps, 2)).SignedBeaconBlockHeader()
}

// Beacon operations
var ProposerSlashingType = &ContainerType{
	{"proposer_index", ValidatorIndexType},
	{"header_1", BeaconBlockHeaderType}, // First proposal
	{"header_2", BeaconBlockHeaderType}, // Second proposal
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
	proposerIndex, err := ps.ProposerIndex()
	if err != nil {
		return err
	}
	if valid, err := f.Meta.IsValidIndex(proposerIndex); err != nil {
		return err
	} else if !valid {
		return errors.New("invalid proposer index")
	}
	signedHeader1, err := ps.SignedHeader1()
	if err != nil {
		return err
	}
	signedHeader2, err := ps.SignedHeader2()
	if err != nil {
		return err
	}
	header1, err := signedHeader1.Message()
	if err != nil {
		return err
	}
	header2, err := signedHeader2.Message()
	if err != nil {
		return err
	}
	slot1, err := header1.Slot()
	if err != nil {
		return err
	}
	slot2, err := header1.Slot()
	if err != nil {
		return err
	}
	// Verify slots match
	if slot1 != slot2 {
		return errors.New("proposer slashing requires slashing headers to have the same slot")
	}
	root1 := header1.ViewRoot(tree.Hash)
	root2 := header2.ViewRoot(tree.Hash)
	// But the headers are different
	if root1 == root2 {
		return errors.New("proposer slashing requires two different headers")
	}
	currentEpoch, err := f.Meta.CurrentEpoch()
	if err != nil {
		return err
	}
	// Check proposer is slashable
	if slashable, err := f.Meta.IsSlashable(proposerIndex, currentEpoch); err != nil {
		return err
	} else if !slashable {
		return errors.New("proposer slashing requires proposer to be slashable")
	}
	pubkey, err := f.Meta.Pubkey(proposerIndex)
	if err != nil {
		return err
	}
	sig1, err := signedHeader1.Signature()
	if err != nil {
		return err
	}
	domain, err := f.Meta.GetDomain(DOMAIN_BEACON_PROPOSER, slot1.ToEpoch())
	if err != nil {
		return err
	}
	// Signatures are valid
	if !bls.BlsVerify(pubkey.Bytes(), root1, sig1.Bytes(), domain) {
		return errors.New("proposer slashing header 1 has invalid BLS signature")
	}
	sig2, err := signedHeader2.Signature()
	if err != nil {
		return err
	}
	if !bls.BlsVerify(pubkey.Bytes(), root2, sig2.Bytes(), domain) {
		return errors.New("proposer slashing header 2 has invalid BLS signature")
	}
	return f.Meta.SlashValidator(proposerIndex, nil)
}
