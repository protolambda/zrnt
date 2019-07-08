package registry

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type Validator struct {
	// BLS public key
	Pubkey BLSPubkey
	// Withdrawal credentials
	WithdrawalCredentials Root
	// Epoch when became eligible for activation
	ActivationEligibilityEpoch Epoch
	// Epoch when validator activated
	ActivationEpoch Epoch
	// Epoch when validator exited
	ExitEpoch Epoch
	// Epoch when validator is eligible to withdraw
	WithdrawableEpoch Epoch
	// Was the validator slashed
	Slashed bool
	// Effective balance
	EffectiveBalance Gwei
}

func (v *Validator) IsActive(epoch Epoch) bool {
	return v.ActivationEpoch <= epoch && epoch < v.ExitEpoch
}

func (v *Validator) IsSlashable(epoch Epoch) bool {
	return !v.Slashed && v.ActivationEpoch <= epoch && epoch < v.WithdrawableEpoch
}
