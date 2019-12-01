package validator

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/ztyp/props"
	. "github.com/protolambda/ztyp/view"
)

var ValidatorType = &ContainerType{
	{"pubkey", BLSPubkeyType},
	{"withdrawal_credentials", Bytes32Type}, // Commitment to pubkey for withdrawals
	{"effective_balance", GweiType},         // Balance at stake
	{"slashed", BoolType},
	// Status epochs
	{"activation_eligibility_epoch", EpochType}, // When criteria for activation were met
	{"activation_epoch", EpochType},
	{"exit_epoch", EpochType},
	{"withdrawable_epoch", EpochType}, // When validator can withdraw funds
}

type Validator struct {
	*ContainerView
}

func NewValidator() *Validator {
	return &Validator{ContainerView: ValidatorType.New(nil)}
}

func (v *Validator) Pubkey() (BLSPubkey, error) {
	return BLSPubkeyReadProp(PropReader(v, 1)).BLSPubkey()
}
func (v *Validator) WithdrawalCredentials() (out Root, err error) {
	return RootReadProp(PropReader(v, 1)).Root()
}
func (v *Validator) EffectiveBalance() (Gwei, error) {
	return GweiReadProp(PropReader(v, 2)).Gwei()
}
func (v *Validator) Slashed() (bool, error) {
	return BoolReadProp(PropReader(v, 3)).Bool()
}
func (v *Validator) MakeSlashed() error {
	return BoolWriteProp(PropWriter(v, 3)).SetBool(true)
}
func (v *Validator) ActivationEligibilityEpoch() (Epoch, error) {
	return EpochReadProp(PropReader(v, 4)).Epoch()
}
func (v *Validator) ActivationEpoch() (Epoch, error) {
	return EpochReadProp(PropReader(v, 5)).Epoch()
}
func (v *Validator) SetActivationEpoch(epoch Epoch) error {
	return EpochWriteProp(PropWriter(v, 5)).SetEpoch(epoch)
}
func (v *Validator) ExitEpoch() (Epoch, error) {
	return EpochReadProp(PropReader(v, 6)).Epoch()
}
func (v *Validator) WithdrawableEpoch() (Epoch, error) {
	return EpochReadProp(PropReader(v, 7)).Epoch()
}
func (v *Validator) SetWithdrawableEpoch(epoch Epoch) error {
	return EpochWriteProp(PropWriter(v, 7)).SetEpoch(epoch)
}

func (v *Validator) IsActive(epoch Epoch) (bool, error) {
	activationEpoch, err := v.ActivationEpoch()
	if err != nil {
		return false, err
	} else if activationEpoch > epoch {
		return false, nil
	}
	exitEpoch, err := v.ExitEpoch()
	if err != nil {
		return false, err
	} else if epoch >= exitEpoch {
		return false, nil
	}
	return true, nil
}

func (v *Validator) IsSlashable(epoch Epoch) (bool, error) {
	slashed, err := v.Slashed()
	if err != nil {
		return false, err
	} else if slashed {
		return false, nil
	}
	activationEpoch, err := v.ActivationEpoch()
	if err != nil {
		return false, err
	} else if activationEpoch > epoch {
		return false, nil
	}
	withdrawableEpoch, err := v.WithdrawableEpoch()
	if err != nil {
		return false, err
	} else if withdrawableEpoch <= epoch {
		return false, nil
	}
	return true, nil
}
