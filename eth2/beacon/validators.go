package beacon

import (
	. "github.com/protolambda/zrnt/eth2/beacon/validator"

	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/props"
	. "github.com/protolambda/ztyp/view"
	"sort"
)


const ValidatorIndexType = Uint64Type

// Index of a validator, pointing to a validator registry location
type ValidatorIndex uint64

type ValidatorIndexProp Uint64ReadProp

func (p ValidatorIndexProp) ValidatorIndex() (ValidatorIndex, error) {
	v, err := Uint64ReadProp(p).Uint64()
	return ValidatorIndex(v), err
}

// Custom constant, not in spec:
// An impossible high validator index used to mark special internal cases. (all 1s binary)
const ValidatorIndexMarker = ValidatorIndex(^uint64(0))

type RegistryIndices []ValidatorIndex

func (*RegistryIndices) Limit() uint64 {
	return VALIDATOR_REGISTRY_LIMIT
}

var registryIndicesSSZ = zssz.GetSSZ((*RegistryIndices)(nil))

func (v *RegistryIndices) HashTreeRoot() Root {
	return ssz.HashTreeRoot(v, registryIndicesSSZ)
}

// Collection of validators, should always be sorted.
type ValidatorSet []ValidatorIndex

func (vs ValidatorSet) Len() int {
	return len(vs)
}

func (vs ValidatorSet) Less(i int, j int) bool {
	return vs[i] < vs[j]
}

func (vs ValidatorSet) Swap(i int, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}

// De-duplicates entries in the set in-place. (util to make a valid set from a list with duplicates)
func (vs *ValidatorSet) Dedup() {
	data := *vs
	if len(data) == 0 {
		return
	}
	sort.Sort(vs)
	j := 0
	for i := 1; i < len(data); i++ {
		if data[j] == data[i] {
			continue
		}
		j++
		data[j] = data[i]
	}
	*vs = data[:j+1]
}

// merges with other disjoint set, producing a new set.
func (vs ValidatorSet) MergeDisjoint(other ValidatorSet) ValidatorSet {
	total := len(vs) + len(other)
	out := make(ValidatorSet, 0, total)
	a, b := 0, 0
	for i := 0; i < total; i++ {
		if a < len(vs) {
			if b >= len(other) || vs[a] < other[b] {
				out = append(out, vs[a])
				a++
				continue
			} else if vs[a] == other[b] {
				panic("invalid disjoint sets merge, sets contain equal item")
			}
		}
		if b < len(other) {
			if b < len(other) && (a >= len(vs) || vs[a] > other[b]) {
				out = append(out, other[b])
				b++
				continue
			}
		}
	}
	return out
}

// Joins two validator sets: check if there is any overlap
func (vs ValidatorSet) Intersects(target ValidatorSet) bool {
	// index for source set side of the zig-zag
	i := 0
	// index for target set side of the zig-zag
	j := 0
	var iV, jV ValidatorIndex
	updateI := func() {
		// if out of bounds, just update to an impossibly high index
		if i < len(vs) {
			iV = vs[i]
		} else {
			iV = ValidatorIndexMarker
		}
	}
	updateJ := func() {
		// if out of bounds, just update to an impossibly high index
		if j < len(target) {
			jV = target[j]
		} else {
			jV = ValidatorIndexMarker
		}
	}
	updateI()
	updateJ()
	for {
		// at some point all items in vs have been processed.
		if i >= len(vs) {
			break
		}
		if iV == jV {
			return true
		} else if iV < jV {
			// go to next
			i++
			updateI()
		} else if iV > jV {
			// if the index is higher than the current item in the target, go to the next item.
			j++
			updateJ()
		}
	}
	return false
}

// Joins two validator sets:
// reports all indices of source that are in the target (call onIn), and those that are not (call onOut)
func (vs ValidatorSet) ZigZagJoin(target ValidatorSet, onIn func(i ValidatorIndex), onOut func(i ValidatorIndex)) {
	// index for source set side of the zig-zag
	i := 0
	// index for target set side of the zig-zag
	j := 0
	var iV, jV ValidatorIndex
	updateI := func() {
		// if out of bounds, just update to an impossibly high index
		if i < len(vs) {
			iV = vs[i]
		} else {
			iV = ValidatorIndexMarker
		}
	}
	updateJ := func() {
		// if out of bounds, just update to an impossibly high index
		if j < len(target) {
			jV = target[j]
		} else {
			jV = ValidatorIndexMarker
		}
	}
	updateI()
	updateJ()
	for {
		// at some point all items in vs have been processed.
		if i >= len(vs) {
			break
		}
		if iV == jV {
			if onIn != nil {
				onIn(iV)
			}
			// go to next
			i++
			updateI()
			j++
			updateJ()
		} else if iV < jV {
			// if the index is lower than the current item in the target, it's not in the target.
			if onOut != nil {
				onOut(iV)
			}
			// go to next
			i++
			updateI()
		} else if iV > jV {
			// if the index is higher than the current item in the target, go to the next item.
			j++
			updateJ()
		}
	}
}


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
	v, err := ContainerProp(PropReader(state, uint64(index))).Container()
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

func (state *ValidatorsRegistry) IsSlashed(index ValidatorIndex) (bool, error) {
	v, err := state.Validator(index)
	if err != nil {
		return false, err
	}
	return v.Slashed()
}

func (state *ValidatorsRegistry) ProcessActivationQueue(currentEpoch Epoch) error {
	// Dequeued validators for activation up to churn limit (without resetting activation epoch)
	queueLen := uint64(len(activationQueue))
	if churnLimit, err := state.GetChurnLimit(); err != nil {
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

func (state *ValidatorsRegistry) EffectiveBalance(index ValidatorIndex) (Gwei, error) {
	v, err := state.Validator(index)
	if err != nil {
		return 0, err
	}
	return v.EffectiveBalance()
}
