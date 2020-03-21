package registry

import (
	. "github.com/protolambda/zrnt/eth2/beacon/validator"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/math"
	. "github.com/protolambda/ztyp/props"
	. "github.com/protolambda/ztyp/view"
	"sort"
)

var ValidatorsRegistryType = ListType(ValidatorType, VALIDATOR_REGISTRY_LIMIT)

type ValidatorsRegistry struct { *ListView }

func (state *ValidatorsRegistry) IsValidIndex(index ValidatorIndex) (bool, error) {
	count, err := state.ValidatorCount()
	return index < ValidatorIndex(count), err
}

func (state *ValidatorsRegistry) ValidatorCount() (uint64, error) {
	return state.Length()
}

func (state *ValidatorsRegistry) Validator(index ValidatorIndex) (*Validator, error) {
	v, err := ContainerReadProp(PropReader(state, uint64(index))).Container()
	if err != nil {
		return nil, err
	}
	return &Validator{ContainerView: v}, nil
}

func (state *ValidatorsRegistry) Pubkey(index ValidatorIndex) (BLSPubkey, error) {
	v, err := state.Validator(index)
	if err != nil {
		return BLSPubkey{}, err
	}
	return v.Pubkey()
}

// TODO: probably really slow, should have a pubkey cache or something
func (state *ValidatorsRegistry) ValidatorIndex(pubkey BLSPubkey) (index ValidatorIndex, exists bool, err error) {
	count, err := state.ValidatorCount()
	if err != nil {
		return 0, false, err
	}
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		v, err := state.Validator(i)
		if err != nil {
			return 0, false, err
		}
		valPub, err := v.Pubkey()
		if err != nil {
			return 0, false, err
		}
		if valPub == pubkey {
			return i, true, nil
		}
	}
	return 0, false, err
}

func (state *ValidatorsRegistry) WithdrawableEpoch(index ValidatorIndex) (Epoch, error) {
	v, err := state.Validator(index)
	if err != nil {
		return 0, err
	}
	return v.WithdrawableEpoch()
}

func (state *ValidatorsRegistry) IsActive(index ValidatorIndex, epoch Epoch) (bool, error) {
	v, err := state.Validator(index)
	if err != nil {
		return false, err
	}
	return v.IsActive(epoch)
}

func (state *ValidatorsRegistry) SlashAndDelayWithdraw(index ValidatorIndex, withdrawalEpoch Epoch) error {
	v, err := state.Validator(index)
	if err != nil {
		return err
	}
	if err := v.MakeSlashed(); err != nil {
		return err
	}
	prevWithdrawalEpoch, err := v.WithdrawableEpoch()
	if err != nil {
		return err
	}
	if withdrawalEpoch > prevWithdrawalEpoch {
		if err := v.SetWithdrawableEpoch(withdrawalEpoch); err != nil {
			return err
		}
	}
	return nil
}

func (state *ValidatorsRegistry) GetActiveValidatorIndices(epoch Epoch) (RegistryIndices, error) {
	count, err := state.ValidatorCount()
	if err != nil {
		return nil, err
	}
	res := make(RegistryIndices, 0, count)
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		v, err := state.Validator(i)
		if err != nil {
			return nil, err
		}
		if active, err := v.IsActive(epoch); err != nil {
			return nil, err
		} else if active {
			res = append(res, i)
		}
	}
	return res, nil
}

func (state *ValidatorsRegistry) ComputeActiveIndexRoot(epoch Epoch) (Root, error) {
	indices, err := state.GetActiveValidatorIndices(epoch)
	if err != nil {
		return Root{}, err
	}
	return indices.HashTreeRoot(), nil
}

func (state *ValidatorsRegistry) GetActiveValidatorCount(epoch Epoch) (activeCount uint64, err error) {
	count, err := state.ValidatorCount()
	if err != nil {
		return 0, err
	}
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		v, err := state.Validator(i)
		if err != nil {
			return 0, err
		}
		if active, err := v.IsActive(epoch); err != nil {
			return 0, err
		} else if active {
			activeCount++
		}
	}
	return
}

func CommitteeCount(activeValidators uint64) uint64 {
	validatorsPerSlot := activeValidators / uint64(SLOTS_PER_EPOCH)
	committeesPerSlot := validatorsPerSlot / TARGET_COMMITTEE_SIZE
	if MAX_COMMITTEES_PER_SLOT < committeesPerSlot {
		committeesPerSlot = MAX_COMMITTEES_PER_SLOT
	}
	if committeesPerSlot == 0 {
		committeesPerSlot = 1
	}
	return committeesPerSlot
}

func (state *ValidatorsRegistry) GetCommitteeCountAtSlot(slot Slot) (uint64, error) {
	activeCount, err := state.GetActiveValidatorCount(slot.ToEpoch())
	if err != nil {
		return 0, err
	}
	return CommitteeCount(activeCount), nil
}

func (state *ValidatorsRegistry) IsSlashed(index ValidatorIndex) (bool, error) {
	v, err := state.Validator(index)
	if err != nil {
		return false, err
	}
	return v.Slashed()
}

func (state *ValidatorsRegistry) GetIndicesToSlash(withdrawal Epoch) (out []ValidatorIndex, err error) {
	count, err := state.ValidatorCount()
	if err != nil {
		return nil, err
	}
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		v, err := state.Validator(i)
		if err != nil {
			return nil, err
		}
		valSlashed, err := v.Slashed()
		if err != nil {
			return nil, err
		}
		valWithdrawal, err := v.WithdrawableEpoch()
		if err != nil {
			return nil, err
		}
		if valSlashed && withdrawal == valWithdrawal {
			out = append(out, ValidatorIndex(i))
		}
	}
	return
}

func (state *ValidatorsRegistry) GetChurnLimit(epoch Epoch) (uint64, error) {
	activeCount, err := state.GetActiveValidatorCount(epoch)
	if err != nil {
		return 0, err
	}
	return math.MaxU64(MIN_PER_EPOCH_CHURN_LIMIT, activeCount/CHURN_LIMIT_QUOTIENT), nil
}

func (state *ValidatorsRegistry) ExitQueueEnd(epoch Epoch) (Epoch, error) {
	count, err := state.ValidatorCount()
	if err != nil {
		return 0, err
	}
	// Compute exit queue epoch
	exitQueueEnd := epoch.ComputeActivationExitEpoch()
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		v, err := state.Validator(i)
		if err != nil {
			return 0, err
		}
		valExit, err := v.ExitEpoch()
		if err != nil {
			return 0, err
		}
		if valExit != FAR_FUTURE_EPOCH && valExit > exitQueueEnd {
			exitQueueEnd = valExit
		}
	}
	exitQueueChurn := uint64(0)
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		v, err := state.Validator(i)
		if err != nil {
			return 0, err
		}
		valExit, err := v.ExitEpoch()
		if valExit == exitQueueEnd {
			exitQueueChurn++
		}
	}
	churnLimit, err := state.GetChurnLimit(epoch)
	if err != nil {
		return 0, err
	}
	if exitQueueChurn >= churnLimit {
		exitQueueEnd++
	}
	return exitQueueEnd, nil
}

type activationQueueItem struct {
	valIndex ValidatorIndex
	activation Epoch
	activationEligibility Epoch
}

func (state *ValidatorsRegistry) ProcessActivationQueue(activationEpoch Epoch, currentEpoch Epoch) error {
	activationQueue := make([]activationQueueItem, 0)
	count, err := state.ValidatorCount()
	if err != nil {
		return err
	}
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		v, err := state.Validator(i)
		if err != nil {
			return err
		}
		valActivationEligibilityEpoch, err := v.ActivationEligibilityEpoch()
		if err != nil {
			return err
		}
		valActivationEpoch, err := v.ActivationEpoch()
		if err != nil {
			return err
		}
		if valActivationEligibilityEpoch != FAR_FUTURE_EPOCH && valActivationEpoch >= activationEpoch {
			activationQueue = append(activationQueue, activationQueueItem{
				valIndex: i,
				activation: valActivationEpoch,
				activationEligibility: valActivationEligibilityEpoch,
			})

		}
	}

	// Order by the sequence of activation_eligibility_epoch setting and then index
	sort.Slice(activationQueue, func(i int, j int) bool {
		return activationQueue[i].activationEligibility < activationQueue[j].activationEligibility
	})
	// Dequeued validators for activation up to churn limit (without resetting activation epoch)
	queueLen := uint64(len(activationQueue))
	if churnLimit, err := state.GetChurnLimit(currentEpoch); err != nil {
		return err
	} else if churnLimit < queueLen {
		queueLen = churnLimit
	}

	for _, item := range activationQueue[:queueLen] {
		if item.activation == FAR_FUTURE_EPOCH {
			v, err := state.Validator(item.valIndex)
			if err != nil {
				return err
			}
			if err := v.SetActivationEpoch(currentEpoch.ComputeActivationExitEpoch()); err != nil {
				return err
			}
		}
	}
	return nil
}

// Return the total balance sum (1 Gwei minimum to avoid divisions by zero.)
func (state *ValidatorsRegistry) GetTotalStakedBalance(epoch Epoch) (sum Gwei, err error) {
	count, err := state.ValidatorCount()
	if err != nil {
		return 0, err
	}
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		v, err := state.Validator(i)
		if err != nil {
			return 0, err
		}

		if active, err := v.IsActive(epoch); err != nil {
			return 0, err
		} else if active {
			effBal, err := v.EffectiveBalance()
			if err != nil {
				return 0, err
			}
			sum += effBal
		}
	}
	return sum, nil
}

func (state *ValidatorsRegistry) GetAttestersStake(statuses []AttesterStatus, mask AttesterFlag) (out Gwei, err error) {
	for i := range statuses {
		status := &statuses[i]
		v, err := state.Validator(ValidatorIndex(i))
		if err != nil {
			return 0, err
		}
		b, err := v.EffectiveBalance()
		if err != nil {
			return 0, err
		}
		if status.Flags.HasMarkers(mask) {
			out += b
		}
	}
	if out == 0 {
		return 1, nil
	}
	return
}

func (state *ValidatorsRegistry) GetTotalStake() (out Gwei, err error) {
	count, err := state.ValidatorCount()
	if err != nil {
		return 0, err
	}
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		v, err := state.Validator(i)
		if err != nil {
			return 0, err
		}
		b, err := v.EffectiveBalance()
		if err != nil {
			return 0, err
		}
		out += b
	}
	if out == 0 {
		return 1, nil
	}
	return
}

func (state *ValidatorsRegistry) EffectiveBalance(index ValidatorIndex) (Gwei, error) {
	v, err := state.Validator(index)
	if err != nil {
		return 0, err
	}
	return v.EffectiveBalance()
}

// Return the combined effective balance of an array of validators. (1 Gwei minimum to avoid divisions by zero.)
func (state *ValidatorsRegistry) SumEffectiveBalanceOf(indices []ValidatorIndex) (sum Gwei, err error) {
	for _, vIndex := range indices {
		v, err := state.Validator(vIndex)
		if err != nil {
			return 0, err
		}
		b, err := v.EffectiveBalance()
		if err != nil {
			return 0, err
		}
		sum += b
	}
	if sum == 0 {
		return 1, nil
	}
	return
}
