package merge

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

func ProcessExecutionPayload(ctx context.Context, spec *common.Spec, state *BeaconStateView, executionPayload *common.ExecutionPayload, engine common.ExecutionEngine) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if engine == nil {
		return errors.New("nil execution engine")
	}
	if completed, err := state.IsTransitionCompleted(); err != nil {
		return err
	} else if completed {
		latestExecHeader, err := state.LatestExecutionPayloadHeader()
		if err != nil {
			return err
		}
		prevHash, err := latestExecHeader.BlockHash()
		if err != nil {
			return err
		}
		if executionPayload.ParentHash != prevHash {
			return fmt.Errorf("expected parent hash %s in execution payload, but got %s",
				prevHash, executionPayload.ParentHash)
		}
		prevNumber, err := latestExecHeader.Number()
		if err != nil {
			return err
		}
		if executionPayload.Number != prevNumber+1 {
			return fmt.Errorf("expected number %d in execution payload, but got %d",
				prevNumber+1, executionPayload.Number)
		}
	}

	slot, err := state.Slot()
	if err != nil {
		return err
	}
	genesisTime, err := state.GenesisTime()
	if err != nil {
		return err
	}
	if expectedTime, err := spec.TimeAtSlot(slot, genesisTime); err != nil {
		return fmt.Errorf("slot or genesis time in state is corrupt, cannot compute time: %v", err)
	} else if executionPayload.Timestamp != expectedTime {
		return fmt.Errorf("state at slot %d, genesis time %d, expected execution payload time %d, but got %d",
			slot, genesisTime, expectedTime, executionPayload.Timestamp)
	}

	if success, err := engine.NewBlock(ctx, executionPayload); err != nil {
		return fmt.Errorf("unexpected problem in execution engine when inserting block %s (height %d), err: %v",
			executionPayload.BlockHash, executionPayload.Number, err)
	} else if !success {
		return fmt.Errorf("cannot process NewBlock in execution engine: %s (height %d)",
			executionPayload.BlockHash, executionPayload.Number)
	}

	return state.SetLatestExecutionPayloadHeader(executionPayload.Header(spec))
}
