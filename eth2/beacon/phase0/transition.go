package phase0

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/ztyp/tree"
)

func ProcessEpoch(ctx context.Context, spec *common.Spec, epc *EpochsContext, state *BeaconStateView) error {
	process, err := PrepareEpochProcess(ctx, spec, epc, state)
	if err != nil {
		return err
	}
	if err := ProcessEpochJustification(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := ProcessEpochRewardsAndPenalties(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := ProcessEpochRegistryUpdates(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := ProcessEpochSlashings(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := ProcessEffectiveBalanceUpdates(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := ProcessEth1DataReset(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := ProcessSlashingsReset(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := ProcessRandaoMixesReset(ctx, spec, epc, process, state); err != nil {
		return err
	}
	if err := ProcessParticipationRecordUpdates(ctx, spec, epc, process, state); err != nil {
		return err
	}
	return nil
}

func ProcessBlock(ctx context.Context, spec *common.Spec, epc *EpochsContext, state *BeaconStateView, block *BeaconBlock) error {
	slot, err := state.Slot()
	if err != nil {
		return err
	}
	proposerIndex, err := epc.GetBeaconProposer(slot)
	if err != nil {
		return err
	}
	if err := common.ProcessHeader(ctx, spec, state, block.Header(spec), proposerIndex); err != nil {
		return err
	}
	body := &block.Body
	if err := ProcessRandaoReveal(ctx, spec, epc, state, body.RandaoReveal); err != nil {
		return err
	}
	if err := ProcessEth1Vote(ctx, spec, epc, state, body.Eth1Data); err != nil {
		return err
	}
	// Safety checks, in case the user of the function provided too many operations
	if err := body.CheckLimits(spec); err != nil {
		return err
	}

	if err := ProcessProposerSlashings(ctx, spec, epc, state, body.ProposerSlashings); err != nil {
		return err
	}
	if err := ProcessAttesterSlashings(ctx, spec, epc, state, body.AttesterSlashings); err != nil {
		return err
	}
	if err := ProcessAttestations(ctx, spec, epc, state, body.Attestations); err != nil {
		return err
	}
	if err := ProcessDeposits(ctx, spec, epc, state, body.Deposits); err != nil {
		return err
	}
	if err := ProcessVoluntaryExits(ctx, spec, epc, state, body.VoluntaryExits); err != nil {
		return err
	}
	return nil
}

// Assuming the slot is valid, and optionally assume the proposer index is valid, check if the signature is valid
func VerifyBlockSignature(spec *common.Spec, epc *EpochsContext, state common.BeaconState, block *SignedBeaconBlock, validateProposerIndex bool) bool {
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
	domain, err := common.GetDomain(state, spec.DOMAIN_BEACON_PROPOSER, spec.SlotToEpoch(block.Message.Slot))
	if err != nil {
		return false
	}
	return bls.Verify(pub, common.ComputeSigningRoot(block.Message.HashTreeRoot(spec, tree.GetHashFn()), domain), block.Signature)
}
