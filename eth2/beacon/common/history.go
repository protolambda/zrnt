package common

// Return the block root at a recent slot. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func GetBlockRootAtSlot(spec *Spec, state BeaconState, slot Slot) (Root, error) {
	blockRoots, err := state.BlockRoots()
	if err != nil {
		return Root{}, err
	}
	return blockRoots.GetRoot(slot)
}

// Return the block root at a recent epoch. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func GetBlockRoot(spec *Spec, state BeaconState, epoch Epoch) (Root, error) {
	blockRoots, err := state.BlockRoots()
	if err != nil {
		return Root{}, err
	}
	startSlot, err := spec.EpochStartSlot(epoch)
	if err != nil {
		return Root{}, err
	}
	return blockRoots.GetRoot(startSlot)
}
