package registry

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/validator"
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

func (vr ValidatorRegistry) IsValidatorIndex(index ValidatorIndex) bool {
	return index < ValidatorIndex(len(vr))
}

func (vr ValidatorRegistry) ValidatorIndex(pubkey BLSPubkey) (index ValidatorIndex, exists bool) {
	valIndex := ValidatorIndexMarker
	for i, v := range vr {
		if v.Pubkey == pubkey {
			valIndex = ValidatorIndex(i)
			break
		}
	}
	return valIndex, valIndex != ValidatorIndexMarker
}

func (vr ValidatorRegistry) GetActiveValidatorIndices(epoch Epoch) RegistryIndices {
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

func (vr ValidatorRegistry) GetIndicesToSlash(withdrawal Epoch) (out []ValidatorIndex) {
	for i, v := range vr {
		if v.Slashed && withdrawal == v.WithdrawableEpoch {
			out = append(out, ValidatorIndex(i))
		}
	}
	return
}

func (vr ValidatorRegistry) GetChurnLimit(epoch Epoch) uint64 {
	return math.MaxU64(MIN_PER_EPOCH_CHURN_LIMIT, vr.GetActiveValidatorCount(epoch)/CHURN_LIMIT_QUOTIENT)
}


func (vr ValidatorRegistry) ExitQueueEnd(epoch Epoch) Epoch {
	// Compute exit queue epoch
	exitQueueEnd := epoch.ComputeActivationExitEpoch()
	for _, v := range vr {
		if v.ExitEpoch != FAR_FUTURE_EPOCH && v.ExitEpoch > exitQueueEnd {
			exitQueueEnd = v.ExitEpoch
		}
	}
	exitQueueChurn := uint64(0)
	for _, v := range vr {
		if v.ExitEpoch == exitQueueEnd {
			exitQueueChurn++
		}
	}
	if exitQueueChurn >= vr.GetChurnLimit(epoch) {
		exitQueueEnd++
	}
	return exitQueueEnd
}

func (vr ValidatorRegistry) ProcessActivationQueue(activationEpoch Epoch, currentEpoch Epoch) {
	activationQueue := make([]*Validator, 0)
	for _, v := range vr {
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
	if churnLimit := vr.GetChurnLimit(currentEpoch); churnLimit < queueLen {
		queueLen = churnLimit
	}
	for _, v := range activationQueue[:queueLen] {
		if v.ActivationEpoch == FAR_FUTURE_EPOCH {
			v.ActivationEpoch = currentEpoch.ComputeActivationExitEpoch()
		}
	}
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
