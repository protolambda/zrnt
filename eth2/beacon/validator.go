package beacon

import (
	"github.com/protolambda/ztyp/codec"
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

func (v *Validator) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&v.Pubkey, &v.WithdrawalCredentials,  &v.EffectiveBalance, (*BoolView)(&v.Slashed),
		&v.ActivationEligibilityEpoch, &v.ActivationEpoch, &v.ExitEpoch, &v.WithdrawableEpoch)
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

const (
	_validatorPubkey = iota
	_validatorWithdrawalCredentials
	_validatorEffectiveBalance
	_validatorSlashed

	_validatorActivationEligibilityEpoch
	_validatorActivationEpoch
	_validatorExitEpoch
	_validatorWithdrawableEpoch
)

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
	return AsBLSPubkey(v.Get(_validatorPubkey))
}
func (v *ValidatorView) WithdrawalCredentials() (out Root, err error) {
	return AsRoot(v.Get(_validatorWithdrawalCredentials))
}
func (v *ValidatorView) EffectiveBalance() (Gwei, error) {
	return AsGwei(v.Get(_validatorEffectiveBalance))
}
func (v *ValidatorView) SetEffectiveBalance(b Gwei) error {
	return v.Set(_validatorEffectiveBalance, Uint64View(b))
}
func (v *ValidatorView) Slashed() (BoolView, error) {
	return AsBool(v.Get(_validatorSlashed))
}
func (v *ValidatorView) MakeSlashed() error {
	return v.Set(_validatorSlashed, BoolView(true))
}
func (v *ValidatorView) ActivationEligibilityEpoch() (Epoch, error) {
	return AsEpoch(v.Get(_validatorActivationEligibilityEpoch))
}
func (v *ValidatorView) SetActivationEligibilityEpoch(epoch Epoch) error {
	return v.Set(_validatorActivationEligibilityEpoch, Uint64View(epoch))
}
func (v *ValidatorView) ActivationEpoch() (Epoch, error) {
	return AsEpoch(v.Get(_validatorActivationEpoch))
}
func (v *ValidatorView) SetActivationEpoch(epoch Epoch) error {
	return v.Set(_validatorActivationEpoch, Uint64View(epoch))
}
func (v *ValidatorView) ExitEpoch() (Epoch, error) {
	return AsEpoch(v.Get(_validatorExitEpoch))
}
func (v *ValidatorView) SetExitEpoch(ep Epoch) error {
	return v.Set(_validatorExitEpoch, Uint64View(ep))
}
func (v *ValidatorView) WithdrawableEpoch() (Epoch, error) {
	return AsEpoch(v.Get(_validatorWithdrawableEpoch))
}
func (v *ValidatorView) SetWithdrawableEpoch(epoch Epoch) error {
	return v.Set(_validatorWithdrawableEpoch, Uint64View(epoch))
}

func (spec *Spec) IsActive(v *ValidatorView, epoch Epoch) (bool, error) {
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

func (spec *Spec) IsSlashable(v *ValidatorView, epoch Epoch) (bool, error) {
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

func (spec *Spec) IsEligibleForActivationQueue(v *ValidatorView) (bool, error) {
	actEligEpoch, err := v.ActivationEligibilityEpoch()
	if err != nil {
		return false, err
	}
	effBalance, err := v.EffectiveBalance()
	if err != nil {
		return false, err
	}
	return actEligEpoch == FAR_FUTURE_EPOCH && effBalance == spec.MAX_EFFECTIVE_BALANCE, nil
}

func (spec *Spec) IsEligibleForActivation(v *ValidatorView, finalizedEpoch Epoch) (bool, error) {
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
