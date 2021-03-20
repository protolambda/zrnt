package phase0

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/util/hashing"
	"sync"
)

func computeProposerIndex(spec *common.Spec, state *BeaconStateView, indices []common.ValidatorIndex, seed common.Root) (common.ValidatorIndex, error) {
	if len(indices) == 0 {
		return 0, errors.New("no validators available to compute proposer")
	}
	buf := make([]byte, 32+8, 32+8)
	copy(buf[0:32], seed[:])

	registry, err := state.Validators()
	if err != nil {
		return 0, err
	}
	hFn := hashing.GetHashFn()
	for i := uint64(0); i < 1000; i++ {
		binary.LittleEndian.PutUint64(buf[32:], i)
		h := hFn(buf)
		for j := uint64(0); j < 32; j++ {
			randomByte := h[j]
			absI := common.ValidatorIndex(((i << 5) | j) % uint64(len(indices)))
			shuffledI := common.PermuteIndex(spec.SHUFFLE_ROUND_COUNT, absI, uint64(len(indices)), seed)
			candidateIndex := indices[int(shuffledI)]
			validator, err := registry.Validator(candidateIndex)
			if err != nil {
				return 0, err
			}
			effectiveBalance, err := validator.EffectiveBalance()
			if err != nil {
				return 0, err
			}
			if effectiveBalance*0xff >= spec.MAX_EFFECTIVE_BALANCE*common.Gwei(randomByte) {
				return candidateIndex, nil
			}
		}
	}
	return 0, errors.New("random (but balance-biased) infinite scrolling through a committee should always find a proposer")
}

type BoundedIndex struct {
	Index      common.ValidatorIndex
	Activation common.Epoch
	Exit       common.Epoch
}

func (state *BeaconStateView) loadIndicesBounded() ([]BoundedIndex, error) {
	validators, err := state.Validators()
	if err != nil {
		return nil, err
	}
	valCount, err := validators.ValidatorCount()
	if err != nil {
		return nil, err
	}
	indicesBounded := make([]BoundedIndex, valCount, valCount)
	valIterNext := validators.Iter()
	i := common.ValidatorIndex(0)
	for {
		val, ok, err := valIterNext()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		actiEp, err := val.ActivationEpoch()
		if err != nil {
			return nil, err
		}
		exitEp, err := val.ExitEpoch()
		if err != nil {
			return nil, err
		}
		indicesBounded[i] = BoundedIndex{
			Index:      i,
			Activation: actiEp,
			Exit:       exitEp,
		}
		i++
	}
	return indicesBounded, nil
}

// PubkeyCache is shared between any state. However, if .Append(index, pubkey) conflicts, a new cache will be forked out.
type PubkeyCache struct {
	parent *PubkeyCache
	// The count up until the conflicting validator index (cause of the pubkey cache fork).
	trustedParentCount common.ValidatorIndex
	pub2idx            map[common.BLSPubkey]common.ValidatorIndex
	// starting at trustedParentCount
	idx2pub []common.CachedPubkey
	// Can have many reads concurrently, but only 1 write.
	rwLock sync.RWMutex
}

func NewPubkeyCache(state *BeaconStateView) (*PubkeyCache, error) {
	vals, err := state.Validators()
	if err != nil {
		return nil, err
	}
	valCount, err := vals.ValidatorCount()
	if err != nil {
		return nil, err
	}
	pc := &PubkeyCache{
		parent:             nil,
		trustedParentCount: 0,
		pub2idx:            make(map[common.BLSPubkey]common.ValidatorIndex),
		idx2pub:            make([]common.CachedPubkey, 0),
	}
	currentCount := uint64(len(pc.idx2pub))
	for i := currentCount; i < valCount; i++ {
		idx := common.ValidatorIndex(i)
		v, err := vals.Validator(idx)
		if err != nil {
			return nil, err
		}
		pub, err := v.Pubkey()
		if err != nil {
			return nil, err
		}
		pc.pub2idx[pub] = idx
		pc.idx2pub = append(pc.idx2pub, common.CachedPubkey{Compressed: pub})
	}
	return pc, nil
}

func EmptyPubkeyCache() *PubkeyCache {
	return &PubkeyCache{
		parent:             nil,
		trustedParentCount: 0,
		pub2idx:            make(map[common.BLSPubkey]common.ValidatorIndex),
		idx2pub:            make([]common.CachedPubkey, 0),
	}
}

// Get the pubkey of a validator index.
// Note: this does not mean the validator is part of the current state.
// It merely means that this is a known pubkey for that particular validator
// (could be in a later part of a forked version of the state).
func (pc *PubkeyCache) Pubkey(index common.ValidatorIndex) (pub *common.CachedPubkey, ok bool) {
	pc.rwLock.RLock()
	defer pc.rwLock.RUnlock()
	return pc.unsafePubkey(index)
}

func (pc *PubkeyCache) unsafePubkey(index common.ValidatorIndex) (pub *common.CachedPubkey, ok bool) {
	if index >= pc.trustedParentCount {
		if index >= pc.trustedParentCount+common.ValidatorIndex(len(pc.idx2pub)) {
			return nil, false
		}
		return &pc.idx2pub[index-pc.trustedParentCount], true
	} else if pc.parent != nil {
		return pc.parent.Pubkey(index)
	} else {
		return nil, false
	}
}

// Get the validator index of a pubkey.
// Note: this does not mean the validator is part of the current state.
// It merely means that this is a known pubkey for that particular validator
// (could be in a later part of a forked version of the state).
func (pc *PubkeyCache) ValidatorIndex(pubkey common.BLSPubkey) (index common.ValidatorIndex, ok bool) {
	pc.rwLock.RLock()
	defer pc.rwLock.RUnlock()
	return pc.unsafeValidatorIndex(pubkey)
}

func (pc *PubkeyCache) unsafeValidatorIndex(pubkey common.BLSPubkey) (index common.ValidatorIndex, ok bool) {
	index, ok = pc.pub2idx[pubkey]
	if !ok && pc.parent != nil {
		return pc.parent.ValidatorIndex(pubkey)
	}
	return index, ok
}

// AddValidator appends the (index, pubkey) pair to the pubkey cache. It returns the same cache if the added entry is not conflicting.
// If it conflicts, the common part is inherited, and a forked pubkey cache is returned.
func (pc *PubkeyCache) AddValidator(index common.ValidatorIndex, pub common.BLSPubkey) (*PubkeyCache, error) {
	existingIndex, indexExists := pc.ValidatorIndex(pub)
	existingPubkey, pubkeyExists := pc.Pubkey(index)

	if indexExists {
		if existingIndex != index {
			// conflict detected! Deposit log fork!
			forkedPc := &PubkeyCache{
				parent: pc,
				// fork out the existing index, only trust the common history
				trustedParentCount: existingIndex,
				pub2idx:            make(map[common.BLSPubkey]common.ValidatorIndex),
				idx2pub:            make([]common.CachedPubkey, 0),
			}
			// Do not have to unlock this cache (parent of forkedPc) early, as the forkedPc is guaranteed to handle it.
			return forkedPc.AddValidator(index, pub)
		}
		if pubkeyExists {
			if existingPubkey.Compressed != pub {
				// conflict detected! Deposit log fork!
				forkedPc := &PubkeyCache{
					parent: pc,
					// fork out the existing index, only trust the common history
					trustedParentCount: index,
					pub2idx:            make(map[common.BLSPubkey]common.ValidatorIndex),
					idx2pub:            make([]common.CachedPubkey, 0),
				}
				// Do not have to unlock this cache (parent of forkedPc) early, as the forkedPc is guaranteed to handle it.
				return forkedPc.AddValidator(index, pub)
			}
		}
		// append is no-op, validator already exists
		return pc, nil
	}
	if pubkeyExists {
		if existingPubkey.Compressed != pub {
			// conflict detected! Deposit log fork!
			forkedPc := &PubkeyCache{
				parent: pc,
				// fork out the existing index, only trust the common history
				trustedParentCount: index,
				pub2idx:            make(map[common.BLSPubkey]common.ValidatorIndex),
				idx2pub:            make([]common.CachedPubkey, 0),
			}
			// Do not have to unlock this cache (parent of forkedPc) early, as the forkedPc is guaranteed to handle it.
			return forkedPc.AddValidator(index, pub)
		}
	}
	pc.rwLock.Lock()
	defer pc.rwLock.Unlock()
	if expected := pc.trustedParentCount + common.ValidatorIndex(len(pc.idx2pub)); index != expected {
		// index is unknown, but too far ahead of cache; in between indices are missing.
		return nil, fmt.Errorf("AddValidator is incorrect, missing earlier index. got: (%d, %x), but currently expecting %d next", index, pub, expected)
	}
	pc.idx2pub = append(pc.idx2pub, common.CachedPubkey{Compressed: pub})
	pc.pub2idx[pub] = index
	return pc, nil
}

type EpochsContext struct {
	Spec *common.Spec
	// PubkeyCache may be replaced when a new forked-out cache takes over to process an alternative Eth1 deposit chain.
	PubkeyCache *PubkeyCache
	// Proposers is a slice of SLOTS_PER_EPOCH proposer indices for the current epoch
	Proposers []common.ValidatorIndex

	PreviousEpoch *ShufflingEpoch
	CurrentEpoch  *ShufflingEpoch
	NextEpoch     *ShufflingEpoch
}

// NewEpochsContext constructs a new context for the processing of the current epoch.
func NewEpochsContext(spec *common.Spec, state *BeaconStateView) (*EpochsContext, error) {
	pc, err := NewPubkeyCache(state)
	if err != nil {
		return nil, err
	}
	epc := &EpochsContext{
		Spec:        spec,
		PubkeyCache: pc,
	}
	if err := epc.LoadShuffling(state); err != nil {
		return nil, err
	}
	if err := epc.LoadProposers(state); err != nil {
		return nil, err
	}
	return epc, nil
}

func (epc *EpochsContext) LoadShuffling(state *BeaconStateView) error {
	indicesBounded, err := state.loadIndicesBounded()
	if err != nil {
		return err
	}
	slot, err := state.Slot()
	if err != nil {
		return err
	}
	currentEpoch := epc.Spec.SlotToEpoch(slot)
	epc.CurrentEpoch, err = ComputeShufflingEpoch(epc.Spec, state, indicesBounded, currentEpoch)
	if err != nil {
		return err
	}
	prevEpoch := currentEpoch.Previous()
	if prevEpoch == currentEpoch { // in case of genesis
		epc.PreviousEpoch = epc.CurrentEpoch
	} else {
		epc.PreviousEpoch, err = ComputeShufflingEpoch(epc.Spec, state, indicesBounded, prevEpoch)
		if err != nil {
			return err
		}
	}
	epc.NextEpoch, err = ComputeShufflingEpoch(epc.Spec, state, indicesBounded, currentEpoch+1)
	if err != nil {
		return err
	}
	return nil
}

func (epc *EpochsContext) LoadProposers(state *BeaconStateView) error {
	// prerequisite to load shuffling: the list of active indices, same as in the shuffling. So load the shuffling first.
	if epc.CurrentEpoch == nil {
		if err := epc.LoadShuffling(state); err != nil {
			return err
		}
	}
	return epc.resetProposers(state)
}

func (epc *EpochsContext) resetProposers(state *BeaconStateView) error {
	// no validators? No proposers to compute. Keep it nil so we can detect and compute it later.
	if len(epc.CurrentEpoch.ActiveIndices) == 0 {
		return nil
	}
	epc.Proposers = make([]common.ValidatorIndex, epc.Spec.SLOTS_PER_EPOCH, epc.Spec.SLOTS_PER_EPOCH)
	mixes, err := state.RandaoMixes()
	if err != nil {
		return err
	}
	epochSeed, err := common.GetSeed(epc.Spec, mixes, epc.CurrentEpoch.Epoch, epc.Spec.DOMAIN_BEACON_PROPOSER)
	if err != nil {
		return err
	}
	slot, err := epc.Spec.EpochStartSlot(epc.CurrentEpoch.Epoch)
	if err != nil {
		return err
	}
	hFn := hashing.GetHashFn()
	var buf [32 + 8]byte
	copy(buf[0:32], epochSeed[:])
	for i := common.Slot(0); i < epc.Spec.SLOTS_PER_EPOCH; i++ {
		binary.LittleEndian.PutUint64(buf[32:], uint64(slot))
		seed := hFn(buf[:])
		proposer, err := computeProposerIndex(epc.Spec, state, epc.CurrentEpoch.ActiveIndices, seed)
		if err != nil {
			return err
		}
		epc.Proposers[i] = proposer
		slot++
	}
	return nil
}

func (epc *EpochsContext) Clone() *EpochsContext {
	// All fields can be reused, just need a fresh shallow copy of the outer container
	epcClone := *epc
	return &epcClone
}

func (epc *EpochsContext) RotateEpochs(state *BeaconStateView) error {
	epc.PreviousEpoch = epc.CurrentEpoch
	epc.CurrentEpoch = epc.NextEpoch
	nextEpoch := epc.CurrentEpoch.Epoch + 1
	// TODO: could use epoch-processing validator data to not read state here
	indicesBounded, err := state.loadIndicesBounded()
	if err != nil {
		return err
	}
	epc.NextEpoch, err = ComputeShufflingEpoch(epc.Spec, state, indicesBounded, nextEpoch)
	if err != nil {
		return err
	}
	return epc.resetProposers(state)
}

func (epc *EpochsContext) getSlotComms(slot common.Slot) ([][]common.ValidatorIndex, error) {
	epoch := epc.Spec.SlotToEpoch(slot)
	epochSlot := slot % epc.Spec.SLOTS_PER_EPOCH
	if epoch == epc.PreviousEpoch.Epoch {
		return epc.PreviousEpoch.Committees[epochSlot], nil
	} else if epoch == epc.CurrentEpoch.Epoch {
		return epc.CurrentEpoch.Committees[epochSlot], nil
	} else if epoch == epc.NextEpoch.Epoch {
		return epc.NextEpoch.Committees[epochSlot], nil
	} else {
		return nil, fmt.Errorf("beacon committee retrieval: out of range epoch: %d", epoch)
	}
}

// Return the beacon committee at slot for index.
func (epc *EpochsContext) GetBeaconCommittee(slot common.Slot, index common.CommitteeIndex) ([]common.ValidatorIndex, error) {
	if index >= common.CommitteeIndex(epc.Spec.MAX_COMMITTEES_PER_SLOT) {
		return nil, fmt.Errorf("beacon committee retrieval: out of range committee index: %d", index)
	}

	slotComms, err := epc.getSlotComms(slot)
	if err != nil {
		return nil, err
	}

	if index >= common.CommitteeIndex(len(slotComms)) {
		return nil, fmt.Errorf("beacon committee retrieval: out of range committee index: %d", index)
	}
	return slotComms[index], nil
}

func (epc *EpochsContext) GetCommitteeCountAtSlot(slot common.Slot) (uint64, error) {
	slotComms, err := epc.getSlotComms(slot)
	return uint64(len(slotComms)), err
}

func (epc *EpochsContext) GetBeaconProposer(slot common.Slot) (common.ValidatorIndex, error) {
	epoch := epc.Spec.SlotToEpoch(slot)
	if epoch != epc.CurrentEpoch.Epoch {
		return 0, fmt.Errorf("expected epoch %d for proposer lookup, but lookup was at slot %d (epoch %d)", epc.CurrentEpoch.Epoch, slot, epoch)
	}
	return epc.Proposers[slot%epc.Spec.SLOTS_PER_EPOCH], nil
}
