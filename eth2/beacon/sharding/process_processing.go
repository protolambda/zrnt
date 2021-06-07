package sharding

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/tree"
)

func ProcessPendingShardConfirmations(ctx context.Context, spec *common.Spec, state *BeaconStateView) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	currentSlot, err := state.Slot()
	if err != nil {
		return err
	}
	currentEpoch := spec.SlotToEpoch(currentSlot)
	// Skip if `GENESIS_EPOCH` because no prior epoch to process.
	if currentEpoch == common.GENESIS_EPOCH {
		return nil
	}

	buffer, err := state.ShardBuffer()
	if err != nil {
		return err
	}

	hFn := tree.GetHashFn()
	defaultDataCommitmentMerkleRoot := DataCommitmentType.DefaultNode().MerkleRoot(hFn)

	previousEpoch := currentEpoch.Previous()
	previousEpochStartSlot, _ := spec.EpochStartSlot(previousEpoch)
	end := previousEpochStartSlot + spec.SLOTS_PER_EPOCH
	for slot := previousEpochStartSlot; slot < end; slot++ {
		bufferIndex := uint64(slot % spec.SHARD_STATE_MEMORY_SLOTS)
		column, err := buffer.Column(bufferIndex)
		if err != nil {
			return err
		}
		shardsCount, err := column.Length()
		if err != nil {
			return err
		}
		for shardIndex := common.Shard(0); shardIndex < common.Shard(shardsCount); shardIndex++ {
			committeeWork, err := column.GetWork(shardIndex)
			if err != nil {
				return err
			}
			status, err := committeeWork.Status()
			if err != nil {
				return err
			}
			selector, err := status.Selector()
			if err != nil {
				return err
			}
			if selector == SHARD_WORK_PENDING {
				pendingHeaders, err := AsPendingShardHeaders(status.Value())
				if err != nil {
					return err
				}
				count, err := pendingHeaders.Length()
				if err != nil {
					return err
				}
				if count == 0 {
					return fmt.Errorf("not enough pending headers in state for slot %d, shard %d", slot, shardIndex)
				}
				// find winning header
				winningHeader, err := pendingHeaders.Header(0)
				if err != nil {
					return err
				}
				if count > 1 {
					winningWeight, err := winningHeader.Weight()
					if err != nil {
						return err
					}
					for i := uint64(0); i < count; i++ {
						h, err := pendingHeaders.Header(i)
						if err != nil {
							return err
						}
						w, err := h.Weight()
						if err != nil {
							return err
						}
						if w > winningWeight {
							winningHeader = h
							winningWeight = w
						}
					}
				}
				commitmentV, err := winningHeader.Commitment()
				if err != nil {
					return err
				}
				if commitmentV.HashTreeRoot(hFn) == defaultDataCommitmentMerkleRoot {
					if err := status.Change(SHARD_WORK_UNCONFIRMED, nil); err != nil {
						return err
					}
				} else {
					winningCommitment, err := winningHeader.Commitment()
					if err != nil {
						return err
					}
					if err := status.Change(SHARD_WORK_CONFIRMED, winningCommitment); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func ChargeConfirmedShardFees(ctx context.Context, spec *common.Spec, state *BeaconStateView, epc *common.EpochsContext) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	gasPrice, err := state.ShardGasPrice()
	if err != nil {
		return err
	}
	newGasPrice := gasPrice
	slot, err := state.Slot()
	if err != nil {
		return err
	}
	currentEpoch := spec.SlotToEpoch(slot)
	previousEpoch := currentEpoch.Previous()
	previousEpochStartSlot, _ := spec.EpochStartSlot(previousEpoch)
	activeShards := spec.ActiveShardCount(previousEpoch)
	adjustmentQuotient := activeShards * uint64(spec.SLOTS_PER_EPOCH) * spec.GASPRICE_ADJUSTMENT_COEFFICIENT

	buffer, err := state.ShardBuffer()
	if err != nil {
		return err
	}
	balances, err := state.Balances()
	if err != nil {
		return err
	}

	// Iterate through confirmed shard-headers
	end := previousEpochStartSlot + spec.SLOTS_PER_EPOCH
	for slot := previousEpochStartSlot; slot < end; slot++ {
		bufferIndex := uint64(slot % spec.SHARD_STATE_MEMORY_SLOTS)
		column, err := buffer.Column(bufferIndex)
		if err != nil {
			return err
		}
		shardsCount, err := column.Length()
		if err != nil {
			return err
		}
		for shardIndex := common.Shard(0); shardIndex < common.Shard(shardsCount); shardIndex++ {
			committeeWork, err := column.GetWork(shardIndex)
			if err != nil {
				return err
			}
			status, err := committeeWork.Status()
			if err != nil {
				return err
			}
			selector, err := status.Selector()
			if err != nil {
				return err
			}
			if selector == SHARD_WORK_CONFIRMED {
				commitment, err := AsDataCommitment(status.Value())
				if err != nil {
					return err
				}
				// Charge EIP 1559 fee
				proposer, err := epc.GetShardProposer(slot, shardIndex)
				if err != nil {
					return err
				}
				length, err := commitment.Length()
				if err != nil {
					return err
				}
				fee := common.Gwei(uint64(gasPrice) * length / spec.TARGET_SAMPLES_PER_BLOCK)
				if err := common.DecreaseBalance(balances, proposer, fee); err != nil {
					return err
				}

				// Track updated gas price
				newGasPrice = spec.ComputeUpdatedGasPrice(newGasPrice, length, adjustmentQuotient)
			}
		}
	}
	if err := state.SetShardGasPrice(newGasPrice); err != nil {
		return err
	}
	return nil
}

func ResetPendingShardWork(ctx context.Context, spec *common.Spec, state *BeaconStateView) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// TODO
	return nil
}
