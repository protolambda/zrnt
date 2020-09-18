package beacon

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

func (spec *Spec) ProcessAttesterSlashings(ctx context.Context, epc *EpochsContext, state *BeaconStateView, ops []AttesterSlashing) error {
	for i := range ops {
		select {
		case <-ctx.Done():
			return TransitionCancelErr
		default: // Don't block.
			break
		}
		if err := spec.ProcessAttesterSlashing(state, epc, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

type AttesterSlashing struct {
	Attestation1 IndexedAttestation
	Attestation2 IndexedAttestation
}

func (p *AttesterSlashing) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(&p.Attestation1, &p.Attestation2)
}

func (c *Phase0Config) BlockAttesterSlashings() ListTypeDef {
	return ListType(c.AttesterSlashing(), c.MAX_ATTESTER_SLASHINGS)
}

func (c *Phase0Config) AttesterSlashing() *ContainerTypeDef {
	return ContainerType("AttesterSlashing", []FieldDef{
		{"attestation_1", c.IndexedAttestation()},
		{"attestation_2", c.IndexedAttestation()},
	})
}

type AttesterSlashings struct {
	Items []AttesterSlashing
	Limit uint64
}

func (li *AttesterSlashings) HashTreeRoot(hFn tree.HashFn) Root {
	length := uint64(len(li.Items))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li.Items[i]
		}
		return nil
	}, length, li.Limit)
}

func (spec *Spec) ProcessAttesterSlashing(state *BeaconStateView, epc *EpochsContext, attesterSlashing *AttesterSlashing) error {
	sa1 := &attesterSlashing.Attestation1
	sa2 := &attesterSlashing.Attestation2

	if !IsSlashableAttestationData(&sa1.Data, &sa2.Data) {
		return errors.New("attester slashing has no valid reasoning")
	}

	if err := spec.ValidateIndexedAttestation(epc, state, sa1); err != nil {
		return errors.New("attestation 1 of attester slashing cannot be verified")
	}
	if err := spec.ValidateIndexedAttestation(epc, state, sa2); err != nil {
		return errors.New("attestation 2 of attester slashing cannot be verified")
	}

	currentEpoch := epc.CurrentEpoch.Epoch

	// keep track of effectiveness
	slashedAny := false
	var errorAny error

	validators, err := state.Validators()
	if err != nil {
		return err
	}
	// run slashings where applicable
	// use ZigZagJoin for efficient intersection: the indicies are already sorted (as validated above)
	ValidatorSet(sa1.AttestingIndices.Indices).ZigZagJoin(ValidatorSet(sa2.AttestingIndices.Indices), func(i ValidatorIndex) {
		if errorAny != nil {
			return
		}
		validator, err := validators.Validator(i)
		if err != nil {
			errorAny = err
			return
		}
		if slashable, err := spec.IsSlashable(validator, currentEpoch); err != nil {
			errorAny = err
		} else if slashable {
			if err := spec.SlashValidator(epc, state, i, nil); err != nil {
				errorAny = err
			} else {
				slashedAny = true
			}
		}
	}, nil)
	if errorAny != nil {
		return fmt.Errorf("error during attester-slashing validators slashable check: %v", errorAny)
	}
	if !slashedAny {
		return errors.New("attester slashing %d is not effective, hence invalid")
	}
	return nil
}

func IsSlashableAttestationData(a *AttestationData, b *AttestationData) bool {
	return IsSurroundVote(a, b) || IsDoubleVote(a, b)
}

// Check if a and b have the same target epoch.
func IsDoubleVote(a *AttestationData, b *AttestationData) bool {
	return *a != *b && a.Target.Epoch == b.Target.Epoch
}

// Check if a surrounds b, i.E. source(a) < source(b) and target(a) > target(b)
func IsSurroundVote(a *AttestationData, b *AttestationData) bool {
	return a.Source.Epoch < b.Source.Epoch && a.Target.Epoch > b.Target.Epoch
}
