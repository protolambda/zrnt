package beacon

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/util/hashing"
)

func (state *BeaconStateView) computeProposerIndex(indices []ValidatorIndex, seed Root) (ValidatorIndex, error) {
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
			absI := ValidatorIndex(((i << 5) | j) % uint64(len(indices)))
			shuffledI := PermuteIndex(absI, uint64(len(indices)), seed)
			candidateIndex := indices[int(shuffledI)]
			validator, err := registry.Validator(candidateIndex)
			if err != nil {
				return 0, err
			}
			effectiveBalance, err := validator.EffectiveBalance()
			if err != nil {
				return 0, err
			}
			if effectiveBalance*0xff >= MAX_EFFECTIVE_BALANCE*Gwei(randomByte) {
				return candidateIndex, nil
			}
		}
	}
	return 0, errors.New("random (but balance-biased) infinite scrolling through a committee should always find a proposer")
}

type BoundedIndex struct {
	Index ValidatorIndex
	Activation Epoch
	Exit Epoch
}

func (state *BeaconStateView) loadIndicesBounded() ([]BoundedIndex, error) {
	validators, err := state.Validators()
	if err != nil {
		return nil, err
	}
	valCount, err := validators.Length()
	if err != nil {
		return nil, err
	}
	indicesBounded := make([]BoundedIndex, valCount, valCount)
	valIter := validators.ReadonlyIter()
	i := ValidatorIndex(0)
	for {
		valContainer, ok, err := valIter.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		val, err := AsValidator(valContainer, nil)
		if err != nil {
			return nil, err
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

type EpochsContext struct {
	// TODO: replace with data-sharing lookups for better forking
	Pubkey2Index map[BLSPubkey]ValidatorIndex
	Index2Pubkey []BLSPubkey

	Proposers [SLOTS_PER_EPOCH]ValidatorIndex

	PreviousEpoch *ShufflingEpoch
	CurrentEpoch  *ShufflingEpoch
	NextEpoch     *ShufflingEpoch
}

func (state *BeaconStateView) NewEpochsContext() (*EpochsContext, error) {
	indicesBounded, err := state.loadIndicesBounded()
	if err != nil {
		return nil, err
	}

	valCount := len(indicesBounded)

	epc := &EpochsContext{
		Pubkey2Index: make(map[BLSPubkey]ValidatorIndex, valCount),
		Index2Pubkey: make([]BLSPubkey, 0, valCount),
	}

	if err := epc.syncPubkeys(state); err != nil {
		return nil, err
	}

	slot, err := state.Slot()
	if err != nil {
		return nil, err
	}
	currentEpoch := slot.ToEpoch()
	epc.CurrentEpoch, err = state.ShufflingEpoch(indicesBounded, currentEpoch)
	if err != nil {
		return nil, err
	}
	prevEpoch := currentEpoch.Previous()
	if prevEpoch == currentEpoch { // in case of genesis
		epc.PreviousEpoch = epc.CurrentEpoch
	} else {
		epc.PreviousEpoch, err = state.ShufflingEpoch(indicesBounded, prevEpoch)
		if err != nil {
			return nil, err
		}
	}
	epc.NextEpoch, err = state.ShufflingEpoch(indicesBounded, currentEpoch+1)
	if err != nil {
		return nil, err
	}

	if err := epc.resetProposers(state); err != nil {
		return nil, err
	}

	return epc, nil
}

func (epc *EpochsContext) resetProposers(state *BeaconStateView) error {
	mixes, err := state.RandaoMixes()
	if err != nil {
		return err
	}
	epochSeed, err := mixes.GetSeed(epc.CurrentEpoch.Epoch, DOMAIN_BEACON_ATTESTER)
	if err != nil {
		return err
	}
	slot := epc.CurrentEpoch.Epoch.GetStartSlot()
	hFn := hashing.GetHashFn()
	var buf [32 + 8]byte
	copy(buf[0:32], epochSeed[:])
	for i := Slot(0); i < SLOTS_PER_EPOCH; i++ {
		binary.LittleEndian.PutUint64(buf[32:], uint64(slot))
		seed := hFn(buf[:])
		proposer, err := state.computeProposerIndex(epc.CurrentEpoch.ActiveIndices, seed)
		if err != nil {
			return err
		}
		epc.Proposers[i] = proposer
		slot++
	}
	return nil
}

func (epc *EpochsContext) Copy() *EpochsContext {
	// Go copies are ugly.
	// TODO: replace with immutable datastructure, for cheap copy
	pub2idx := make(map[BLSPubkey]ValidatorIndex, len(epc.Pubkey2Index))
	for k, v := range epc.Pubkey2Index {
		pub2idx[k] = v
	}
	idx2pub := make([]BLSPubkey, len(epc.Index2Pubkey), len(epc.Index2Pubkey))
	for i, v := range epc.Index2Pubkey {
		idx2pub[i] = v
	}
	return &EpochsContext{
		Pubkey2Index: pub2idx,
		Index2Pubkey: idx2pub,
		// Only shallow-copy the other data, it doesn't mutate (only completely replaced on rotation)
		Proposers: epc.Proposers,
		PreviousEpoch: epc.PreviousEpoch,
		CurrentEpoch: epc.CurrentEpoch,
		NextEpoch: epc.NextEpoch,
	}
}

func (epc *EpochsContext) syncPubkeys(state *BeaconStateView) error {
	vals, err := state.Validators()
	if err != nil {
		return err
	}
	valCount, err := vals.Length()
	if err != nil {
		return err
	}
	if epc.Pubkey2Index == nil {
		epc.Pubkey2Index = make(map[BLSPubkey]ValidatorIndex, valCount)
	}
	if epc.Index2Pubkey == nil {
		epc.Index2Pubkey = make([]BLSPubkey, 0, valCount)
	}
	currentCount := uint64(len(epc.Index2Pubkey))
	for i := currentCount; i < valCount; i++ {
		v, err := AsValidator(vals.Get(i))
		if err != nil {
			return err
		}
		pub, err := v.Pubkey()
		if err != nil {
			return err
		}
		idx := ValidatorIndex(i)
		epc.Pubkey2Index[pub] = idx
		epc.Index2Pubkey = append(epc.Index2Pubkey, pub)
	}
	return nil
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
	epc.NextEpoch, err = state.ShufflingEpoch(indicesBounded, nextEpoch)
	if err != nil {
		return err
	}
	return epc.resetProposers(state)
}

func (epc *EpochsContext) getSlotComms(slot Slot) ([][]ValidatorIndex, error) {
	epoch := slot.ToEpoch()
	epochSlot := slot%SLOTS_PER_EPOCH
	if epoch == epc.PreviousEpoch.Epoch {
		return epc.PreviousEpoch.Committees[epochSlot], nil
	} else if epoch == epc.CurrentEpoch.Epoch {
		return epc.CurrentEpoch.Committees[epochSlot], nil
	} else if epoch == epc.NextEpoch.Epoch {
		return epc.NextEpoch.Committees[epochSlot], nil
	} else {
		return nil, fmt.Errorf("crosslink committee retrieval: out of range epoch: %d", epoch)
	}
}

func (epc *EpochsContext) ValCount() uint64 {
	return uint64(len(epc.Index2Pubkey))
}

func (epc *EpochsContext) IsValidIndex(index ValidatorIndex) bool {
	return index < ValidatorIndex(epc.ValCount())
}

func (epc *EpochsContext) Pubkey(index ValidatorIndex) (BLSPubkey, bool) {
	if index < ValidatorIndex(len(epc.Index2Pubkey)) {
		return BLSPubkey{}, false
	}
	return epc.Index2Pubkey[index], true
}

func (epc *EpochsContext) ValidatorIndex(pub BLSPubkey) (ValidatorIndex, bool) {
	idx, ok := epc.Pubkey2Index[pub]
	return idx, ok
}

// Return the beacon committee at slot for index.
func (epc *EpochsContext) GetBeaconCommittee(slot Slot, index CommitteeIndex) ([]ValidatorIndex, error) {
	if index >= MAX_COMMITTEES_PER_SLOT {
		return nil, fmt.Errorf("crosslink committee retrieval: out of range committee index: %d", index)
	}

	slotComms, err := epc.getSlotComms(slot)
	if err != nil {
		return nil, err
	}

	if index >= CommitteeIndex(len(slotComms)) {
		return nil, fmt.Errorf("crosslink committee retrieval: out of range committee index: %d", index)
	}
	return slotComms[index], nil
}

func (epc *EpochsContext) GetCommitteeCountAtSlot(slot Slot) (uint64, error) {
	slotComms, err := epc.getSlotComms(slot)
	return uint64(len(slotComms)), err
}

func (epc *EpochsContext) GetBeaconProposer(slot Slot) (ValidatorIndex, error) {
	epoch := slot.ToEpoch()
	if epoch != epc.CurrentEpoch.Epoch {
		return 0, fmt.Errorf("expected epoch %d for proposer lookup, but lookup was at slot %d (epoch %d)", epc.CurrentEpoch.Epoch, slot, epoch)
	}
	return epc.Proposers[slot % SLOTS_PER_EPOCH], nil
}
