package phase0

import (
	"context"
	"errors"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/util/bls"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

// RandaoMixes is a EPOCHS_PER_HISTORICAL_VECTOR vector
type RandaoMixes []common.Root

func (a *RandaoMixes) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return tree.ReadRoots(dr, (*[]common.Root)(a), uint64(spec.EPOCHS_PER_HISTORICAL_VECTOR))
}

func (a RandaoMixes) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return tree.WriteRoots(w, a)
}

func (a RandaoMixes) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(spec.EPOCHS_PER_HISTORICAL_VECTOR) * 32
}

func (a *RandaoMixes) FixedLength(spec *common.Spec) uint64 {
	return uint64(spec.EPOCHS_PER_HISTORICAL_VECTOR) * 32
}

func (li RandaoMixes) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexVectorHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length)
}

func RandaoMixesType(spec *common.Spec) VectorTypeDef {
	return VectorType(common.Bytes32Type, uint64(spec.EPOCHS_PER_HISTORICAL_VECTOR))
}

// Randomness and committees
type RandaoMixesView struct{ *ComplexVectorView }

var _ common.RandaoMixes = (*RandaoMixesView)(nil)

func AsRandaoMixes(v View, err error) (*RandaoMixesView, error) {
	c, err := AsComplexVector(v, err)
	return &RandaoMixesView{c}, nil
}

// Provides a source of randomness for the state, for e.g. shuffling
func (mixes *RandaoMixesView) GetRandomMix(epoch common.Epoch) (common.Root, error) {
	i := uint64(epoch) % mixes.VectorLength
	return AsRoot(mixes.Get(i))
}

func (mixes *RandaoMixesView) SetRandomMix(epoch common.Epoch, mix common.Root) error {
	i := uint64(epoch) % mixes.VectorLength
	r := RootView(mix)
	return mixes.Set(i, &r)
}

func SeedRandao(spec *common.Spec, seed common.Root) (*RandaoMixesView, error) {
	filler := seed
	length := uint64(spec.EPOCHS_PER_HISTORICAL_VECTOR)
	c, err := tree.SubtreeFillToLength(&filler, tree.CoverDepth(length), length)
	if err != nil {
		return nil, err
	}
	v, err := RandaoMixesType(spec).ViewFromBacking(c, nil)
	if err != nil {
		return nil, err
	}
	vecView, ok := v.(*ComplexVectorView)
	if !ok {
		return nil, errors.New("expected vector view from RandaoMixesType")
	}
	return &RandaoMixesView{ComplexVectorView: vecView}, nil
}

func ProcessRandaoReveal(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, state common.BeaconState, reveal common.BLSSignature) error {
	select {
	case <-ctx.Done():
		return common.TransitionCancelErr
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
	domain, err := common.GetDomain(state, spec.DOMAIN_RANDAO, epoch)
	if err != nil {
		return err
	}
	// Verify RANDAO reveal
	if !bls.Verify(
		proposerPubkey,
		common.ComputeSigningRoot(
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
