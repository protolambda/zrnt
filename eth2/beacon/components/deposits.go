package components

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zssz"
)

var DepositDataSSZ = zssz.GetSSZ((*DepositData)(nil))

type DepositData struct {
	// BLS pubkey
	Pubkey BLSPubkey
	// Withdrawal credentials
	WithdrawalCredentials Root
	// Amount in Gwei
	Amount Gwei
	// Container self-signature
	Signature BLSSignature
}
