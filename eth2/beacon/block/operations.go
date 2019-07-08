package block

import (
	. "github.com/protolambda/zrnt/eth2/beacon/block/operations"
	. "github.com/protolambda/zrnt/eth2/beacon/components"
)

type BlockOperations struct {
	ProposerSlashings ProposerSlashings
	AttesterSlashings AttesterSlashings
	Attestations      Attestations
	Deposits          Deposits
	VoluntaryExits    VoluntaryExits
	Transfers         Transfers
}

func (ops *BlockOperations) Process(state *BeaconState) error {
	if err := ops.ProposerSlashings.Process(state); err != nil {
		return nil
	}
	if err := ops.AttesterSlashings.Process(state); err != nil {
		return nil
	}
	if err := ops.Deposits.Process(state); err != nil {
		return nil
	}
	if err := ops.VoluntaryExits.Process(state); err != nil {
		return nil
	}
	if err := ops.Transfers.Process(state); err != nil {
		return nil
	}
	return nil
}
