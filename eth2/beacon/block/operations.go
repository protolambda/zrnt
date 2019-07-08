package block

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/beacon/block/operations"
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
)

type BlockOperations struct {
	ProposerSlashings []ProposerSlashing
	AttesterSlashings []AttesterSlashing
	Attestations      []Attestation
	Deposits          []Deposit
	VoluntaryExits    []VoluntaryExit
	Transfers         []Transfer
}

func (ops *BlockOperations) Process(state *BeaconState) error {
	depositCount := DepositIndex(len(ops.Deposits))
	expectedCount := state.LatestEth1Data.DepositCount - state.DepositIndex
	if expectedCount > MAX_DEPOSITS {
		expectedCount = MAX_DEPOSITS
	}
	if depositCount != expectedCount {
		return errors.New("block does not contain expected deposits amount")
	}
	// check if all transfers are distinct
	distinctionCheckSet := make(map[BLSSignature]struct{})
	for i, v := range ops.Transfers {
		if existing, ok := distinctionCheckSet[v.Signature]; ok {
			return fmt.Errorf("transfer %d is the same as transfer %d, aborting", i, existing)
		}
		distinctionCheckSet[v.Signature] = struct{}{}
	}

}
