package beacon

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

func (state *BeaconStateView) ProcessAttesterSlashings(ctx context.Context, epc *EpochsContext, ops []AttesterSlashing) error {
	for i := range ops {
		select {
		case <-ctx.Done():
			return TransitionCancelErr
		default: // Don't block.
			break
		}
		if err := state.ProcessAttesterSlashing(epc, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

var AttesterSlashingSSZ = zssz.GetSSZ((*AttesterSlashing)(nil))

type AttesterSlashing struct {
	Attestation1 IndexedAttestation
	Attestation2 IndexedAttestation
}

var AttesterSlashingType = ContainerType("AttesterSlashing", []FieldDef{
	{"attestation_1", IndexedAttestationType},
	{"attestation_2", IndexedAttestationType},
})

var AttesterSlashingsType = ListType(AttesterSlashingType, MAX_ATTESTER_SLASHINGS)

type AttesterSlashings []AttesterSlashing

func (*AttesterSlashings) Limit() uint64 {
	return MAX_ATTESTER_SLASHINGS
}

func (state *BeaconStateView) ProcessAttesterSlashing(epc *EpochsContext, attesterSlashing *AttesterSlashing) error {
	sa1 := &attesterSlashing.Attestation1
	sa2 := &attesterSlashing.Attestation2

	if !IsSlashableAttestationData(&sa1.Data, &sa2.Data) {
		return errors.New("attester slashing has no valid reasoning")
	}

	if err := state.ValidateIndexedAttestation(epc, sa1); err != nil {
		return errors.New("attestation 1 of attester slashing cannot be verified")
	}
	if err := state.ValidateIndexedAttestation(epc, sa2); err != nil {
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
	ValidatorSet(sa1.AttestingIndices).ZigZagJoin(ValidatorSet(sa2.AttestingIndices), func(i ValidatorIndex) {
		if errorAny != nil {
			return
		}
		validator, err := validators.Validator(i)
		if err != nil {
			errorAny = err
			return
		}
		if slashable, err := validator.IsSlashable(currentEpoch); err != nil {
			errorAny = err
		} else if slashable {
			if err := state.SlashValidator(epc, i, nil); err != nil {
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
