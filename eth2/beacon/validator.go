package beacon

import (
	. "github.com/protolambda/ztyp/props"
	. "github.com/protolambda/ztyp/view"
)

var ValidatorType = ContainerType("Validator", []FieldDef{
	{"pubkey", BLSPubkeyType},
	{"withdrawal_credentials", Bytes32Type}, // Commitment to pubkey for withdrawals
	{"effective_balance", GweiType},         // Balance at stake
	{"slashed", BoolType},
	// Status epochs
	{"activation_eligibility_epoch", EpochType}, // When criteria for activation were met
	{"activation_epoch", EpochType},
	{"exit_epoch", EpochType},
	{"withdrawable_epoch", EpochType}, // When validator can withdraw funds
})

type ValidatorView struct {
	*ContainerView
}

func NewValidatorView() *ValidatorView {
	return &ValidatorView{ContainerView: ValidatorType.New()}
}

func AsValidator(v View, err error) (*ValidatorView, error) {
	c, err := AsContainer(v, err)
	return &ValidatorView{c}, err
}

func (v *ValidatorView) Pubkey() (BLSPubkey, error) {
	return BLSPubkeyProp(PropReader(v, 1)).BLSPubkey()
}
func (v *ValidatorView) WithdrawalCredentials() (out Root, err error) {
	return RootReadProp(PropReader(v, 1)).Root()
}
func (v *ValidatorView) EffectiveBalance() (Gwei, error) {
	return GweiReadProp(PropReader(v, 2)).Gwei()
}
func (v *ValidatorView) Slashed() (bool, error) {
	return BoolReadProp(PropReader(v, 3)).Bool()
}
func (v *ValidatorView) MakeSlashed() error {
	return BoolWriteProp(PropWriter(v, 3)).SetBool(true)
}
func (v *ValidatorView) ActivationEligibilityEpoch() (Epoch, error) {
	return EpochReadProp(PropReader(v, 4)).Epoch()
}
func (v *ValidatorView) ActivationEpoch() (Epoch, error) {
	return EpochReadProp(PropReader(v, 5)).Epoch()
}
func (v *ValidatorView) SetActivationEpoch(epoch Epoch) error {
	return EpochWriteProp(PropWriter(v, 5)).SetEpoch(epoch)
}
func (v *ValidatorView) ExitEpoch() (Epoch, error) {
	return EpochReadProp(PropReader(v, 6)).Epoch()
}
func (v *ValidatorView) WithdrawableEpoch() (Epoch, error) {
	return EpochReadProp(PropReader(v, 7)).Epoch()
}
func (v *ValidatorView) SetWithdrawableEpoch(epoch Epoch) error {
	return EpochWriteProp(PropWriter(v, 7)).SetEpoch(epoch)
}

func (v *ValidatorView) IsActive(epoch Epoch) (bool, error) {
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

func (v *ValidatorView) IsSlashable(epoch Epoch) (bool, error) {
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

func (v *ValidatorView) IsEligibleForActivationQueue() (bool, error) {
	actEligEpoch, err := v.ActivationEligibilityEpoch()
	if err != nil {
		return false, err
	}
	effBalance, err := v.EffectiveBalance()
	if err != nil {
		return false, err
	}
	return actEligEpoch == FAR_FUTURE_EPOCH && effBalance == MAX_EFFECTIVE_BALANCE, nil
}

func (v *ValidatorView) IsEligibleForActivation(finalizedEpoch Epoch) (bool, error) {
	actEligEpoch, err := v.ActivationEligibilityEpoch()
	if err != nil {
		return false, err
	}
	actEpoch, err := v.ActivationEpoch()
	if err != nil {
		return false, err
	}
	return actEligEpoch <= finalizedEpoch && actEpoch == FAR_FUTURE_EPOCH, nil
}
