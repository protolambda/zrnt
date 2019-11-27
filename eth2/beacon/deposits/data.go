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
	Signature             BLSSignature
}

var DepositDataType = &ContainerType{
	{"pubkey", BLSPubkeyType},
	{"withdrawal_credentials", Bytes32Type},
	{"amount", GweiType},
	{"signature", BLSSignatureType},
}
