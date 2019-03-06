package transition

import (
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
)

//  Return the randao mix at a recent epoch
func Get_randao_mix(state *beacon.BeaconState, epoch eth2.Epoch) eth2.Bytes32 {
	// Every usage is a trusted input (i.e. state is already up to date to handle the requested epoch number).
	// If something is wrong due to unforeseen usage, panic to catch it during development.
	if !(state.Epoch()-eth2.LATEST_RANDAO_MIXES_LENGTH < epoch && epoch <= state.Epoch()) {
		panic("cannot get randao mix for out-of-bounds epoch")
	}
	return state.Latest_randao_mixes[epoch%eth2.LATEST_RANDAO_MIXES_LENGTH]
}
