package randao

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/core"
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
type RandaoMixes struct { *VectorView }

var RandaoMixesType = VectorType(Bytes32Type, uint64(EPOCHS_PER_HISTORICAL_VECTOR))

// Provides a source of randomness for the state, for e.g. shuffling
func (mixes *RandaoMixes) GetRandomMix(epoch Epoch) (Root, error) {
	return RootReadProp(PropReader(mixes, uint64(epoch%EPOCHS_PER_HISTORICAL_VECTOR))).Root()
}

func (mixes *RandaoMixes) SetRandomMix(epoch Epoch, mix Root) error {
	return RootWriteProp(PropWriter(mixes, uint64(epoch%EPOCHS_PER_HISTORICAL_VECTOR))).SetRoot(mix)
}

// Prepare the randao mix for the given epoch by copying over the mix from the privious epoch.
func (mixes *RandaoMixes) PrepareRandao(epoch Epoch) error {
	prev, err := mixes.GetRandomMix(epoch.Previous())
	if err != nil {
		return err
	}
	return mixes.SetRandomMix(epoch, prev)
}

func SeedRandao(seed Root) (*RandaoMixes, error) {
	c := &tree.Commit{}
	filler := seed
	c.ExpandInplaceDepth(&filler, tree.GetDepth(uint64(EPOCHS_PER_HISTORICAL_VECTOR)))
	v, err := RandaoMixesType.ViewFromBacking(c)
	if err != nil {
		return nil, err
	}
	vecView, ok := v.(*VectorView)
	if !ok {
		return nil, errors.New("expected vector view from RandaoMixesType")
	}
	return &RandaoMixes{VectorView: vecView}, nil
}

type RandaoMixesReadProp VectorReadProp

func (p RandaoMixesReadProp) RandaoMixes() (*RandaoMixes, error) {
	v, err := VectorReadProp(p).Vector()
	if err != nil {
		return nil, err
	}
	return &RandaoMixes{VectorView: v}, nil
}

func (p *RandaoMixesReadProp) GetRandomMix(epoch Epoch) (Root, error) {
	mixes, err := p.RandaoMixes()
	if err != nil {
		return Root{}, err
	}
	return mixes.GetRandomMix(epoch)
}

type RandaoMixesWriteProp VectorWriteProp

func (p RandaoMixesWriteProp) SetRandaoMixes(v *RandaoMixes) error {
	return p(v)
}

func (p RandaoMixesWriteProp) SeedRandao(seed Root) error {
	mixes, err := SeedRandao(seed)
	if err != nil {
		return err
	}
	return p.SetRandaoMixes(mixes)
}

type RandaoMixesMutProp struct {
	RandaoMixesReadProp
	RandaoMixesWriteProp
}

func (p *RandaoMixesMutProp) SetRandomMix(epoch Epoch, mix Root) error {
	mixes, err := p.RandaoMixes()
	if err != nil {
		return err
	}
	if err := mixes.SetRandomMix(epoch, mix); err != nil {
		return err
	}
	return p.SetRandaoMixes(mixes)
}

func (p *RandaoMixesMutProp) PrepareRandao(epoch Epoch) error {
	mixes, err := p.RandaoMixes()
	if err != nil {
		return err
	}
	if err := mixes.PrepareRandao(epoch); err != nil {
		return err
	}
	return p.SetRandaoMixes(mixes)
}

var RandaoEpochSSZ = zssz.GetSSZ((*Epoch)(nil))

type RandaoFeature struct {
	State *RandaoMixesMutProp
	Meta  interface {
		meta.Versioning
		meta.Proposers
		meta.Pubkeys
		meta.SigDomain
	}
}

func (f *RandaoFeature) ProcessRandaoReveal(reveal BLSSignature) error {
	slot, err := f.Meta.CurrentSlot()
	if err != nil {
		return err
	}
	propIndex, err := f.Meta.GetBeaconProposerIndex(slot)
	if err != nil {
		return err
	}
	proposerPubkey, err := f.Meta.Pubkey(propIndex)
	if err != nil {
		return err
	}
	epoch := slot.ToEpoch()
	domain, err := f.Meta.GetDomain(DOMAIN_RANDAO, epoch)
	if err != nil {
		return err
	}
	// Verify RANDAO reveal
	if !bls.BlsVerify(proposerPubkey, ssz.HashTreeRoot(epoch, RandaoEpochSSZ), reveal, domain) {
		return errors.New("randao invalid")
	}
	// Mix in RANDAO reveal
	randMix, err := f.State.GetRandomMix(epoch)
	mix := XorBytes32(randMix, Hash(reveal[:]))
	return f.State.SetRandomMix(epoch%EPOCHS_PER_HISTORICAL_VECTOR, mix)
}
