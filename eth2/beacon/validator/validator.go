package validator

import (
	. "github.com/protolambda/zrnt/eth2/core"
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

var ValidatorType = &ContainerType{
	{"pubkey", BLSPubkeyType},
	{"withdrawal_credentials", Bytes32Type}, // Commitment to pubkey for withdrawals
	{"effective_balance", GweiType},         // Balance at stake
	{"slashed", BoolType},
	// Status epochs
	{"activation_eligibility_epoch", EpochType}, // When criteria for activation were met
	{"activation_epoch", EpochType},
	{"exit_epoch", EpochType},
	{"withdrawable_epoch", EpochType}, // When validator can withdraw funds
}

func (v *Validator) IsActive(epoch Epoch) bool {
	return v.ActivationEpoch <= epoch && epoch < v.ExitEpoch
}

func (v *Validator) IsSlashable(epoch Epoch) bool {
	return !v.Slashed && v.ActivationEpoch <= epoch && epoch < v.WithdrawableEpoch
}
