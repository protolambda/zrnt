package components

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bls"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

// Randomness and committees
type RandaoState struct {
	RandaoMixes [EPOCHS_PER_HISTORICAL_VECTOR]Root
}

func (state *RandaoState) GetRandaoMix(epoch Epoch) Root {
	return state.RandaoMixes[epoch%EPOCHS_PER_HISTORICAL_VECTOR]
}

// Provides a source of randomness for the state, for e.g. shuffling
func (state *RandaoState) GetRandomMix(epoch Epoch) Root {
	return state.GetRandaoMix(epoch)
}

// Prepare the randao mix for the given epoch by copying over the mix from the privious epoch.
func (state *BeaconState) PrepareRandao(epoch Epoch) {
	state.RandaoMixes[epoch%EPOCHS_PER_HISTORICAL_VECTOR] = state.GetRandaoMix(epoch.Previous())
}

type RandaoBlockData struct {
	RandaoReveal BLSSignature
}

var RandaoEpochSSZ = zssz.GetSSZ((*Epoch)(nil))

func (revealData *RandaoBlockData) Process(state *BeaconState) error {
	propIndex := state.GetBeaconProposerIndex()
	proposer := state.Validators[propIndex]
	currentEpoch := state.Epoch()
	if !bls.BlsVerify(
		proposer.Pubkey,
		ssz.HashTreeRoot(state.Epoch(), RandaoEpochSSZ),
		revealData.RandaoReveal,
		state.GetDomain(DOMAIN_RANDAO, currentEpoch),
	) {
		return errors.New("randao invalid")
	}
	state.RandaoMixes[state.Epoch()%EPOCHS_PER_HISTORICAL_VECTOR] = XorBytes32(
		state.GetRandaoMix(currentEpoch),
		Hash(revealData.RandaoReveal[:]))
	return nil
}
