package beacon

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/merge"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/beacon/sharding"
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

type OpaqueBlock interface {
	common.SpecObj
	common.EnvelopeBuilder
}

func (d *ForkDecoder) AllocBlock(digest common.ForkDigest) (OpaqueBlock, error) {
	switch digest {
	case d.Genesis:
		return new(phase0.SignedBeaconBlock), nil
	case d.Altair:
		return new(altair.SignedBeaconBlock), nil
	case d.Merge:
		return new(merge.SignedBeaconBlock), nil
	case d.Sharding:
		return new(sharding.SignedBeaconBlock), nil
	default:
		return nil, fmt.Errorf("unrecognized fork digest: %s", digest)
	}
}

type StandardUpgradeableBeaconState struct {
	common.BeaconState
}

func (s *StandardUpgradeableBeaconState) UpgradeMaybe(ctx context.Context, spec *common.Spec, epc *common.EpochsContext) error {
	slot, err := s.BeaconState.Slot()
	if err != nil {
		return err
	}
	if tpre, ok := s.BeaconState.(*phase0.BeaconStateView); ok && slot == common.Slot(spec.ALTAIR_FORK_EPOCH)*spec.SLOTS_PER_EPOCH {
		post, err := altair.UpgradeToAltair(spec, epc, tpre)
		if err != nil {
			return fmt.Errorf("failed to upgrade phase0 to altair state: %v", err)
		}
		if err := epc.LoadSyncCommittees(post); err != nil {
			return fmt.Errorf("failed to pre-compute sync committees: %v", err)
		}
		s.BeaconState = post
	}
	if tpre, ok := s.BeaconState.(*altair.BeaconStateView); ok && slot == common.Slot(spec.MERGE_FORK_EPOCH)*spec.SLOTS_PER_EPOCH {
		post, err := merge.UpgradeToMerge(spec, epc, tpre)
		if err != nil {
			return fmt.Errorf("failed to upgrade atalir to merge state: %v", err)
		}
		s.BeaconState = post
	}
	if slot == common.Slot(spec.SHARDING_FORK_EPOCH)*spec.SLOTS_PER_EPOCH {
		// TODO: upgrade
	}
	return nil
}

var _ common.UpgradeableBeaconState = (*StandardUpgradeableBeaconState)(nil)
