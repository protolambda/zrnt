package beacon

import (
	"encoding/binary"
	"errors"


	"github.com/protolambda/zrnt/eth2/util/bls"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/props"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

var RandaoMixesType = VectorType(Bytes32Type, uint64(EPOCHS_PER_HISTORICAL_VECTOR))

// Randomness and committees
type RandaoMixesView struct{ *ComplexVectorView }

func AsRandaoMixes(v View, err error) (*RandaoMixesView, error) {
	c, err := AsComplexVector(v, err)
	return &RandaoMixesView{c}, nil
}

// Provides a source of randomness for the state, for e.g. shuffling
func (mixes *RandaoMixesView) GetRandomMix(epoch Epoch) (Root, error) {
	i := uint64(epoch%EPOCHS_PER_HISTORICAL_VECTOR)
	return AsRoot(mixes.Get(i))
}

func (mixes *RandaoMixesView) SetRandomMix(epoch Epoch, mix Root) error {
	i := uint64(epoch%EPOCHS_PER_HISTORICAL_VECTOR)
	r := RootView(mix)
	return mixes.Set(i, &r)
}

// Prepare the randao mix for the given epoch by copying over the mix from the previous epoch.
func (mixes *RandaoMixesView) PrepareRandao(epoch Epoch) error {
	prev, err := mixes.GetRandomMix(epoch.Previous())
	if err != nil {
		return err
	}
	return mixes.SetRandomMix(epoch, prev)
}

func (mixes *RandaoMixesView) GetSeed(epoch Epoch, domainType BLSDomainType) (Root, error) {
	buf := make([]byte, 4+8+32)

	// domain type
	copy(buf[0:4], domainType[:])

	// epoch
	binary.LittleEndian.PutUint64(buf[4:4+8], uint64(epoch))

	// Avoid underflow
	mix, err := mixes.GetRandomMix(epoch + EPOCHS_PER_HISTORICAL_VECTOR - MIN_SEED_LOOKAHEAD - 1)
	if err != nil {
		return Root{}, err
	}
	copy(buf[4+8:], mix[:])

	return Hash(buf), nil
}

func SeedRandao(seed Root, hook BackingHook) (*RandaoMixesView, error) {
	filler := seed
	length := uint64(EPOCHS_PER_HISTORICAL_VECTOR)
	c, err := tree.SubtreeFillToLength(&filler, tree.CoverDepth(length), length)
	if err != nil {
		return nil, err
	}
	v, err := RandaoMixesType.ViewFromBacking(c, hook)
	if err != nil {
		return nil, err
	}
	vecView, ok := v.(*ComplexVectorView)
	if !ok {
		return nil, errors.New("expected vector view from RandaoMixesType")
	}
	return &RandaoMixesView{ComplexVectorView: vecView}, nil
}

var RandaoEpochSSZ = zssz.GetSSZ((*Epoch)(nil))

func (state *BeaconStateView) ProcessRandaoReveal(reveal BLSSignature) error {
	slot, err := state.Slot()
	if err != nil {
		return err
	}
	propIndex, err := input.GetBeaconProposerIndex(slot)
	if err != nil {
		return err
	}
	proposerPubkey, err := input.Pubkey(propIndex)
	if err != nil {
		return err
	}
	epoch := slot.ToEpoch()
	domain, err := state.GetDomain(DOMAIN_RANDAO, epoch)
	if err != nil {
		return err
	}
	// Verify RANDAO reveal
	if !bls.Verify(
		proposerPubkey,
		ComputeSigningRoot(
			ssz.HashTreeRoot(epoch, RandaoEpochSSZ),
			domain),
		reveal,
	) {
		return errors.New("randao invalid")
	}
	mixes, err := state.RandaoMixes()
	if err != nil {
		return err
	}
	// Mix in RANDAO reveal
	randMix, err := mixes.GetRandomMix(epoch)
	if err != nil {
		return err
	}
	mix := XorBytes32(randMix, Hash(reveal[:]))
	return mixes.SetRandomMix(epoch%EPOCHS_PER_HISTORICAL_VECTOR, mix)
}
