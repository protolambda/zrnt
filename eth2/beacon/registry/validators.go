package registry

import (
	. "github.com/protolambda/zrnt/eth2/beacon/validator"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/zssz"
	"sort"
)

type RegistryIndices []ValidatorIndex

func (_ *RegistryIndices) Limit() uint64 {
	return VALIDATOR_REGISTRY_LIMIT
}

var RegistryIndicesSSZ = zssz.GetSSZ((*RegistryIndices)(nil))

type ValidatorRegistry []*Validator

func (_ *ValidatorRegistry) Limit() uint64 {
	return VALIDATOR_REGISTRY_LIMIT
}

type ValidatorsState struct {
	Validators ValidatorRegistry
}

func (state *ValidatorsState) IsValidIndex(index ValidatorIndex) bool {
	return index < ValidatorIndex(len(state.Validators))
}

func (state *ValidatorsState) ValidatorCount() uint64 {
	return uint64(len(state.Validators))
}

func (state *ValidatorsState) Validator(index ValidatorIndex) *Validator {
	return state.Validators[index]
}

func (state *ValidatorsState) Pubkey(index ValidatorIndex) BLSPubkey {
	return state.Validators[index].Pubkey
}

func (state *ValidatorsState) ValidatorIndex(pubkey BLSPubkey) (index ValidatorIndex, exists bool) {
	for i, v := range state.Validators {
		if v.Pubkey == pubkey {
			return ValidatorIndex(i), true
		}
	}
	return ValidatorIndexMarker, false
}

func (state *ValidatorsState) GetActiveValidatorIndices(epoch Epoch) RegistryIndices {
	res := make([]ValidatorIndex, 0, len(state.Validators))
	for i, v := range state.Validators {
		if v.IsActive(epoch) {
			res = append(res, ValidatorIndex(i))
		}
	}
	return res
}

func (state *ValidatorsState) GetActiveValidatorCount(epoch Epoch) (count uint64) {
	for _, v := range state.Validators {
		if v.IsActive(epoch) {
			count++
		}
	}
	return
}

func (state *ValidatorsState) IsSlashed(index ValidatorIndex) bool {
	return state.Validators[index].Slashed
}

// Filters a slice in-place. Only keeps the unslashed validators.
// If input is sorted, then the result will be sorted.
func (state *ValidatorsState) FilterUnslashed(indices []ValidatorIndex) []ValidatorIndex {
	unslashed := indices[:0]
	for _, x := range indices {
		if !state.Validators[x].Slashed {
			unslashed = append(unslashed, x)
		}
	}
	return unslashed
}

func (state *ValidatorsState) GetIndicesToSlash(withdrawal Epoch) (out []ValidatorIndex) {
	for i, v := range state.Validators {
		if v.Slashed && withdrawal == v.WithdrawableEpoch {
			out = append(out, ValidatorIndex(i))
		}
	}
	return
}

func (state *ValidatorsState) GetChurnLimit(epoch Epoch) uint64 {
	return math.MaxU64(MIN_PER_EPOCH_CHURN_LIMIT, state.GetActiveValidatorCount(epoch)/CHURN_LIMIT_QUOTIENT)
}


func (state *ValidatorsState) ExitQueueEnd(epoch Epoch) Epoch {
	// Compute exit queue epoch
	exitQueueEnd := epoch.ComputeActivationExitEpoch()
	for _, v := range state.Validators {
		if v.ExitEpoch != FAR_FUTURE_EPOCH && v.ExitEpoch > exitQueueEnd {
			exitQueueEnd = v.ExitEpoch
		}
	}
	exitQueueChurn := uint64(0)
	for _, v := range state.Validators {
		if v.ExitEpoch == exitQueueEnd {
			exitQueueChurn++
		}
	}
	if exitQueueChurn >= state.GetChurnLimit(epoch) {
		exitQueueEnd++
	}
	return exitQueueEnd
}

func (state *ValidatorsState) ProcessActivationQueue(activationEpoch Epoch, currentEpoch Epoch) {
	activationQueue := make([]*Validator, 0)
	for _, v := range state.Validators {
		if v.ActivationEligibilityEpoch != FAR_FUTURE_EPOCH &&
			v.ActivationEpoch >= activationEpoch {
			activationQueue = append(activationQueue, v)
		}
	}
	sort.Slice(activationQueue, func(i int, j int) bool {
		return activationQueue[i].ActivationEligibilityEpoch <
			activationQueue[j].ActivationEligibilityEpoch
	})
	// Dequeued validators for activation up to churn limit (without resetting activation epoch)
	queueLen := uint64(len(activationQueue))
	if churnLimit := state.GetChurnLimit(currentEpoch); churnLimit < queueLen {
		queueLen = churnLimit
	}
	for _, v := range activationQueue[:queueLen] {
		if v.ActivationEpoch == FAR_FUTURE_EPOCH {
			v.ActivationEpoch = currentEpoch.ComputeActivationExitEpoch()
		}
	}
}

// Return the total balance sum (1 Gwei minimum to avoid divisions by zero.)
func (state *ValidatorsState) GetTotalStakedBalance(epoch Epoch) (sum Gwei) {
	for _, v := range state.Validators {
		if v.IsActive(epoch) {
			sum += v.EffectiveBalance
		}
	}
	if sum == 0 {
		return 1
	}
	return sum
}

func (state *ValidatorsState) EffectiveBalance(index ValidatorIndex) Gwei {
	return state.Validators[index].EffectiveBalance
}

// Return the combined effective balance of an array of validators. (1 Gwei minimum to avoid divisions by zero.)
func (state *ValidatorsState) SumEffectiveBalanceOf(indices []ValidatorIndex) (sum Gwei) {
	for _, vIndex := range indices {
		sum += state.Validators[vIndex].EffectiveBalance
	}
	if sum == 0 {
		return 1
	}
	return sum
}
