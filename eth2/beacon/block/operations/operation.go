package operations

import . "github.com/protolambda/zrnt/eth2/beacon/components"

type Operation interface {
	Process(state *BeaconState) error
}
