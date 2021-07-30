package common

import (
	"fmt"
	"github.com/protolambda/zrnt/eth2/util/math"
)

type IndexedSyncCommittee struct {
	CachedPubkeys []*CachedPubkey
	Indices       []ValidatorIndex
}

type EpochsContext struct {
	Spec *Spec

	// ValidatorPubkeyCache may be replaced when a new forked-out cache takes over to process an alternative Eth1 deposit chain.
	ValidatorPubkeyCache *PubkeyCache
	Proposers            *ProposersEpoch

	BuilderPubkeyCache *BuilderPubkeyCache

	PreviousEpoch *ShufflingEpoch
	CurrentEpoch  *ShufflingEpoch
	NextEpoch     *ShufflingEpoch

	// nil for pre-altair chain
	CurrentSyncCommittee *IndexedSyncCommittee
	NextSyncCommittee    *IndexedSyncCommittee

	// TODO: track active effective balances
	// TODO: track total active stake
	// Effective balances of all validators at the start of the epoch.
	EffectiveBalances []Gwei
	// Total effective balance of the active validators at the start of the epoch.
	TotalActiveStake Gwei
	// cached integer square root of TotalActiveStake
	TotalActiveStakeSqRoot Gwei
}

// NewEpochsContext constructs a new context for the processing of the current epoch.
func NewEpochsContext(spec *Spec, state BeaconState) (*EpochsContext, error) {
	vals, err := state.Validators()
	if err != nil {
		return nil, err
	}
	pc, err := NewPubkeyCache(vals)
	if err != nil {
		return nil, err
	}

	// nil if no builder data in the state (pre-sharding)
	var bpc *BuilderPubkeyCache
	if builderState, ok := state.(BuilderBeaconState); ok {
		builders, err := builderState.BlobBuilders()
		if err != nil {
			return nil, err
		}
		bpc, err = NewBuilderPubkeyCache(builders)
	}

	epc := &EpochsContext{
		Spec:                 spec,
		ValidatorPubkeyCache: pc,
		BuilderPubkeyCache: bpc,
	}
	if err := epc.LoadShuffling(state); err != nil {
		return nil, err
	}
	if err := epc.LoadProposers(state); err != nil {
		return nil, err
	}
	if syncState, ok := state.(SyncCommitteeBeaconState); ok {
		if err := epc.LoadSyncCommittees(syncState); err != nil {
			return nil, err
		}
	}
	return epc, nil
}

func (epc *EpochsContext) LoadShuffling(state BeaconState) error {
	vals, err := state.Validators()
	if err != nil {
		return err
	}
	indicesBounded, err := LoadBoundedIndices(vals)
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
	if err := epc.loadCurrentStake(state, indicesBounded); err != nil {
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

func (epc *EpochsContext) loadCurrentStake(state BeaconState, indicesBounded []BoundedIndex) error {
	epc.EffectiveBalances = make([]Gwei, len(indicesBounded), len(indicesBounded))
	epc.TotalActiveStake = 0
	vals, err := state.Validators()
	if err != nil {
		return err
	}
	currentEpoch := epc.CurrentEpoch.Epoch
	for i, v := range indicesBounded {
		// TODO: optimize effective balance retrieval
		val, err := vals.Validator(ValidatorIndex(i))
		if err != nil {
			return err
		}
		eff, err := val.EffectiveBalance()
		if err != nil {
			return err
		}
		epc.EffectiveBalances[i] = eff
		if v.Activation <= currentEpoch && currentEpoch < v.Exit {
			epc.TotalActiveStake += eff
		}
	}
	if epc.TotalActiveStake < epc.Spec.EFFECTIVE_BALANCE_INCREMENT {
		epc.TotalActiveStake = epc.Spec.EFFECTIVE_BALANCE_INCREMENT
	}
	epc.TotalActiveStakeSqRoot = Gwei(math.IntegerSquareroot(uint64(epc.TotalActiveStake)))
	return nil
}

func (epc *EpochsContext) LoadProposers(state BeaconState) error {
	// prerequisite to load shuffling: the list of active indices, same as in the shuffling. So load the shuffling first.
	if epc.CurrentEpoch == nil {
		if err := epc.LoadShuffling(state); err != nil {
			return err
		}
	}
	props, err := ComputeProposers(epc.Spec, state, epc.CurrentEpoch.Epoch, epc.CurrentEpoch.ActiveIndices)
	if err != nil {
		return err
	}
	epc.Proposers = props
	return nil
}

func (epc *EpochsContext) LoadSyncCommittees(state SyncCommitteeBeaconState) error {
	current, err := state.CurrentSyncCommittee()
	if err != nil {
		return fmt.Errorf("failed to get current sync committee to hydrate EPC")
	}
	epc.CurrentSyncCommittee, err = epc.hydrateSyncCommittee(current)
	if err != nil {
		return fmt.Errorf("failed to hydrate current sync committee in EPC")
	}
	next, err := state.NextSyncCommittee()
	if err != nil {
		return fmt.Errorf("failed to get next sync committee to hydrate EPC")
	}
	epc.NextSyncCommittee, err = epc.hydrateSyncCommittee(next)
	if err != nil {
		return fmt.Errorf("failed to hydrate next sync committee in EPC")
	}
	return nil
}

func (epc *EpochsContext) Clone() *EpochsContext {
	// All fields can be reused, just need a fresh shallow copy of the outer container
	epcClone := *epc
	return &epcClone
}

func (epc *EpochsContext) RotateEpochs(state BeaconState) error {
	epc.PreviousEpoch = epc.CurrentEpoch
	epc.CurrentEpoch = epc.NextEpoch
	nextEpoch := epc.CurrentEpoch.Epoch + 1
	vals, err := state.Validators()
	if err != nil {
		return err
	}
	// TODO: could use epoch-transition processing validator data to not read state here
	indicesBounded, err := LoadBoundedIndices(vals)
	if err != nil {
		return err
	}
	epc.NextEpoch, err = ComputeShufflingEpoch(epc.Spec, state, indicesBounded, nextEpoch)
	if err != nil {
		return err
	}
	if err := epc.LoadProposers(state); err != nil {
		return err
	}
	if err := epc.loadCurrentStake(state, indicesBounded); err != nil {
		return err
	}
	if syncState, ok := state.(SyncCommitteeBeaconState); ok {
		// if the state has a list of sync committee pubkeys, we want to cache the indices of that sync committee
		if epc.CurrentEpoch.Epoch%epc.Spec.EPOCHS_PER_SYNC_COMMITTEE_PERIOD == 0 {
			// just got into the epoch, we just need to re-hydrate the EPC
			if epc.NextSyncCommittee != nil {
				epc.CurrentSyncCommittee = epc.NextSyncCommittee
			} else {
				current, err := syncState.CurrentSyncCommittee()
				if err != nil {
					return fmt.Errorf("failed to get current sync committee to hydrate EPC")
				}
				epc.CurrentSyncCommittee, err = epc.hydrateSyncCommittee(current)
				if err != nil {
					return fmt.Errorf("failed to hydrate current sync committee in EPC")
				}
			}
			next, err := syncState.NextSyncCommittee()
			if err != nil {
				return fmt.Errorf("failed to get next sync committee to hydrate EPC")
			}
			epc.NextSyncCommittee, err = epc.hydrateSyncCommittee(next)
			if err != nil {
				return fmt.Errorf("failed to hydrate next sync committee in EPC")
			}
		}
	}
	return nil
}

func (epc *EpochsContext) hydrateSyncCommittee(view *SyncCommitteeView) (*IndexedSyncCommittee, error) {
	pubsView, err := view.Pubkeys()
	if err != nil {
		return nil, err
	}
	pubs, err := pubsView.Flatten()
	if err != nil {
		return nil, err
	}
	indices := make([]ValidatorIndex, len(pubs), len(pubs))
	cachedPubs := make([]*CachedPubkey, len(pubs), len(pubs))
	for i := 0; i < len(indices); i++ {
		idx, ok := epc.ValidatorPubkeyCache.ValidatorIndex(pubs[i])
		if !ok {
			return nil, fmt.Errorf("missing validator index for pubkey %d (%s) of sync committee", i, pubs[i])
		}
		indices[i] = idx
		cachedPubkey, ok := epc.ValidatorPubkeyCache.Pubkey(idx)
		if !ok {
			return nil, fmt.Errorf("pubkey cache is inconsistent, sync committee member with validator index %d has no cached pubkey", idx)
		}
		cachedPubs[i] = cachedPubkey
	}
	return &IndexedSyncCommittee{
		CachedPubkeys: cachedPubs,
		Indices:       indices,
	}, nil
}

func (epc *EpochsContext) getSlotComms(slot Slot) ([][]ValidatorIndex, error) {
	epochSlot := slot % epc.Spec.SLOTS_PER_EPOCH
	epoch := epc.Spec.SlotToEpoch(slot)
	comms, err := epc.getEpochComms(epoch)
	if err != nil {
		return nil, err
	}
	return comms[epochSlot], nil
}

func (epc *EpochsContext) getEpochComms(epoch Epoch) ([][][]ValidatorIndex, error) {
	if epoch == epc.PreviousEpoch.Epoch {
		return epc.PreviousEpoch.Committees, nil
	} else if epoch == epc.CurrentEpoch.Epoch {
		return epc.CurrentEpoch.Committees, nil
	} else if epoch == epc.NextEpoch.Epoch {
		return epc.NextEpoch.Committees, nil
	} else {
		return nil, fmt.Errorf("beacon committee retrieval: out of range epoch: %d", epoch)
	}
}

// Return the beacon committee at slot for index.
func (epc *EpochsContext) GetBeaconCommittee(slot Slot, index CommitteeIndex) ([]ValidatorIndex, error) {
	if index >= CommitteeIndex(epc.Spec.MAX_COMMITTEES_PER_SLOT) {
		return nil, fmt.Errorf("beacon committee retrieval: out of range committee index: %d", index)
	}

	slotComms, err := epc.getSlotComms(slot)
	if err != nil {
		return nil, err
	}

	if index >= CommitteeIndex(len(slotComms)) {
		return nil, fmt.Errorf("beacon committee retrieval: out of range committee index: %d", index)
	}
	return slotComms[index], nil
}

func (epc *EpochsContext) GetCommitteeCountPerSlot(epoch Epoch) (uint64, error) {
	epochComms, err := epc.getEpochComms(epoch)
	return uint64(len(epochComms[0])), err
}

func (epc *EpochsContext) GetBeaconProposer(slot Slot) (ValidatorIndex, error) {
	return epc.Proposers.GetBeaconProposer(slot)
}

func (epc *EpochsContext) GetShardProposer(slot Slot, shard Shard) (ValidatorIndex, error) {
	// TODO
	return 0, nil
}

func (epc *EpochsContext) StartShard(slot Slot) (Shard, error) {
	// TODO
	return 0, nil
}

func (epc *EpochsContext) ComputeShardFromCommitteeIndex(index CommitteeIndex) (Shard, error) {
	// TODO
	return 0, nil
}
