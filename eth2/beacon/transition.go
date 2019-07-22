package beacon

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

type BlockTransition struct {
	BlockMeta
}

func (state *BlockTransition) ProcessBlock(block *BeaconBlock) error {
	if err := state.ProcessHeader(block.Header()); err != nil {
		return err
	}
	body := &block.Body
	if err := state.ProcessRandaoReveal(body.RandaoReveal); err != nil {
		return err
	}
	if err := state.ProcessEth1Vote(body.Eth1Data); err != nil {
		return err
	}
	if err := state.ProcessProposerSlashings(body.ProposerSlashings); err != nil {
		return err
	}
	if err := state.ProcessAttesterSlashings(body.AttesterSlashings); err != nil {
		return err
	}
	if err := state.ProcessDeposits(body.Deposits); err != nil {
		return err
	}
	if err := state.ProcessVoluntaryExits(body.VoluntaryExits); err != nil {
		return err
	}
	if err := state.ProcessTransfers(body.Transfers); err != nil {
		return err
	}
	return nil
}

func (state *BeaconState) ProcessSlot() {
	// Cache latest known state root (for previous slot)
	latestStateRoot := ssz.HashTreeRoot(state, BeaconStateSSZ)

	previousBlockRoot := state.UpdateLatestBlockRoot(latestStateRoot)

	// Cache latest known block and state root
	state.SetRecentRoots(state.Slot, previousBlockRoot, latestStateRoot)
}

type EpochTransition struct {
	EpochMeta
}

func (state *EpochTransition) ProcessEpoch() {
	state.ProcessEpochJustification()
	state.ProcessEpochCrosslinks()
	state.ProcessEpochRewardsAndPenalties()
	state.ProcessEpochRegistryUpdates()
	state.ProcessEpochSlashings()
	state.ProcessEpochFinalUpdates()
}

// TODO rewardpenalty feature package
func (state *BeaconState) ProcessEpochRewardsAndPenalties() {
	if state.Epoch() == GENESIS_EPOCH {
		return
	}
	sum := NewDeltas(state.ValidatorCount())
	sum.Add(state.AttestationDeltas(state))
	sum.Add(state.CrosslinksDeltas(state))
	state.ApplyDeltas(sum)
}

// TODO final feature package
func (state *BeaconState) ProcessEpochFinalUpdates() {
	nextEpoch := state.Epoch() + 1

	// Reset eth1 data votes if it is the end of the voting period.
	if (state.Slot+1)%SLOTS_PER_ETH1_VOTING_PERIOD == 0 {
		state.ResetEth1Votes()
	}

	state.UpdateEffectiveBalances()
	state.RotateStartShard(state)
	state.UpdateActiveIndexRoot(nextEpoch + ACTIVATION_EXIT_DELAY)
	state.UpdateCompactCommitteesRoot(nextEpoch)
	state.ResetSlashings(nextEpoch)
	state.PrepareRandao(nextEpoch)

	// Set historical root accumulator
	if nextEpoch%SLOTS_PER_HISTORICAL_ROOT.ToEpoch() == 0 {
		state.UpdateHistoricalRoots()
	}

	state.RotateEpochAttestations()
}


// Process the state to the given slot.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
func (state *BeaconState) ProcessSlots(slot Slot) error {
	if state.Slot > slot {
		return errors.New("cannot transition from pre-state with higher slot than transition target")
	}
	// happens at the start of every CurrentSlot
	for state.Slot < slot {
		state.ProcessSlot()
		// Per-epoch transition happens at the start of the first slot of every epoch.
		if (state.Slot+1)%SLOTS_PER_EPOCH == 0 {
			state.ProcessEpoch()
		}
		state.Slot++
	}
	return nil
}

// Transition the state to the slot of the given block, then processes the block.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
func (state *BeaconState) StateTransition(block *BeaconBlock, verifyStateRoot bool) error {
	if err := state.ProcessSlots(block.Slot); err != nil {
		return err
	}

	if err := state.ProcessBlock(block); err != nil {
		return err
	}
	// State root verification
	if verifyStateRoot && block.StateRoot != ssz.HashTreeRoot(state, BeaconStateSSZ) {
		return errors.New("block has invalid state root")
	}
	return nil
}