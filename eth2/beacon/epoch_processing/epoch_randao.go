package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
)

func ProcessEpochRandao(state *BeaconState) {
	state.LatestRandaoMixes[(state.Epoch()+1)%LATEST_RANDAO_MIXES_LENGTH] = state.GetRandaoMix(state.Epoch())
}
