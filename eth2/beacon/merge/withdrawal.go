package merge

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Withdrawal struct {
	ValidatorIndex        common.ValidatorIndex `json:"validator_index" yaml:"validator_index"`
	WithdrawalCredentials Bytes32               `json:"withdrawal_credentials" yaml:"withdrawal_credentials"`
	WithdrawnEpoch        common.Epoch          `json:"withdrawn_epoch" yaml:"withdrawn_epoch"`
	Amount                common.Gwei           `json:"amount" yaml:"amount"`
}

func (w *Withdrawal) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&w.ValidatorIndex, &w.WithdrawalCredentials,
		&w.WithdrawnEpoch, &w.Amount)
}

func (w *Withdrawal) Serialize(dw *codec.EncodingWriter) error {
	return dw.FixedLenContainer(&w.ValidatorIndex, &w.WithdrawalCredentials,
		&w.WithdrawnEpoch, &w.Amount)
}

func (a *Withdrawal) ByteLength() uint64 {
	return WithdrawalType.TypeByteLength()
}

func (*Withdrawal) FixedLength() uint64 {
	return WithdrawalType.TypeByteLength()
}

func (w *Withdrawal) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(w.ValidatorIndex, w.WithdrawalCredentials, w.WithdrawnEpoch,
		w.Amount)
}

func (w *Withdrawal) View() *WithdrawalView {
	wCred := RootView(w.WithdrawalCredentials)
	c, _ := WithdrawalType.FromFields(
		Uint64View(w.ValidatorIndex),
		&wCred,
		Uint64View(w.WithdrawnEpoch),
		Uint64View(w.Amount),
	)
	return &WithdrawalView{c}
}

var WithdrawalType = ContainerType("Withdrawal", []FieldDef{
	{"validator_index", common.ValidatorIndexType},
	{"withdrawal_credentials", common.Bytes32Type}, // Commitment to pubkey for withdrawals
	{"withdrawn_epoch", common.EpochType},
	{"amount", common.GweiType},
})

const (
	_withdrawalValidatorIndex = iota
	_withdrawalWithdrawalCredentials
	_withdrawalWithdrawnEpoch
	_withdrawalAmount
)

type WithdrawalView struct {
	*ContainerView
}

func NewWithdrawalView() *WithdrawalView {
	return &WithdrawalView{ContainerView: WithdrawalType.New()}
}

func AsWithdrawal(v View, err error) (*Withdrawal, error) {
	c, err := AsContainer(v, err)
	if err != nil {
		return nil, err
	}
	validatorIndex, err := common.AsValidatorIndex(c.Get(_withdrawalValidatorIndex))
	if err != nil {
		return nil, err
	}
	withdrawalCredentials, err := AsRoot(c.Get(_withdrawalWithdrawalCredentials))
	if err != nil {
		return nil, err
	}
	withdrawnEpoch, err := common.AsEpoch(c.Get(_withdrawalWithdrawnEpoch))
	if err != nil {
		return nil, err
	}
	amount, err := common.AsGwei(c.Get(_withdrawalAmount))
	if err != nil {
		return nil, err
	}
	w := Withdrawal{
		ValidatorIndex:        validatorIndex,
		WithdrawalCredentials: withdrawalCredentials,
		WithdrawnEpoch:        withdrawnEpoch,
		Amount:                amount,
	}
	return &w, nil
}

func (w *WithdrawalView) ValidatorIndex() (common.ValidatorIndex, error) {
	return common.AsValidatorIndex(w.Get(_withdrawalValidatorIndex))
}
func (w *WithdrawalView) WithdrawalCredentials() (out common.Root, err error) {
	return AsRoot(w.Get(_withdrawalWithdrawalCredentials))
}
func (w *WithdrawalView) WithdrawnEpoch() (common.Epoch, error) {
	return common.AsEpoch(w.Get(_withdrawalWithdrawnEpoch))
}
func (w *WithdrawalView) Amount() (common.Gwei, error) {
	return common.AsGwei(w.Get(_withdrawalAmount))
}
