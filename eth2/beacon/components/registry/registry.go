package registry

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zssz"
)

// Validator registry
type RegistryState struct {
	Validators ValidatorRegistry
	Balances   Balances
}

// Update effective balances with hysteresis
func (state *RegistryState) UpdateEffectiveBalances() {
	for i, v := range state.Validators {
		balance := state.Balances[i]
		if balance < v.EffectiveBalance ||
			v.EffectiveBalance+3*HALF_INCREMENT < balance {
			v.EffectiveBalance = balance - (balance % EFFECTIVE_BALANCE_INCREMENT)
			if MAX_EFFECTIVE_BALANCE < v.EffectiveBalance {
				v.EffectiveBalance = MAX_EFFECTIVE_BALANCE
			}
		}
	}
}

var ValidatorIndexListSSZ = zssz.GetSSZ((*[]ValidatorIndex)(nil))

type ValidatorRegistry []*Validator

func (vr ValidatorRegistry) IsValidatorIndex(index ValidatorIndex) bool {
	return index < ValidatorIndex(len(vr))
}

func (vr ValidatorRegistry) GetActiveValidatorIndices(epoch Epoch) []ValidatorIndex {
	res := make([]ValidatorIndex, 0, len(vr))
	for i, v := range vr {
		if v.IsActive(epoch) {
			res = append(res, ValidatorIndex(i))
		}
	}
	return res
}

func (vr ValidatorRegistry) GetActiveValidatorCount(epoch Epoch) (count uint64) {
	for _, v := range vr {
		if v.IsActive(epoch) {
			count++
		}
	}
	return
}

// Return the total balance sum (1 Gwei minimum to avoid divisions by zero.)
func (vr ValidatorRegistry) GetTotalActiveEffectiveBalance(epoch Epoch) (sum Gwei) {
	for _, v := range vr {
		if v.IsActive(epoch) {
			sum += v.EffectiveBalance
		}
	}
	if sum == 0 {
		return 1
	}
	return sum
}

// Return the combined effective balance of an array of validators. (1 Gwei minimum to avoid divisions by zero.)
func (vr ValidatorRegistry) GetTotalEffectiveBalanceOf(indices []ValidatorIndex) (sum Gwei) {
	for _, vIndex := range indices {
		sum += vr[vIndex].EffectiveBalance
	}
	if sum == 0 {
		return 1
	}
	return sum
}

// Filters a slice in-place. Only keeps the unslashed validators.
// If input is sorted, then the result will be sorted.
func (vr ValidatorRegistry) FilterUnslashed(indices []ValidatorIndex) []ValidatorIndex {
	unslashed := indices[:0]
	for _, x := range indices {
		if !vr[x].Slashed {
			unslashed = append(unslashed, x)
		}
	}
	return unslashed
}
