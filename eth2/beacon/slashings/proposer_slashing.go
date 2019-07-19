package slashings

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon/header"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

type ProposerSlashingReq interface {
	VersioningMeta
	RegistrySizeMeta
	ValidatorMeta
	ProposingMeta
	BalanceMeta
	ExitMeta
}

func (state *SlashingsState) ProcessProposerSlashings(meta AttesterSlashingReq, ops []ProposerSlashing) error {
	for i := range ops {
		if err := state.ProcessProposerSlashing(meta, &ops[i]); err != nil {
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

func (state *SlashingsState) ProcessProposerSlashing(meta ProposerSlashingReq, ps *ProposerSlashing) error {
	if !meta.IsValidIndex(ps.ProposerIndex) {
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
	proposer := meta.Validator(ps.ProposerIndex)
	// Check proposer is slashable
	if !proposer.IsSlashable(meta.Epoch()) {
		return errors.New("proposer slashing requires proposer to be slashable")
	}
	// Signatures are valid
	if !bls.BlsVerify(proposer.Pubkey, ssz.SigningRoot(ps.Header1, BeaconBlockHeaderSSZ), ps.Header1.Signature,
		meta.GetDomain(DOMAIN_BEACON_PROPOSER, ps.Header1.Slot.ToEpoch())) {
		return errors.New("proposer slashing header 1 has invalid BLS signature")
	}
	if !bls.BlsVerify(proposer.Pubkey, ssz.SigningRoot(ps.Header2, BeaconBlockHeaderSSZ), ps.Header2.Signature,
		meta.GetDomain(DOMAIN_BEACON_PROPOSER, ps.Header2.Slot.ToEpoch())) {
		return errors.New("proposer slashing header 2 has invalid BLS signature")
	}
	state.SlashValidator(meta, ps.ProposerIndex, nil)
	return nil
}
