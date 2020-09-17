package beacon

import (
	"context"
	"errors"

	"github.com/protolambda/zrnt/eth2/util/bls"
	. "github.com/protolambda/ztyp/view"
)

type ProposerSlashing struct {
	SignedHeader1 SignedBeaconBlockHeader
	SignedHeader2 SignedBeaconBlockHeader
}

func (c *Phase0Config) ProposerSlashing() *ContainerTypeDef {
	return ContainerType("ProposerSlashing", []FieldDef{
		{"header_1", c.SignedBeaconBlockHeader()},
		{"header_2", c.SignedBeaconBlockHeader()},
	})
}

func (c *Phase0Config) BlockProposerSlashings() ListTypeDef {
	return ListType(c.ProposerSlashing(), c.MAX_PROPOSER_SLASHINGS)
}

type ProposerSlashings []ProposerSlashing

func (spec *Spec) ProcessProposerSlashings(ctx context.Context, epc *EpochsContext, state *BeaconStateView, ops []ProposerSlashing) error {
	for i := range ops {
		select {
		case <-ctx.Done():
			return TransitionCancelErr
		default: // Don't block.
			break
		}
		if err := spec.ProcessProposerSlashing(epc, state, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

func (spec *Spec) ProcessProposerSlashing(epc *EpochsContext, state *BeaconStateView, ps *ProposerSlashing) error {
	// Verify header slots match
	if ps.SignedHeader1.Message.Slot != ps.SignedHeader2.Message.Slot {
		return errors.New("proposer slashing requires slashing headers to have the same slot")
	}
	// Verify header proposer indices match
	if ps.SignedHeader1.Message.ProposerIndex != ps.SignedHeader2.Message.ProposerIndex {
		return errors.New("proposer slashing headers proposer-indices do not match")
	}
	proposerIndex := ps.SignedHeader1.Message.ProposerIndex
	// Verify header proposer index is valid
	if valid, err := state.IsValidIndex(proposerIndex); err != nil {
		return err
	} else if !valid {
		return errors.New("invalid proposer index")
	}
	// Verify the headers are different
	if ps.SignedHeader1.Message == ps.SignedHeader2.Message {
		return errors.New("proposer slashing requires two different headers")
	}
	currentEpoch := epc.CurrentEpoch.Epoch
	// Verify the proposer is slashable
	validators, err := state.Validators()
	if err != nil {
		return err
	}
	validator, err := validators.Validator(proposerIndex)
	if err != nil {
		return err
	}
	if slashable, err := spec.IsSlashable(validator, currentEpoch); err != nil {
		return err
	} else if !slashable {
		return errors.New("proposer slashing requires proposer to be slashable")
	}
	domain, err := state.GetDomain(spec.DOMAIN_BEACON_PROPOSER, spec.SlotToEpoch(ps.SignedHeader1.Message.Slot))
	if err != nil {
		return err
	}
	pubkey, ok := epc.PubkeyCache.Pubkey(proposerIndex)
	if !ok {
		return errors.New("could not find pubkey of proposer")
	}
	// Verify signatures
	if !bls.Verify(
		pubkey,
		ComputeSigningRoot(ps.SignedHeader1.Message.HashTreeRoot(), domain),
		ps.SignedHeader1.Signature) {
		return errors.New("proposer slashing header 1 has invalid BLS signature")
	}
	if !bls.Verify(
		pubkey,
		ComputeSigningRoot(ps.SignedHeader2.Message.HashTreeRoot(), domain),
		ps.SignedHeader2.Signature) {
		return errors.New("proposer slashing header 2 has invalid BLS signature")
	}
	return spec.SlashValidator(epc, state, proposerIndex, nil)
}
