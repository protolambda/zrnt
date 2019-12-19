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

func (v *Validator) IsActive(epoch Epoch) bool {
	return v.ActivationEpoch <= epoch && epoch < v.ExitEpoch
}

func (v *Validator) IsSlashable(epoch Epoch) bool {
	return !v.Slashed && v.ActivationEpoch <= epoch && epoch < v.WithdrawableEpoch
}

func (v *Validator) IsEligibleForActivationQueue() bool {
	return v.ActivationEligibilityEpoch == FAR_FUTURE_EPOCH && v.EffectiveBalance == MAX_EFFECTIVE_BALANCE
}

func (v *Validator) IsEligibleForActivation(finalizedEpoch Epoch) bool {
	return v.ActivationEligibilityEpoch <= finalizedEpoch && v.ActivationEpoch == FAR_FUTURE_EPOCH
}
