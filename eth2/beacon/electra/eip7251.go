package electra

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type PendingDeposit struct {
	Pubkey                common.BLSPubkey    `json:"pubkey" yaml:"pubkey"`
	WithdrawalCredentials common.Root         `json:"withdrawal_credentials" yaml:"withdrawal_credentials"`
	Amount                common.Gwei         `json:"amount" yaml:"amount"`
	Signature             common.BLSSignature `json:"signature" yaml:"signature"`
	Slot                  common.Slot         `json:"slot" yaml:"slot"`
}

func (d *PendingDeposit) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&d.Pubkey, &d.WithdrawalCredentials, &d.Amount, &d.Signature, &d.Slot)
}

func (d *PendingDeposit) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&d.Pubkey, &d.WithdrawalCredentials, &d.Amount, &d.Signature, &d.Slot)
}

func (*PendingDeposit) ByteLength(spec *common.Spec) uint64 {
	return PendingDepositType.TypeByteLength()
}

func (*PendingDeposit) FixedLength(spec *common.Spec) uint64 {
	return PendingDepositType.TypeByteLength()
}

func (d *PendingDeposit) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&d.Pubkey, &d.WithdrawalCredentials, &d.Amount, &d.Signature, &d.Slot)
}

func (d *PendingDeposit) View() *PendingDepositView {
	wCred := RootView(d.WithdrawalCredentials)
	c, _ := PendingDepositType.FromFields(
		common.ViewPubkey(&d.Pubkey),
		&wCred,
		Uint64View(d.Amount),
		common.ViewSignature(&d.Signature),
		Uint64View(d.Slot),
	)
	return &PendingDepositView{c}
}

var PendingDepositType = ContainerType("PendingDeposit", []FieldDef{
	{"pubkey", common.BLSPubkeyType},
	{"withdrawal_credentials", common.Bytes32Type}, // Commitment to pubkey for withdrawals
	{"amount", common.GweiType},                    // Balance at stake
	{"signature", common.BLSSignatureType},
	{"slot", common.SlotType},
})

const (
	_pendingDepositPubkey = iota
	_pendingDepositWithdrawalCredentials
	_pendingDepositAmount
	_pendingDepositSignature
	_pendingDepositSlot
)

type PendingDepositView struct {
	*ContainerView
}

func NewPendingDepositView() *PendingDepositView {
	return &PendingDepositView{ContainerView: PendingDepositType.New()}
}

func AsPendingDeposit(v View, err error) (*PendingDepositView, error) {
	c, err := AsContainer(v, err)
	return &PendingDepositView{c}, err
}

func (d *PendingDepositView) Pubkey() (common.BLSPubkey, error) {
	return common.AsBLSPubkey(d.Get(_pendingDepositPubkey))
}
func (d *PendingDepositView) WithdrawalCredentials() (out common.Root, err error) {
	return AsRoot(d.Get(_pendingDepositWithdrawalCredentials))
}
func (d *PendingDepositView) SetWithdrawalCredentials(w common.Root) (err error) {
	wCred := RootView(w)
	return d.Set(_pendingDepositWithdrawalCredentials, &wCred)
}
func (d *PendingDepositView) Amount() (common.Gwei, error) {
	return common.AsGwei(d.Get(_pendingDepositAmount))
}
func (d *PendingDepositView) SetAmount(a common.Gwei) error {
	return d.Set(_pendingDepositAmount, Uint64View(a))
}
func (d *PendingDepositView) Signature() (common.BLSSignature, error) {
	return common.AsBLSSignature(d.Get(_pendingDepositSignature))
}
func (d *PendingDepositView) SetSignature(s common.BLSSignature) error {
	return d.Set(_pendingDepositSignature, common.ViewSignature(&s))
}
func (d *PendingDepositView) Slot() (common.Slot, error) {
	return common.AsSlot(d.Get(_pendingDepositSlot))
}
func (d *PendingDepositView) SetSlot(slot common.Slot) error {
	return d.Set(_pendingDepositSlot, Uint64View(slot))
}

type PendingDeposits []PendingDeposit

func (d *PendingDeposits) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*d)
		*d = append(*d, PendingDeposit{})
		return spec.Wrap(&((*d)[i]))
	}, PendingDepositType.TypeByteLength(), uint64(spec.MAX_PENDING_DEPOSITS))
}

func (d PendingDeposits) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return spec.Wrap(&d[i])
	}, PendingDepositType.TypeByteLength(), uint64(len(d)))
}

func (d PendingDeposits) ByteLength(spec *common.Spec) uint64 {
	return PendingDepositType.TypeByteLength() * uint64(len(d))
}

func (d *PendingDeposits) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li PendingDeposits) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(&li[i])
		}
		return nil
	}, length, uint64(spec.MAX_PENDING_DEPOSITS))
}

func PendingDepositsType(spec *common.Spec) *ComplexListTypeDef {
	return ComplexListType(PendingDepositType, uint64(spec.MAX_PENDING_DEPOSITS))
}

type PendingDepositsView struct{ *ComplexListView }

var _ PendingDepositsList = (*PendingDepositsView)(nil)

func AsPendingDeposits(v View, err error) (*PendingDepositsView, error) {
	c, err := AsComplexList(v, err)
	return &PendingDepositsView{c}, err
}

func (d *PendingDepositsView) Append(deposit PendingDeposit) error {
	v := deposit.View()
	return d.ComplexListView.Append(v)
}

type PendingPartialWithdrawal struct {
	Index             common.ValidatorIndex `json:"index" yaml:"index"`
	Amount            common.Gwei           `json:"amount" yaml:"amount"`
	WithdrawableEpoch common.Epoch          `json:"withdrawable_epoch" yaml:"withdrawable_epoch"`
}

func (w *PendingPartialWithdrawal) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&w.Index, &w.Amount, &w.WithdrawableEpoch)
}

func (w *PendingPartialWithdrawal) Serialize(spec *common.Spec, ew *codec.EncodingWriter) error {
	return ew.Container(&w.Index, &w.Amount, &w.WithdrawableEpoch)
}

func (*PendingPartialWithdrawal) ByteLength(spec *common.Spec) uint64 {
	return PendingPartialWithdrawalType.TypeByteLength()
}

func (*PendingPartialWithdrawal) FixedLength(spec *common.Spec) uint64 {
	return PendingPartialWithdrawalType.TypeByteLength()
}

func (w *PendingPartialWithdrawal) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&w.Index, &w.Amount, &w.WithdrawableEpoch)
}

func (w *PendingPartialWithdrawal) View() *PendingPartialWithdrawalView {
	c, _ := PendingPartialWithdrawalType.FromFields(
		Uint64View(w.Index),
		Uint64View(w.Amount),
		Uint64View(w.WithdrawableEpoch),
	)
	return &PendingPartialWithdrawalView{c}
}

var PendingPartialWithdrawalType = ContainerType("PendingPartialWithdrawal", []FieldDef{
	{"index", common.ValidatorIndexType},
	{"amount", common.GweiType},
	{"withdrawable_epoch", common.EpochType},
})

const (
	_PendingPartialWithdrawalIndex = iota
	_PendingPartialWithdrawalAmount
	_PendingPartialWithdrawalWithdrawableEpoch
)

type PendingPartialWithdrawalView struct {
	*ContainerView
}

func NewPendingPartialWithdrawalView() *PendingPartialWithdrawalView {
	return &PendingPartialWithdrawalView{ContainerView: PendingPartialWithdrawalType.New()}
}

func AsPendingPartialWithdrawal(v View, err error) (*PendingPartialWithdrawalView, error) {
	c, err := AsContainer(v, err)
	return &PendingPartialWithdrawalView{c}, err
}

func (w *PendingPartialWithdrawalView) Index() (common.ValidatorIndex, error) {
	return common.AsValidatorIndex(w.Get(_PendingPartialWithdrawalIndex))
}
func (w *PendingPartialWithdrawalView) SetIndex(i common.ValidatorIndex) (err error) {
	return w.Set(_PendingPartialWithdrawalIndex, Uint64View(i))
}
func (w *PendingPartialWithdrawalView) Amount() (out common.Gwei, err error) {
	return common.AsGwei(w.Get(_PendingPartialWithdrawalAmount))
}
func (w *PendingPartialWithdrawalView) SetAmount(a common.Gwei) (err error) {
	return w.Set(_PendingPartialWithdrawalAmount, Uint64View(a))
}
func (w *PendingPartialWithdrawalView) WithdrawableEpoch() (common.Epoch, error) {
	return common.AsEpoch(w.Get(_PendingPartialWithdrawalWithdrawableEpoch))
}
func (w *PendingPartialWithdrawalView) SetWithdrawableEpoch(epoch common.Epoch) error {
	return w.Set(_PendingPartialWithdrawalWithdrawableEpoch, Uint64View(epoch))
}

type PendingPartialWithdrawals []PendingPartialWithdrawal

func (w *PendingPartialWithdrawals) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*w)
		*w = append(*w, PendingPartialWithdrawal{})
		return spec.Wrap(&((*w)[i]))
	}, PendingPartialWithdrawalType.TypeByteLength(), uint64(spec.MAX_PENDING_PARTIAL_WITHDRAWALS))
}

func (w PendingPartialWithdrawals) Serialize(spec *common.Spec, ew *codec.EncodingWriter) error {
	return ew.List(func(i uint64) codec.Serializable {
		return spec.Wrap(&w[i])
	}, PendingPartialWithdrawalType.TypeByteLength(), uint64(len(w)))
}

func (w PendingPartialWithdrawals) ByteLength(spec *common.Spec) uint64 {
	return PendingPartialWithdrawalType.TypeByteLength() * uint64(len(w))
}

func (w *PendingPartialWithdrawals) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li PendingPartialWithdrawals) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(&li[i])
		}
		return nil
	}, length, uint64(spec.MAX_PENDING_PARTIAL_WITHDRAWALS))
}

func PendingPartialWithdrawalsType(spec *common.Spec) *ComplexListTypeDef {
	return ComplexListType(PendingPartialWithdrawalType, uint64(spec.MAX_PENDING_PARTIAL_WITHDRAWALS))
}

type PendingPartialWithdrawalsView struct{ *ComplexListView }

var _ PendingPartialWithdrawalsList = (*PendingPartialWithdrawalsView)(nil)

func AsPendingPartialWithdrawals(v View, err error) (*PendingPartialWithdrawalsView, error) {
	c, err := AsComplexList(v, err)
	return &PendingPartialWithdrawalsView{c}, err
}

func (w *PendingPartialWithdrawalsView) Append(withdrawal PendingPartialWithdrawal) error {
	v := withdrawal.View()
	return w.ComplexListView.Append(v)
}
