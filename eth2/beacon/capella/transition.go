package capella

import (
	"bytes"
	"context"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/tree"
)

func (state *BeaconStateView) ProcessEpoch(ctx context.Context, spec *common.Spec, epc *common.EpochsContext) error {
	vals, err := state.Validators()
	if err != nil {
		return err
	}
	flats, err := common.FlattenValidators(vals)
	if err != nil {
		return err
	}
	attesterData, err := altair.ComputeEpochAttesterData(ctx, spec, epc, flats, state)
	if err != nil {
		return err
	}
	just := phase0.JustificationStakeData{
		CurrentEpoch:                  epc.CurrentEpoch.Epoch,
		TotalActiveStake:              epc.TotalActiveStake,
		PrevEpochUnslashedTargetStake: attesterData.PrevEpochUnslashedStake.TargetStake,
		CurrEpochUnslashedTargetStake: attesterData.CurrEpochUnslashedTargetStake,
	}
	if err := phase0.ProcessEpochJustification(ctx, spec, &just, state); err != nil {
		return err
	}
	if err := altair.ProcessInactivityUpdates(ctx, spec, attesterData, state); err != nil {
		return err
	}
	if err := altair.ProcessEpochRewardsAndPenalties(ctx, spec, epc, attesterData, state); err != nil {
		return err
	}
	if err := phase0.ProcessEpochRegistryUpdates(ctx, spec, epc, flats, state); err != nil {
		return err
	}
	// phase0 implementation, but with fork-logic, will account for changed slashing multiplier
	if err := phase0.ProcessEpochSlashings(ctx, spec, epc, flats, state); err != nil {
		return err
	}
	if err := phase0.ProcessEth1DataReset(ctx, spec, epc, state); err != nil {
		return err
	}
	if err := phase0.ProcessEffectiveBalanceUpdates(ctx, spec, epc, flats, state); err != nil {
		return err
	}
	if err := phase0.ProcessSlashingsReset(ctx, spec, epc, state); err != nil {
		return err
	}
	if err := phase0.ProcessRandaoMixesReset(ctx, spec, epc, state); err != nil {
		return err
	}
	if err := phase0.ProcessHistoricalRootsUpdate(ctx, spec, epc, state); err != nil {
		return err
	}
	if err := altair.ProcessParticipationFlagUpdates(ctx, spec, state); err != nil {
		return err
	}
	if err := altair.ProcessSyncCommitteeUpdates(ctx, spec, epc, state); err != nil {
		return err
	}
	return nil
}

func (state *BeaconStateView) ProcessBlock(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, benv *common.BeaconBlockEnvelope) error {
	body, ok := benv.Body.(*BeaconBlockBody)
	if !ok {
		return fmt.Errorf("unexpected block type %T in Bellatrix ProcessBlock", benv.Body)
	}
	expectedProposer, err := epc.GetBeaconProposer(benv.Slot)
	if err != nil {
		return err
	}
	if err := common.ProcessHeader(ctx, spec, state, &benv.BeaconBlockHeader, expectedProposer); err != nil {
		return err
	}
	block := &BeaconBlock{
		Slot:          benv.Slot,
		ProposerIndex: benv.ProposerIndex,
		ParentRoot:    benv.ParentRoot,
		StateRoot:     benv.StateRoot,
		Body:          *body,
	}
	if enabled, err := state.IsExecutionEnabled(spec, block); err != nil {
		return err
	} else if enabled {
		if err := ProcessWithdrawals(ctx, spec, state, &body.ExecutionPayload); err != nil {
			return err
		}
		if err := ProcessExecutionPayload(ctx, spec, state, &body.ExecutionPayload, spec.ExecutionEngine); err != nil {
			return err
		}
	}
	if err := phase0.ProcessRandaoReveal(ctx, spec, epc, state, body.RandaoReveal); err != nil {
		return err
	}
	if err := phase0.ProcessEth1Vote(ctx, spec, epc, state, body.Eth1Data); err != nil {
		return err
	}
	// Safety checks, in case the user of the function provided too many operations
	if err := body.CheckLimits(spec); err != nil {
		return err
	}

	if err := phase0.ProcessProposerSlashings(ctx, spec, epc, state, body.ProposerSlashings); err != nil {
		return err
	}
	if err := phase0.ProcessAttesterSlashings(ctx, spec, epc, state, body.AttesterSlashings); err != nil {
		return err
	}
	if err := altair.ProcessAttestations(ctx, spec, epc, state, body.Attestations); err != nil {
		return err
	}
	// Note: state.AddValidator changed in Altair, but the deposit processing itself stayed the same.
	if err := phase0.ProcessDeposits(ctx, spec, epc, state, body.Deposits); err != nil {
		return err
	}
	if err := phase0.ProcessVoluntaryExits(ctx, spec, epc, state, body.VoluntaryExits); err != nil {
		return err
	}
	if err := ProcessBLSToExecutionChanges(ctx, spec, epc, state, body.BLSToExecutionChanges); err != nil {
		return err
	}
	if err := altair.ProcessSyncAggregate(ctx, spec, epc, state, &body.SyncAggregate); err != nil {
		return err
	}
	return nil
}

type ExecutionUpgradeBeaconState interface {
	IsExecutionEnabled(spec *common.Spec, block *BeaconBlock) (bool, error)
	IsTransitionCompleted() (bool, error)
	IsTransitionBlock(spec *common.Spec, block *BeaconBlock) (bool, error)
}

type ExecutionTrackingBeaconState interface {
	common.BeaconState

	LatestExecutionPayloadHeader() (*ExecutionPayloadHeaderView, error)
	SetLatestExecutionPayloadHeader(h *ExecutionPayloadHeader) error
}

func (state *BeaconStateView) IsExecutionEnabled(spec *common.Spec, block *BeaconBlock) (bool, error) {
	isTransitionCompleted, err := state.IsTransitionCompleted()
	if err != nil {
		return false, err
	}
	if isTransitionCompleted {
		return true, nil
	}
	return state.IsTransitionBlock(spec, block)
}

func (state *BeaconStateView) IsTransitionCompleted() (bool, error) {
	execHeader, err := state.LatestExecutionPayloadHeader()
	if err != nil {
		return false, err
	}
	empty := ExecutionPayloadHeaderType.DefaultNode().MerkleRoot(tree.GetHashFn())
	return execHeader.HashTreeRoot(tree.GetHashFn()) != empty, nil
}

func (state *BeaconStateView) IsTransitionBlock(spec *common.Spec, block *BeaconBlock) (bool, error) {
	isTransitionCompleted, err := state.IsTransitionCompleted()
	if err != nil {
		return false, err
	}
	if isTransitionCompleted {
		return false, nil
	}
	empty := ExecutionPayloadType(spec).DefaultNode().MerkleRoot(tree.GetHashFn())
	return block.Body.ExecutionPayload.HashTreeRoot(spec, tree.GetHashFn()) != empty, nil
}

func HasEth1WithdrawalCredential(validator common.Validator) bool {
	withdrawalCredentials, err := validator.WithdrawalCredentials()
	if err != nil {
		panic(err)
	}
	return bytes.Equal(withdrawalCredentials[:1], []byte{common.ETH1_ADDRESS_WITHDRAWAL_PREFIX})
}

func Eth1WithdrawalCredential(validator common.Validator) common.Eth1Address {
	withdrawalCredentials, err := validator.WithdrawalCredentials()
	if err != nil {
		panic(err)
	}
	var address common.Eth1Address
	copy(address[:], withdrawalCredentials[12:])
	return address
}

func IsFullyWithdrawableValidator(validator common.Validator, balance common.Gwei, epoch common.Epoch) bool {
	withdrawableEpoch, err := validator.WithdrawableEpoch()
	if err != nil {
		panic(err)
	}
	return HasEth1WithdrawalCredential(validator) && withdrawableEpoch <= epoch && balance > 0
}

func IsPartiallyWithdrawableValidator(spec *common.Spec, validator common.Validator, balance common.Gwei, epoch common.Epoch) bool {
	effectiveBalance, err := validator.EffectiveBalance()
	if err != nil {
		panic(err)
	}
	hasMaxEffectiveBalance := effectiveBalance == spec.MAX_EFFECTIVE_BALANCE
	hasExcessBalance := balance > spec.MAX_EFFECTIVE_BALANCE
	return HasEth1WithdrawalCredential(validator) && hasMaxEffectiveBalance && hasExcessBalance
}

func (state *BeaconStateView) GetExpectedWithdrawals(spec *common.Spec) (common.Withdrawals, error) {
	slot, err := state.Slot()
	if err != nil {
		return nil, err
	}
	epoch := spec.SlotToEpoch(slot)
	withdrawalIndex, err := state.NextWithdrawalIndex()
	if err != nil {
		return nil, err
	}
	validatorIndex, err := state.NextWithdrawalValidatorIndex()
	if err != nil {
		return nil, err
	}
	validators, err := state.Validators()
	if err != nil {
		return nil, err
	}
	validatorCount, err := validators.ValidatorCount()
	if err != nil {
		return nil, err
	}
	balances, err := state.Balances()
	if err != nil {
		return nil, err
	}
	withdrawals := make(common.Withdrawals, 0)
	var i uint64 = 0
	for {
		validator, err := validators.Validator(validatorIndex)
		if err != nil {
			return nil, err
		}
		balance, err := balances.GetBalance(validatorIndex)
		if err != nil {
			return nil, err
		}
		if i >= validatorCount || i >= uint64(spec.MAX_VALIDATORS_PER_WITHDRAWALS_SWEEP) {
			break
		}
		if IsFullyWithdrawableValidator(validator, balance, epoch) {
			withdrawals = append(withdrawals, common.Withdrawal{
				Index:          withdrawalIndex,
				ValidatorIndex: validatorIndex,
				Address:        Eth1WithdrawalCredential(validator),
				Amount:         balance,
			})
			withdrawalIndex += 1
		} else if IsPartiallyWithdrawableValidator(spec, validator, balance, epoch) {
			withdrawals = append(withdrawals, common.Withdrawal{
				Index:          withdrawalIndex,
				ValidatorIndex: validatorIndex,
				Address:        Eth1WithdrawalCredential(validator),
				Amount:         balance - spec.MAX_EFFECTIVE_BALANCE,
			})
			withdrawalIndex += 1
		}
		validatorIndex = common.ValidatorIndex(uint64(validatorIndex+1) % validatorCount)
		i += 1
	}
	return withdrawals, nil
}

func ProcessWithdrawals(ctx context.Context, spec *common.Spec, state *BeaconStateView, executionPayload *ExecutionPayload) error {
	expectedWithdrawals, err := state.GetExpectedWithdrawals(spec)
	if err != nil {
		return err
	}
	if len(expectedWithdrawals) != len(executionPayload.Withdrawals) {
		return fmt.Errorf("unexpected number of withdrawals in Capella ProcessWithdrawals: want=%d, got=%d", len(expectedWithdrawals), len(executionPayload.Withdrawals))
	}
	bals, err := state.Balances()
	if err != nil {
		return err
	}
	for w := 0; w < len(expectedWithdrawals); w++ {
		withdrawal := executionPayload.Withdrawals[w]
		expectedWithdrawal := expectedWithdrawals[w]
		if withdrawal.Index != expectedWithdrawal.Index ||
			withdrawal.ValidatorIndex != expectedWithdrawal.ValidatorIndex ||
			!bytes.Equal(withdrawal.Address[:], expectedWithdrawal.Address[:]) ||
			withdrawal.Amount != expectedWithdrawal.Amount {
			return fmt.Errorf("unexpected withdrawal in Capella ProcessWithdrawals: want=%s, got=%s", expectedWithdrawal, withdrawal)
		}
		common.DecreaseBalance(bals, expectedWithdrawal.ValidatorIndex, expectedWithdrawal.Amount)
	}
	if len(expectedWithdrawals) > 0 {
		latestWithdrawal := expectedWithdrawals[len(expectedWithdrawals)-1]
		state.SetNextWithdrawalIndex(latestWithdrawal.Index + 1)
	}
	validators, err := state.Validators()
	if err != nil {
		return err
	}
	validatorCount, err := validators.ValidatorCount()
	if err != nil {
		return err
	}
	if len(expectedWithdrawals) == int(spec.MAX_WITHDRAWALS_PER_PAYLOAD) {
		nextIndex, err := state.NextWithdrawalValidatorIndex()
		if err != nil {
			return err
		}
		nextIndex = common.ValidatorIndex(uint64(nextIndex+1) % validatorCount)
		if err = state.SetNextWithdrawalValidatorIndex(nextIndex); err != nil {
			return err
		}
	} else {
		nextIndex, err := state.NextWithdrawalValidatorIndex()
		if err != nil {
			return err
		}
		nextIndex = common.ValidatorIndex((uint64(nextIndex) + uint64(spec.MAX_VALIDATORS_PER_WITHDRAWALS_SWEEP)) % validatorCount)
		if err = state.SetNextWithdrawalValidatorIndex(nextIndex); err != nil {
			return err
		}
	}
	return nil
}
