package deposits

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

var DepositDataSSZ = zssz.GetSSZ((*DepositData)(nil))

type DepositData struct {
	Pubkey                BLSPubkey
	WithdrawalCredentials Root
	Amount                Gwei
	Signature             BLSSignature
}

func (d *DepositData) ToMessage() *DepositMessage {
	return &DepositMessage{
		Pubkey:                d.Pubkey,
		WithdrawalCredentials: d.WithdrawalCredentials,
		Amount:                d.Amount,
	}
}

func (d *DepositData) MessageRoot() Root {
	return ssz.HashTreeRoot(d.ToMessage(), DepositMessageSSZ)
}

var DepositMessageSSZ = zssz.GetSSZ((*DepositMessage)(nil))

type DepositMessage struct {
	Pubkey                BLSPubkey
	WithdrawalCredentials Root
	Amount                Gwei
}

var DepositDataType = &ContainerType{
	{"pubkey", BLSPubkeyType},
	{"withdrawal_credentials", Bytes32Type},
	{"amount", GweiType},
	{"signature", BLSSignatureType},
}
