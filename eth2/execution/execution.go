package execution

import (
	"context"

	"github.com/protolambda/zrnt/eth2/beacon/bellatrix"
	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"
)

type NoOpExecutionEngine struct{}

func (n NoOpExecutionEngine) DenebNotifyNewPayload(ctx context.Context, executionPayload *deneb.ExecutionPayload, parentBeaconBlockRoot common.Root) (valid bool, err error) {
	return true, nil
}

func (n NoOpExecutionEngine) DenebIsValidVersionedHashes(ctx context.Context, payload *deneb.ExecutionPayload, versionedHashes []common.Hash32) (bool, error) {
	return true, nil
}

func (n NoOpExecutionEngine) DenebIsValidBlockHash(ctx context.Context, payload *deneb.ExecutionPayload, parentBeaconBlockRoot common.Root) (bool, error) {
	return true, nil
}

func (n NoOpExecutionEngine) CapellaNotifyNewPayload(ctx context.Context, executionPayload *capella.ExecutionPayload) (valid bool, err error) {
	return true, nil
}

func (n NoOpExecutionEngine) CapellaIsValidBlockHash(ctx context.Context, payload *capella.ExecutionPayload) (bool, error) {
	return true, nil
}

func (n NoOpExecutionEngine) BellatrixNotifyNewPayload(ctx context.Context, executionPayload *bellatrix.ExecutionPayload) (valid bool, err error) {
	return true, nil
}

func (n NoOpExecutionEngine) BellatrixIsValidBlockHash(ctx context.Context, payload *bellatrix.ExecutionPayload) (bool, error) {
	return true, nil
}

var _ bellatrix.ExecutionEngine = (*NoOpExecutionEngine)(nil)
var _ capella.ExecutionEngine = (*NoOpExecutionEngine)(nil)
var _ deneb.ExecutionEngine = (*NoOpExecutionEngine)(nil)

var _ common.ExecutionEngine = (*NoOpExecutionEngine)(nil)
