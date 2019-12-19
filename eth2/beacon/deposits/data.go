package deposits

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zssz"
)

var DepositDataSSZ = zssz.GetSSZ((*DepositData)(nil))

type DepositData struct {
	Pubkey                BLSPubkey
	WithdrawalCredentials Root
	Amount                Gwei
	// signing over DepositMessage
	Signature             BLSSignature
}

func (data *DepositData) Message() DepositMessage {
	return DepositMessage{
		Pubkey:                data.Pubkey,
		WithdrawalCredentials: data.WithdrawalCredentials,
		Amount:                data.Amount,
	}
}

var DepositMessageSSZ = zssz.GetSSZ((*DepositMessage)(nil))

type DepositMessage struct {
	Pubkey                BLSPubkey
	WithdrawalCredentials Root
	Amount                Gwei
}
