package electra

import (
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"

	"github.com/protolambda/zrnt/eth2/beacon/common"
)

func ExecutionRequestsType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("ExecutionRequests", []FieldDef{
		{"deposit_requests", DepositRequestsType(spec)},
		{"withdrawal_requests", WithdrawalRequestsType(spec)},
	})
}

type ExecutionRequestView struct {
	*ContainerView
}

func AsExecutionRequest(v View, err error) (*ExecutionRequestView, error) {
	c, err := AsContainer(v, err)
	return &ExecutionRequestView{c}, err
}

type ExecutionRequests struct {
	Deposits    DepositRequests    `json:"deposits" yaml:"deposits"`
	Withdrawals WithdrawalRequests `json:"withdrawals" yaml:"withdrawals"`
}

func (r *ExecutionRequests) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&r.Deposits), spec.Wrap(&r.Withdrawals))
}

func (r *ExecutionRequests) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&r.Deposits), spec.Wrap(&r.Withdrawals))
}

func (r *ExecutionRequests) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&r.Deposits), spec.Wrap(&r.Withdrawals))
}

func (r *ExecutionRequests) FixedLength(*common.Spec) uint64 {
	// transactions list is not fixed length, so the whole thing is not fixed length.
	return 0
}

func (r *ExecutionRequests) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&r.Deposits), spec.Wrap(&r.Withdrawals))
}

func (r *ExecutionRequests) GetDeposits() DepositRequests {
	return r.Deposits
}

func (r *ExecutionRequests) GetWitdrawals() WithdrawalRequests {
	return r.Withdrawals
}

// //////////////////////////////////////////////////////////////
type DepositRequest struct {
	Pubkey                common.BLSPubkey    `json:"pubkey" yaml:"pubkey"`
	WithdrawalCredentials common.Root         `json:"withdrawal_credentials" yaml:"withdrawal_credentials"`
	Amount                common.Gwei         `json:"amount" yaml:"amount"`
	Signature             common.BLSSignature `json:"signature" yaml:"signature"`
	Index                 common.Number       `json:"index" yaml:"index"`
}

func (r *DepositRequest) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&r.Pubkey, &r.WithdrawalCredentials, &r.Amount, &r.Signature, &r.Index)
}

func (r *DepositRequest) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&r.Pubkey, &r.WithdrawalCredentials, &r.Amount, &r.Signature, &r.Index)
}

func (*DepositRequest) ByteLength(spec *common.Spec) uint64 {
	return DepositRequestType.TypeByteLength()
}

func (*DepositRequest) FixedLength(spec *common.Spec) uint64 {
	return DepositRequestType.TypeByteLength()
}

func (r *DepositRequest) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&r.Pubkey, &r.WithdrawalCredentials, &r.Amount, &r.Signature, &r.Index)
}

func (r *DepositRequest) View() *DepositRequestView {
	wCred := RootView(r.WithdrawalCredentials)
	c, _ := DepositRequestType.FromFields(
		common.ViewPubkey(&r.Pubkey),
		&wCred,
		Uint64View(r.Amount),
		common.ViewSignature(&r.Signature),
		Uint64View(r.Index),
	)
	return &DepositRequestView{c}
}

var DepositRequestType = ContainerType("DepositRequest", []FieldDef{
	{"pubkey", common.BLSPubkeyType},
	{"withdrawal_credentials", common.Bytes32Type},
	{"amount", common.GweiType},
	{"signature", common.BLSSignatureType},
	{"index", common.NumberType},
})

const (
	_depositRequestPubkey = iota
	_depositRequestWithdrawalCredentials
	_depositRequestAmount
	_depositRequestSignature
	_depositRequestIndex
)

type DepositRequestView struct {
	*ContainerView
}

func NewDepositRequestView() *DepositRequestView {
	return &DepositRequestView{ContainerView: DepositRequestType.New()}
}

func AsDepositRequest(v View, err error) (*DepositRequestView, error) {
	c, err := AsContainer(v, err)
	return &DepositRequestView{c}, err
}

func (d *DepositRequestView) Pubkey() (common.BLSPubkey, error) {
	return common.AsBLSPubkey(d.Get(_depositRequestPubkey))
}
func (d *DepositRequestView) WithdrawalCredentials() (out common.Root, err error) {
	return AsRoot(d.Get(_depositRequestWithdrawalCredentials))
}
func (d *DepositRequestView) SetWithdrawalCredentials(w common.Root) (err error) {
	wCred := RootView(w)
	return d.Set(_depositRequestWithdrawalCredentials, &wCred)
}
func (d *DepositRequestView) Amount() (common.Gwei, error) {
	return common.AsGwei(d.Get(_depositRequestAmount))
}
func (d *DepositRequestView) SetAmount(a common.Gwei) error {
	return d.Set(_depositRequestAmount, Uint64View(a))
}
func (d *DepositRequestView) Signature() (common.BLSSignature, error) {
	return common.AsBLSSignature(d.Get(_depositRequestSignature))
}
func (d *DepositRequestView) SetSignature(s common.BLSSignature) error {
	return d.Set(_depositRequestSignature, common.ViewSignature(&s))
}
func (d *DepositRequestView) Index() (common.Number, error) {
	return common.AsNumber(d.Get(_depositRequestIndex))
}
func (d *DepositRequestView) SetIndex(number common.Number) error {
	return d.Set(_depositRequestIndex, Uint64View(number))
}

type DepositRequests []DepositRequest

func (d *DepositRequests) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*d)
		*d = append(*d, DepositRequest{})
		return spec.Wrap(&((*d)[i]))
	}, DepositRequestType.TypeByteLength(), uint64(spec.MAX_DEPOSIT_REQUESTS_PER_PAYLOAD))
}

func (d DepositRequests) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return spec.Wrap(&d[i])
	}, DepositRequestType.TypeByteLength(), uint64(len(d)))
}

func (d DepositRequests) ByteLength(spec *common.Spec) uint64 {
	return DepositRequestType.TypeByteLength() * uint64(len(d))
}

func (d *DepositRequests) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li DepositRequests) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(&li[i])
		}
		return nil
	}, length, uint64(spec.MAX_DEPOSIT_REQUESTS_PER_PAYLOAD))
}

func DepositRequestsType(spec *common.Spec) *ComplexListTypeDef {
	return ComplexListType(DepositRequestType, uint64(spec.MAX_DEPOSIT_REQUESTS_PER_PAYLOAD))
}

type DepositRequestsView struct{ *ComplexListView }

func AsDepositRequests(v View, err error) (*DepositRequestsView, error) {
	c, err := AsComplexList(v, err)
	return &DepositRequestsView{c}, err
}

func (d *DepositRequestsView) Append(deposit DepositRequest) error {
	v := deposit.View()
	return d.ComplexListView.Append(v)
}

type WithdrawalRequest struct {
	SourceAddress common.Root      `json:"source_address" yaml:"source_address"`
	Pubkey        common.BLSPubkey `json:"pubkey" yaml:"pubkey"`
	Amount        common.Gwei      `json:"amount" yaml:"amount"`
}

func (r *WithdrawalRequest) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&r.SourceAddress, &r.Pubkey, &r.Amount)
}

func (r *WithdrawalRequest) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&r.SourceAddress, &r.Pubkey, &r.Amount)
}

func (*WithdrawalRequest) ByteLength(spec *common.Spec) uint64 {
	return WithdrawalRequestType.TypeByteLength()
}

func (*WithdrawalRequest) FixedLength(spec *common.Spec) uint64 {
	return WithdrawalRequestType.TypeByteLength()
}

func (r *WithdrawalRequest) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&r.SourceAddress, &r.Pubkey, &r.Amount)
}

func (r *WithdrawalRequest) View() *WithdrawalRequestView {
	sAddr := RootView(r.SourceAddress)
	c, _ := DepositRequestType.FromFields(
		&sAddr,
		common.ViewPubkey(&r.Pubkey),
		Uint64View(r.Amount),
	)
	return &WithdrawalRequestView{c}
}

var WithdrawalRequestType = ContainerType("WithdrawalRequest", []FieldDef{
	{"source_address", common.Bytes32Type},
	{"pubkey", common.BLSPubkeyType},
	{"amount", common.GweiType},
})

const (
	_withdrawalRequestSourceAddress = iota
	_withdrawalRequestPubkey
	_withdrawalRequestAmount
)

type WithdrawalRequestView struct {
	*ContainerView
}

func NewWithdrawalRequestView() *WithdrawalRequestView {
	return &WithdrawalRequestView{ContainerView: WithdrawalRequestType.New()}
}

func AsWithdrawalRequest(v View, err error) (*WithdrawalRequestView, error) {
	c, err := AsContainer(v, err)
	return &WithdrawalRequestView{c}, err
}

func (d *WithdrawalRequestView) SourceAddress() (out common.Root, err error) {
	return AsRoot(d.Get(_withdrawalRequestSourceAddress))
}
func (d *WithdrawalRequestView) Pubkey() (common.BLSPubkey, error) {
	return common.AsBLSPubkey(d.Get(_withdrawalRequestPubkey))
}
func (d *WithdrawalRequestView) Amount() (common.Gwei, error) {
	return common.AsGwei(d.Get(_withdrawalRequestAmount))
}
func (d *WithdrawalRequestView) SetAmount(a common.Gwei) error {
	return d.Set(_withdrawalRequestAmount, Uint64View(a))
}

type WithdrawalRequests []WithdrawalRequest

func (d *WithdrawalRequests) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*d)
		*d = append(*d, WithdrawalRequest{})
		return spec.Wrap(&((*d)[i]))
	}, WithdrawalRequestType.TypeByteLength(), uint64(spec.MAX_WITHDRAWAL_REQUESTS_PER_PAYLOAD))
}

func (d WithdrawalRequests) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return spec.Wrap(&d[i])
	}, WithdrawalRequestType.TypeByteLength(), uint64(len(d)))
}

func (d WithdrawalRequests) ByteLength(spec *common.Spec) uint64 {
	return WithdrawalRequestType.TypeByteLength() * uint64(len(d))
}

func (d *WithdrawalRequests) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li WithdrawalRequests) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(&li[i])
		}
		return nil
	}, length, uint64(spec.MAX_WITHDRAWAL_REQUESTS_PER_PAYLOAD))
}

func WithdrawalRequestsType(spec *common.Spec) *ComplexListTypeDef {
	return ComplexListType(WithdrawalRequestType, uint64(spec.MAX_WITHDRAWAL_REQUESTS_PER_PAYLOAD))
}

type WithdrawalRequestsView struct{ *ComplexListView }

func AsWithdrawalRequests(v View, err error) (*WithdrawalRequestsView, error) {
	c, err := AsComplexList(v, err)
	return &WithdrawalRequestsView{c}, err
}

func (d *WithdrawalRequestsView) Append(withdrawal WithdrawalRequest) error {
	v := withdrawal.View()
	return d.ComplexListView.Append(v)
}
