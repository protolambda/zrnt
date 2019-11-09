package randao

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type RandaoProcessor interface {
	ProcessRandaoReveal(reveal BLSSignature) error
}

// Randomness and committees
type RandaoState struct {
	RandaoMixes [EPOCHS_PER_HISTORICAL_VECTOR]Root
}

// Provides a source of randomness for the state, for e.g. shuffling
func (state *RandaoState) GetRandomMix(epoch Epoch) Root {
	return state.RandaoMixes[epoch%EPOCHS_PER_HISTORICAL_VECTOR]
}

// Prepare the randao mix for the given epoch by copying over the mix from the privious epoch.
func (state *RandaoState) PrepareRandao(epoch Epoch) {
	state.RandaoMixes[epoch%EPOCHS_PER_HISTORICAL_VECTOR] = state.GetRandomMix(epoch.Previous())
}

func (state *RandaoState) SeedRandao(seed Root) {
	for i := Epoch(0); i < EPOCHS_PER_HISTORICAL_VECTOR; i++ {
		state.RandaoMixes[i] = seed
	}
}

var RandaoEpochSSZ = zssz.GetSSZ((*Epoch)(nil))

type RandaoFeature struct {
	State *RandaoState
	Meta  interface {
		meta.Versioning
		meta.Proposers
		meta.Pubkeys
	}
}

func (f *RandaoFeature) ProcessRandaoReveal(reveal BLSSignature) error {
	slot := f.Meta.CurrentSlot()
	propIndex := f.Meta.GetBeaconProposerIndex(slot)
	proposerPubkey := f.Meta.Pubkey(propIndex)
	epoch := slot.ToEpoch()
	// Verify RANDAO reveal
	if !bls.BlsVerify(
		proposerPubkey,
		ssz.HashTreeRoot(epoch, RandaoEpochSSZ),
		reveal,
		f.Meta.GetDomain(DOMAIN_RANDAO, epoch),
	) {
		return errors.New("randao invalid")
	}
	// Mix in RANDAO reveal
	mix := XorBytes32(f.State.GetRandomMix(epoch), Hash(reveal[:]))
	f.State.RandaoMixes[epoch%EPOCHS_PER_HISTORICAL_VECTOR] = mix
	return nil
}
