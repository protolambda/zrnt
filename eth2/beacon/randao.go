package beacon

import (
	"context"
	"encoding/binary"
	"errors"
	"github.com/protolambda/zrnt/eth2/util/bls"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

// RandaoMixes is a EPOCHS_PER_HISTORICAL_VECTOR vector
type RandaoMixes []Root

func (a *RandaoMixes) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	if Epoch(len(*a)) != spec.EPOCHS_PER_HISTORICAL_VECTOR {
		// re-use space if available (for recycling old state objects)
		if Epoch(cap(*a)) >= spec.EPOCHS_PER_HISTORICAL_VECTOR {
			*a = (*a)[:spec.EPOCHS_PER_HISTORICAL_VECTOR]
		} else {
			*a = make([]Root, spec.EPOCHS_PER_HISTORICAL_VECTOR, spec.EPOCHS_PER_HISTORICAL_VECTOR)
		}
	}
	return dr.Vector(func(i uint64) codec.Deserializable {
		return &(*a)[i]
	}, 32, uint64(spec.EPOCHS_PER_HISTORICAL_VECTOR))
}

func (a RandaoMixes) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.Vector(func(i uint64) codec.Serializable {
		return &a[i]
	}, 32, uint64(spec.EPOCHS_PER_HISTORICAL_VECTOR))
}

func (a RandaoMixes) ByteLength(spec *Spec) (out uint64) {
	return uint64(spec.EPOCHS_PER_HISTORICAL_VECTOR) * 32
}

func (a *RandaoMixes) FixedLength(spec *Spec) uint64 {
	return uint64(spec.EPOCHS_PER_HISTORICAL_VECTOR) * 32
}

func (li RandaoMixes) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	length := uint64(len(li))
	return hFn.ComplexVectorHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length)
}

func (c *Phase0Config) RandaoMixes() VectorTypeDef {
	return VectorType(Bytes32Type, uint64(c.EPOCHS_PER_HISTORICAL_VECTOR))
}

// Randomness and committees
type RandaoMixesView struct{ *ComplexVectorView }

func AsRandaoMixes(v View, err error) (*RandaoMixesView, error) {
	c, err := AsComplexVector(v, err)
	return &RandaoMixesView{c}, nil
}

// Provides a source of randomness for the state, for e.g. shuffling
func (mixes *RandaoMixesView) GetRandomMix(epoch Epoch) (Root, error) {
	i := uint64(epoch) % mixes.VectorLength
	return AsRoot(mixes.Get(i))
}

func (mixes *RandaoMixesView) SetRandomMix(epoch Epoch, mix Root) error {
	i := uint64(epoch) % mixes.VectorLength
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

func (spec *Spec) GetSeed(mixes *RandaoMixesView, epoch Epoch, domainType BLSDomainType) (Root, error) {
	buf := make([]byte, 4+8+32)

	// domain type
	copy(buf[0:4], domainType[:])

	// epoch
	binary.LittleEndian.PutUint64(buf[4:4+8], uint64(epoch))

	// Avoid underflow
	mix, err := mixes.GetRandomMix(epoch + spec.EPOCHS_PER_HISTORICAL_VECTOR - spec.MIN_SEED_LOOKAHEAD - 1)
	if err != nil {
		return Root{}, err
	}
	copy(buf[4+8:], mix[:])

	return Hash(buf), nil
}

func (spec *Spec) SeedRandao(seed Root) (*RandaoMixesView, error) {
	filler := seed
	length := uint64(spec.EPOCHS_PER_HISTORICAL_VECTOR)
	c, err := tree.SubtreeFillToLength(&filler, tree.CoverDepth(length), length)
	if err != nil {
		return nil, err
	}
	v, err := spec.RandaoMixes().ViewFromBacking(c, nil)
	if err != nil {
		return nil, err
	}
	vecView, ok := v.(*ComplexVectorView)
	if !ok {
		return nil, errors.New("expected vector view from RandaoMixesType")
	}
	return &RandaoMixesView{ComplexVectorView: vecView}, nil
}

func (spec *Spec) ProcessRandaoReveal(ctx context.Context, epc *EpochsContext, state *BeaconStateView, reveal BLSSignature) error {
	select {
	case <-ctx.Done():
		return TransitionCancelErr
	default: // Don't block.
		break
	}
	slot, err := state.Slot()
	if err != nil {
		return err
	}
	propIndex, err := epc.GetBeaconProposer(slot)
	if err != nil {
		return err
	}
	proposerPubkey, ok := epc.PubkeyCache.Pubkey(propIndex)
	if !ok {
		return errors.New("could not find pubkey of proposer")
	}
	epoch := spec.SlotToEpoch(slot)
	domain, err := state.GetDomain(spec.DOMAIN_RANDAO, epoch)
	if err != nil {
		return err
	}
	// Verify RANDAO reveal
	if !bls.Verify(
		proposerPubkey,
		ComputeSigningRoot(
			epoch.HashTreeRoot(tree.GetHashFn()),
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
	return mixes.SetRandomMix(epoch, mix)
}
