package beacon

import (
	"errors"
	"github.com/protolambda/ztyp/tree"
)

func (state *BeaconStateView) ProcessSlot() error {
	// Cache latest known state root (for previous slot)
	latestStateRoot := f.Meta.StateRoot()

	if err := f.Meta.UpdateLatestBlockStateRoot(latestStateRoot); err != nil {
		return err
	}

	previousBlockRoot, err := f.Meta.GetLatestBlockRoot()
	if err != nil {
		return err
	}

	currentSlot, err := f.Meta.CurrentSlot()
	if err != nil {
		return err
	}

	// Cache latest known block and state root
	if err := state.SetRecentRoots(currentSlot, previousBlockRoot, latestStateRoot); err != nil {
		return err
	}

	return nil
}

func (state *BeaconStateView) ProcessEpoch() error {
	if err := state.ProcessEpochJustification(); err != nil {
		return err
	}
	if err := state.ProcessEpochRewardsAndPenalties(); err != nil {
		return err
	}
	if err := state.ProcessEpochRegistryUpdates(); err != nil {
		return err
	}
	if err := state.ProcessEpochSlashings(); err != nil {
		return err
	}
	if err := state.ProcessEpochFinalUpdates(); err != nil {
		return err
	}
	return nil
}

// Process the state to the given slot.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
func (state *BeaconStateView) ProcessSlots(slot Slot) error {
	// happens at the start of every CurrentSlot
	currentSlot, err := state.Slot()
	if err != nil {
		return err
	}
	for currentSlot < slot {
		state.ProcessSlot()
		// Per-epoch transition happens at the start of the first slot of every epoch.
		// (with the slot still at the end of the last epoch)
		isEpochEnd := (currentSlot + 1).ToEpoch() != currentSlot.ToEpoch()
		if isEpochEnd {
			state.ProcessEpoch()
		}
		state.IncrementSlot()
		currentSlot += 1
	}
}

func (state *BeaconStateView) ProcessBlock(block *BeaconBlock) error {
	if err := state.ProcessHeader(block); err != nil {
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
	if err := state.ProcessAttestations(body.Attestations); err != nil {
		return err
	}
	if err := state.ProcessDeposits(body.Deposits); err != nil {
		return err
	}
	if err := state.ProcessVoluntaryExits(body.VoluntaryExits); err != nil {
		return err
	}
	return nil
}

// Transition the state to the slot of the given block, then processes the block.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
//
func (state *BeaconStateView) StateTransition(block *SignedBeaconBlock, validateResult bool) error {
	if input.CurrentSlot() > block.Message.Slot {
		return errors.New("cannot transition from pre-state with higher slot than transition target")
	}
	ProcessSlots(input, block.Message.Slot)
	if validateResult {
		fork, err := state.Fork()
		if err != nil {
			return err
		}
		currentVersion, err := fork.CurrentVersion()
		if err != nil {
			return err
		}
		genValRoot, err := state.GenesisValidatorsRoot()
		if err != nil {
			return err
		}
		if !block.VerifySignature(block.Message.ProposerIndex, currentVersion, genValRoot) {
			return errors.New("block has invalid signature")
		}
	}

	if err := state.ProcessBlock(&block.Message); err != nil {
		return err
	}

	// State root verification
	if validateResult && block.Message.StateRoot == state.HashTreeRoot(tree.GetHashFn()) {
		return errors.New("block has invalid state root")
	}
	return nil
}

func (state *BeaconStateView) VerifySignature(pubkey BLSPubkey, version Version, genValRoot Root) bool {
	return bls.Verify(
		pubkey,
		ComputeSigningRoot(
			f.BlockRoot(),
			ComputeDomain(DOMAIN_BEACON_PROPOSER, version, genValRoot)),
		f.Signature())
}

