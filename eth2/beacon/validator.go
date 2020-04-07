package beacon

import (
	. "github.com/protolambda/ztyp/view"
)

type Validator struct {
	Pubkey                BLSPubkey
	WithdrawalCredentials Root // Commitment to pubkey for withdrawals
	EffectiveBalance      Gwei // Balance at stake
	Slashed               bool

	// Status epochs
	ActivationEligibilityEpoch Epoch // When criteria for activation were met
	ActivationEpoch            Epoch
	ExitEpoch                  Epoch
	WithdrawableEpoch          Epoch // When validator can withdraw funds
}

func (v *Validator) View() *ValidatorView {
	wCred := RootView(v.WithdrawalCredentials)
	c, _ := ValidatorType.FromFields(
		ViewPubkey(&v.Pubkey),
		&wCred,
		Uint64View(v.EffectiveBalance),
		BoolView(v.Slashed),
		Uint64View(v.ActivationEligibilityEpoch),
		Uint64View(v.ActivationEpoch),
		Uint64View(v.ExitEpoch),
		Uint64View(v.WithdrawableEpoch),
	)
	return &ValidatorView{c}
}

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
	return AsBLSPubkey(v.Get(0))
}
func (v *ValidatorView) WithdrawalCredentials() (out Root, err error) {
	return AsRoot(v.Get(1))
}
func (v *ValidatorView) EffectiveBalance() (Gwei, error) {
	return AsGwei(v.Get(2))
}
func (v *ValidatorView) SetEffectiveBalance(b Gwei) error {
	return v.Set(2, Uint64View(b))
}
func (v *ValidatorView) Slashed() (BoolView, error) {
	return AsBool(v.Get(3))
}
func (v *ValidatorView) MakeSlashed() error {
	return v.Set(3, BoolView(true))
}
func (v *ValidatorView) ActivationEligibilityEpoch() (Epoch, error) {
	return AsEpoch(v.Get(4))
}
func (v *ValidatorView) SetActivationEligibilityEpoch(epoch Epoch) error {
	return v.Set(4, Uint64View(epoch))
}
func (v *ValidatorView) ActivationEpoch() (Epoch, error) {
	return AsEpoch(v.Get(5))
}
func (v *ValidatorView) SetActivationEpoch(epoch Epoch) error {
	return v.Set(5, Uint64View(epoch))
}
func (v *ValidatorView) ExitEpoch() (Epoch, error) {
	return AsEpoch(v.Get(6))
}
func (v *ValidatorView) SetExitEpoch(ep Epoch) error {
	return v.Set(6, Uint64View(ep))
}
func (v *ValidatorView) WithdrawableEpoch() (Epoch, error) {
	return AsEpoch(v.Get(7))
}
func (v *ValidatorView) SetWithdrawableEpoch(epoch Epoch) error {
	return v.Set(6, Uint64View(epoch))
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
