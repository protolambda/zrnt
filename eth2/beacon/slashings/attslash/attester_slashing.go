package attslash

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon/attestations"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zssz"
)

type AttesterSlashingProcessor interface {
	ProcessAttesterSlashings(ops []AttesterSlashing) error
	ProcessAttesterSlashing(attesterSlashing *AttesterSlashing) error
}

type AttestSlashFeature struct {
	Meta interface {
		meta.RegistrySize
		meta.Pubkeys
		meta.Versioning
		meta.Proposers
		meta.Balance
		meta.Slasher
		meta.Exits
	}
}

func (f *AttestSlashFeature) ProcessAttesterSlashings(ops []AttesterSlashing) error {
	for i := range ops {
		if err := f.ProcessAttesterSlashing(&ops[i]); err != nil {
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

var AttesterSlashingType = &ContainerType{
	{"attestation_1", IndexedAttestationType},
	{"attestation_2", IndexedAttestationType},
}

func (f *AttestSlashFeature) ProcessAttesterSlashing(attesterSlashing *AttesterSlashing) error {
	sa1 := &attesterSlashing.Attestation1
	sa2 := &attesterSlashing.Attestation2

	if !IsSlashableAttestationData(&sa1.Data, &sa2.Data) {
		return errors.New("attester slashing has no valid reasoning")
	}

	if err := sa1.Validate(f.Meta); err != nil {
		return errors.New("attestation 1 of attester slashing cannot be verified")
	}
	if err := sa2.Validate(f.Meta); err != nil {
		return errors.New("attestation 2 of attester slashing cannot be verified")
	}

	// keep track of effectiveness
	slashedAny := false

	currentEpoch := f.Meta.CurrentEpoch()

	// run slashings where applicable
	// use ZigZagJoin for efficient intersection: the indicies are already sorted (as validated above)
	ValidatorSet(sa1.AttestingIndices).ZigZagJoin(ValidatorSet(sa2.AttestingIndices), func(i ValidatorIndex) {
		if f.Meta.Validator(i).IsSlashable(currentEpoch) {
			f.Meta.SlashValidator(i, nil)
			slashedAny = true
		}
	}, nil)
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
