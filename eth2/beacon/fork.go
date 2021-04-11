package beacon

import (
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/codec"
	"io"
)

type ForkDecoder struct {
	Spec    *common.Spec
	Genesis common.ForkDigest
	Altair  common.ForkDigest
	// TODO more forks
}

func NewForkDecoder(spec *common.Spec, genesisValRoot common.Root) *ForkDecoder {
	return &ForkDecoder{
		Spec:    spec,
		Genesis: common.ComputeForkDigest(spec.GENESIS_FORK_VERSION, genesisValRoot),
		Altair:  common.ComputeForkDigest(spec.ALTAIR_FORK_VERSION, genesisValRoot),
	}
}

func (d *ForkDecoder) DecodeBlock(digest common.ForkDigest,
	length uint64, r io.Reader) (*common.BeaconBlockEnvelope, error) {

	var block interface {
		common.EnvelopeBuilder
		common.SpecObj
	}

	switch digest {
	case d.Genesis:
		block = new(phase0.SignedBeaconBlock)
	case d.Altair:
		block = new(altair.SignedBeaconBlock)
	default:
		return nil, fmt.Errorf("unrecognized fork digest: %s", digest)
	}

	if err := block.Deserialize(d.Spec, codec.NewDecodingReader(r, length)); err != nil {
		return nil, err
	}
	return block.Envelope(d.Spec, digest), nil
}
