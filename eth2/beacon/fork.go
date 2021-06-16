package beacon

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/merge"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/beacon/sharding"
	"github.com/protolambda/ztyp/codec"
	"io"
)

type ForkDecoder struct {
	Spec     *common.Spec
	Genesis  common.ForkDigest
	Altair   common.ForkDigest
	Merge    common.ForkDigest
	Sharding common.ForkDigest
	// TODO more forks
}

func NewForkDecoder(spec *common.Spec, genesisValRoot common.Root) *ForkDecoder {
	return &ForkDecoder{
		Spec:     spec,
		Genesis:  common.ComputeForkDigest(spec.GENESIS_FORK_VERSION, genesisValRoot),
		Altair:   common.ComputeForkDigest(spec.ALTAIR_FORK_VERSION, genesisValRoot),
		Merge:    common.ComputeForkDigest(spec.MERGE_FORK_VERSION, genesisValRoot),
		Sharding: common.ComputeForkDigest(spec.SHARDING_FORK_VERSION, genesisValRoot),
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
	case d.Merge:
		block = new(merge.SignedBeaconBlock)
	case d.Sharding:
		block = new(sharding.SignedBeaconBlock)
	default:
		return nil, fmt.Errorf("unrecognized fork digest: %s", digest)
	}

	if err := block.Deserialize(d.Spec, codec.NewDecodingReader(r, length)); err != nil {
		return nil, err
	}
	return block.Envelope(d.Spec, digest), nil
}

type StandardUpgradeableBeaconState struct {
	common.BeaconState
}

func (s *StandardUpgradeableBeaconState) UpgradeMaybe(ctx context.Context, spec *common.Spec, epc *common.EpochsContext) error {
	pre := s.BeaconState
	slot, err := pre.Slot()
	if err != nil {
		return err
	}
	switch pre.(type) {
	case *phase0.BeaconStateView:
		if slot == common.Slot(spec.ALTAIR_FORK_EPOCH)*spec.SLOTS_PER_EPOCH {
			// TODO: upgrade
		}
		return nil
	case *altair.BeaconStateView:
		if slot == common.Slot(spec.MERGE_FORK_EPOCH)*spec.SLOTS_PER_EPOCH {
			// TODO: upgrade
		}
		return nil
	case *merge.BeaconStateView:
		if slot == common.Slot(spec.SHARDING_FORK_EPOCH)*spec.SLOTS_PER_EPOCH {
			// TODO: upgrade
		}
		return nil
	default:
		return nil
	}
}

var _ common.UpgradeableBeaconState = (*StandardUpgradeableBeaconState)(nil)
