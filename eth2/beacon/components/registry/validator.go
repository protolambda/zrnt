package registry

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type Validator struct {
	Pubkey                BLSPubkey
	WithdrawalCredentials Root // Commitment to pubkey for withdrawals and transfers
	EffectiveBalance      Gwei // Balance at stake
	Slashed               bool

	// Status epochs
	ActivationEligibilityEpoch Epoch // When criteria for activation were met
	ActivationEpoch            Epoch
	ExitEpoch                  Epoch
	WithdrawableEpoch          Epoch // When validator can withdraw or transfer funds
}

func (v *Validator) IsActive(epoch Epoch) bool {
	return v.ActivationEpoch <= epoch && epoch < v.ExitEpoch
}

func (v *Validator) IsSlashable(epoch Epoch) bool {
	return !v.Slashed && v.ActivationEpoch <= epoch && epoch < v.WithdrawableEpoch
}
