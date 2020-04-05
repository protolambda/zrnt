package beacon

import (
	"encoding/binary"
	"errors"
	"github.com/protolambda/zrnt/eth2/meta"

	"github.com/protolambda/zrnt/eth2/util/bls"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/props"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type RandaoProcessor interface {
	ProcessRandaoReveal(reveal BLSSignature) error
}

// Randomness and committees
type RandaoMixes struct{ *ComplexVectorView }

var RandaoMixesType = VectorType(Bytes32Type, uint64(EPOCHS_PER_HISTORICAL_VECTOR))

// Provides a source of randomness for the state, for e.g. shuffling
func (mixes *RandaoMixes) GetRandomMix(epoch Epoch) (Root, error) {
	return RootReadProp(PropReader(mixes, uint64(epoch%EPOCHS_PER_HISTORICAL_VECTOR))).Root()
}

func (mixes *RandaoMixes) SetRandomMix(epoch Epoch, mix Root) error {
	return RootWriteProp(PropWriter(mixes, uint64(epoch%EPOCHS_PER_HISTORICAL_VECTOR))).SetRoot(mix)
}

// Prepare the randao mix for the given epoch by copying over the mix from the previous epoch.
func (mixes *RandaoMixes) PrepareRandao(epoch Epoch) error {
	prev, err := mixes.GetRandomMix(epoch.Previous())
	if err != nil {
		return err
	}
	return mixes.SetRandomMix(epoch, prev)
}

func SeedRandao(seed Root, hook BackingHook) (*RandaoMixes, error) {
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
	return &RandaoMixes{ComplexVectorView: vecView}, nil
}

type RandaoMixesProp ComplexVectorProp

func (p RandaoMixesProp) RandaoMixes() (*RandaoMixes, error) {
	v, err := ComplexVectorProp(p).Vector()
	if err != nil {
		return nil, err
	}
	return &RandaoMixes{ComplexVectorView: v}, nil
}

func (p *RandaoMixesProp) GetRandomMix(epoch Epoch) (Root, error) {
	mixes, err := p.RandaoMixes()
	if err != nil {
		return Root{}, err
	}
	return mixes.GetRandomMix(epoch)
}

func (p *RandaoMixesProp) SetRandomMix(epoch Epoch, mix Root) error {
	mixes, err := p.RandaoMixes()
	if err != nil {
		return err
	}
	return mixes.SetRandomMix(epoch, mix)
}

func (p *RandaoMixesProp) PrepareRandao(epoch Epoch) error {
	mixes, err := p.RandaoMixes()
	if err != nil {
		return err
	}
	return mixes.PrepareRandao(epoch)
}

var RandaoEpochSSZ = zssz.GetSSZ((*Epoch)(nil))

func (state *RandaoMixesProp) ProcessRandaoReveal(input RandaoProcessInput, reveal BLSSignature) error {
	slot, err := input.CurrentSlot()
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
	domain, err := input.GetDomain(DOMAIN_RANDAO, epoch)
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
	// Mix in RANDAO reveal
	randMix, err := state.GetRandomMix(epoch)
	mix := XorBytes32(randMix, Hash(reveal[:]))
	return state.SetRandomMix(epoch%EPOCHS_PER_HISTORICAL_VECTOR, mix)
}

func (state *RandaoMixesProp) GetSeed(epoch Epoch, domainType BLSDomainType) (Root, error) {
	buf := make([]byte, 4+8+32)

	// domain type
	copy(buf[0:4], domainType[:])

	// epoch
	binary.LittleEndian.PutUint64(buf[4:4+8], uint64(epoch))

	// Avoid underflow
	mix, err := state.GetRandomMix(epoch + EPOCHS_PER_HISTORICAL_VECTOR - MIN_SEED_LOOKAHEAD - 1)
	if err != nil {
		return Root{}, err
	}
	copy(buf[4+8:], mix[:])

	return Hash(buf), nil
}
