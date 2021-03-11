package phase0

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/view"
)

type AttesterFlag uint8

func (flags AttesterFlag) HasMarkers(markers AttesterFlag) bool {
	return flags&markers == markers
}

const (
	PrevSourceAttester AttesterFlag = 1 << iota
	PrevTargetAttester
	PrevHeadAttester

	CurrSourceAttester
	CurrTargetAttester
	CurrHeadAttester

	UnslashedAttester
	EligibleAttester
)

type FlatValidator struct {
	EffectiveBalance           common.Gwei
	Slashed                    bool
	ActivationEligibilityEpoch common.Epoch
	ActivationEpoch            common.Epoch
	ExitEpoch                  common.Epoch
	WithdrawableEpoch          common.Epoch
}

func (v *FlatValidator) IsActive(epoch common.Epoch) bool {
	return v.ActivationEpoch <= epoch && epoch < v.ExitEpoch
}

func ToFlatValidator(v *ValidatorView) (*FlatValidator, error) {
	/*
	   pubkey: BLSPubkey
	   withdrawal_credentials: Bytes32
	   effective_balance: Gwei  # Balance at stake
	   slashed: boolean
	   activation_eligibility_epoch: Epoch
	   activation_epoch: Epoch
	   exit_epoch: Epoch
	   withdrawable_epoch: Epoch
	*/
	fields, err := v.FieldValues()
	if err != nil {
		return nil, err
	}
	effBal, err := common.AsGwei(fields[2], err)
	slashed, err := view.AsBool(fields[3], err)
	acitvEligEp, err := common.AsEpoch(fields[4], err)
	activEp, err := common.AsEpoch(fields[5], err)
	exitEp, err := common.AsEpoch(fields[6], err)
	withEp, err := common.AsEpoch(fields[7], err)
	if err != nil {
		return nil, err
	}
	return &FlatValidator{
		EffectiveBalance:           effBal,
		Slashed:                    bool(slashed),
		ActivationEligibilityEpoch: acitvEligEp,
		ActivationEpoch:            activEp,
		ExitEpoch:                  exitEp,
		WithdrawableEpoch:          withEp,
	}, nil
}

type AttesterStatus struct {
	// The delay of inclusion of the latest attestation by the attester.
	// No delay (i.e. 0) by default
	InclusionDelay common.Slot
	// The validator index of the proposer of the attested beacon block.
	// Only valid if the validator has an attesting flag set.
	AttestedProposer common.ValidatorIndex
	// A bitfield of markers describing the recent actions of the validator
	Flags     AttesterFlag
	Validator *FlatValidator
	// If the validator is active
	Active bool
}
