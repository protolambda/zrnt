package beacon

import (
	"errors"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/ztyp/tree"
)

func (state *BeaconStateView) ProcessSlot() error {
	// Cache state root
	previousStateRoot := state.HashTreeRoot(tree.GetHashFn())

	stateRootsBatch, err := state.StateRoots()
	if err != nil {
		return err
	}
	slot, err := state.Slot()
	if err != nil {
		return err
	}
	if err := stateRootsBatch.SetRoot(slot, previousStateRoot); err != nil {
		return err
	}

	latestHeader, err := state.LatestBlockHeader()
	if err != nil {
		return err
	}
	stateRoot, err := latestHeader.StateRoot()
	if err != nil {
		return err
	}
	if stateRoot == (Root{}) {
		if err := latestHeader.SetStateRoot(previousStateRoot); err != nil {
			return err
		}
	}
	previousBlockRoot := latestHeader.HashTreeRoot(tree.GetHashFn())

	// Cache latest known block and state root
	blockRootsBatch, err := state.BlockRoots()
	if err != nil {
		return err
	}
	if err := blockRootsBatch.SetRoot(slot, previousBlockRoot); err != nil {
		return err
	}

	return nil
}

func (state *BeaconStateView) ProcessEpoch(epc *EpochsContext) error {
	process, err := state.PrepareEpochProcess(epc)
	if err != nil {
		return err
	}
	if err := state.ProcessEpochJustification(epc, process); err != nil {
		return err
	}
	if err := state.ProcessEpochRewardsAndPenalties(epc, process); err != nil {
		return err
	}
	if err := state.ProcessEpochRegistryUpdates(epc, process); err != nil {
		return err
	}
	if err := state.ProcessEpochSlashings(epc, process); err != nil {
		return err
	}
	if err := state.ProcessEpochFinalUpdates(epc, process); err != nil {
		return err
	}
	return nil
}

// Process the state to the given slot.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
func (state *BeaconStateView) ProcessSlots(epc *EpochsContext, slot Slot) error {
	// happens at the start of every CurrentSlot
	currentSlot, err := state.Slot()
	if err != nil {
		return err
	}
	if currentSlot > slot {
		return errors.New("cannot transition from pre-state with higher or equal slot than transition target")
	}
	for currentSlot < slot {
		if err := state.ProcessSlot(); err != nil {
			return err
		}
		// Per-epoch transition happens at the start of the first slot of every epoch.
		// (with the slot still at the end of the last epoch)
		isEpochEnd := (currentSlot + 1).ToEpoch() != currentSlot.ToEpoch()
		if isEpochEnd {
			if err := state.ProcessEpoch(epc); err != nil {
				return err
			}
		}
		currentSlot += 1
		if err := state.SetSlot(currentSlot); err != nil {
			return err
		}
		if isEpochEnd {
			if err := epc.RotateEpochs(state); err != nil {
				return err
			}
		}
	}
	return nil
}

func (state *BeaconStateView) ProcessBlock(epc *EpochsContext, block *BeaconBlock) error {
	if err := state.ProcessHeader(epc, block); err != nil {
		return err
	}
	body := &block.Body
	if err := state.ProcessRandaoReveal(epc, body.RandaoReveal); err != nil {
		return err
	}
	if err := state.ProcessEth1Vote(epc, body.Eth1Data); err != nil {
		return err
	}
	if err := state.ProcessProposerSlashings(epc, body.ProposerSlashings); err != nil {
		return err
	}
	if err := state.ProcessAttesterSlashings(epc, body.AttesterSlashings); err != nil {
		return err
	}
	if err := state.ProcessAttestations(epc, body.Attestations); err != nil {
		return err
	}
	if err := state.ProcessDeposits(epc, body.Deposits); err != nil {
		return err
	}
	if err := state.ProcessVoluntaryExits(epc, body.VoluntaryExits); err != nil {
		return err
	}
	return nil
}

// Transition the state to the slot of the given block, then processes the block.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
//
func (state *BeaconStateView) StateTransition(epc *EpochsContext, block *SignedBeaconBlock, validateResult bool) error {
	if err := state.ProcessSlots(epc, block.Message.Slot); err != nil {
		return err
	}
	if validateResult {
		// Safe to ignore proposer index, it will be checked as part of the ProcessHeader call.
		if !state.VerifySignature(epc, block, false) {
			return errors.New("block has invalid signature")
		}
	}

	if err := state.ProcessBlock(epc, &block.Message); err != nil {
		return err
	}

	// State root verification
	if validateResult && block.Message.StateRoot != state.HashTreeRoot(tree.GetHashFn()) {
		return errors.New("block has invalid state root")
	}
	return nil
}

// Assuming the slot is valid, and optionally assume the proposer index is valid, check if the signature is valid
func (state *BeaconStateView) VerifySignature(epc *EpochsContext, block *SignedBeaconBlock, validateProposerIndex bool) bool {
	if validateProposerIndex {
		proposerIndex, err := epc.GetBeaconProposer(block.Message.Slot)
		if err != nil {
			return false
		}
		if proposerIndex != block.Message.ProposerIndex {
			return false
		}
	}
	pub, ok := epc.PubkeyCache.Pubkey(block.Message.ProposerIndex)
	if !ok {
		return false
	}
	domain, err := state.GetDomain(DOMAIN_BEACON_PROPOSER, block.Message.Slot.ToEpoch())
	if err != nil {
		return false
	}
	return bls.Verify(pub, ComputeSigningRoot(block.Message.HashTreeRoot(), domain), block.Signature)
}
