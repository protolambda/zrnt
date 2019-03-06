package epoch_processing

import (
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
)

func ProcessRandao(state *beacon.BeaconState) {
	state.Latest_randao_mixes[(state.Epoch()+1)%eth2.LATEST_RANDAO_MIXES_LENGTH] = Get_randao_mix(state, current_epoch)
}
