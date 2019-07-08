package block

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
)

type TransitionProcess interface {
	Process(state *BeaconState) error
}

func (block *BeaconBlock) Transition(state *BeaconState) error {
	if err := block.Header().Process(state); err != nil {
		return err
	}
	if err := block.Body.Process(state); err != nil {
		return err
	}
	return nil
}
