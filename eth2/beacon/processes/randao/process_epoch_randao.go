package randao

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
)

func ProcessEpochRandao(state *beacon.BeaconState) {
	state.Latest_randao_mixes[(state.Epoch()+1)%beacon.LATEST_RANDAO_MIXES_LENGTH] = state.Get_randao_mix(state.Epoch())
}
