package randao

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
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
func (state *RandaoState) PrepareRandao(epoch Epoch) {
	state.RandaoMixes[epoch%EPOCHS_PER_HISTORICAL_VECTOR] = state.GetRandaoMix(epoch.Previous())
}

var RandaoEpochSSZ = zssz.GetSSZ((*Epoch)(nil))

type RandaoRevealReq interface {
	VersioningMeta
	ProposingMeta
	PubkeyMeta
}

func (state *RandaoState) ProcessRandaoReveal(meta RandaoRevealReq, reveal BLSSignature) error {
	epoch := meta.CurrentEpoch()
	propIndex := meta.GetBeaconProposerIndex()
	proposerPubkey := meta.Pubkey(propIndex)
	// Verify RANDAO reveal
	if !bls.BlsVerify(
		proposerPubkey,
		ssz.HashTreeRoot(epoch, RandaoEpochSSZ),
		reveal,
		meta.GetDomain(DOMAIN_RANDAO, epoch),
	) {
		return errors.New("randao invalid")
	}
	// Mix in RANDAO reveal
	mix := XorBytes32(state.GetRandaoMix(epoch), Hash(reveal[:]))
	state.RandaoMixes[epoch%EPOCHS_PER_HISTORICAL_VECTOR] = mix
	return nil
}
