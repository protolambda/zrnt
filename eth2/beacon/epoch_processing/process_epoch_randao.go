package epoch_processing

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
)

func ProcessEpochRandao(state *beacon.BeaconState) {
	state.LatestRandaoMixes[(state.Epoch()+1)%beacon.LATEST_RANDAO_MIXES_LENGTH] = state.GetRandaoMix(state.Epoch())
}
