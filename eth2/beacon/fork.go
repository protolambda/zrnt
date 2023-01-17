package beacon

import (
	"context"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/bellatrix"
	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
)

type ForkDecoder struct {
	Spec      *common.Spec
	Genesis   common.ForkDigest
	Altair    common.ForkDigest
	Bellatrix common.ForkDigest
	Capella   common.ForkDigest
	Deneb     common.ForkDigest
	// TODO more forks
}

func NewForkDecoder(spec *common.Spec, genesisValRoot common.Root) *ForkDecoder {
	return &ForkDecoder{
		Spec:      spec,
		Genesis:   common.ComputeForkDigest(spec.GENESIS_FORK_VERSION, genesisValRoot),
		Altair:    common.ComputeForkDigest(spec.ALTAIR_FORK_VERSION, genesisValRoot),
		Bellatrix: common.ComputeForkDigest(spec.BELLATRIX_FORK_VERSION, genesisValRoot),
		Capella:   common.ComputeForkDigest(spec.CAPELLA_FORK_VERSION, genesisValRoot),
		Deneb:     common.ComputeForkDigest(spec.DENEB_FORK_VERSION, genesisValRoot),
	}
}

type OpaqueBlock interface {
	common.SpecObj
	common.EnvelopeBuilder
}

func (d *ForkDecoder) BlockAllocator(digest common.ForkDigest) (func() OpaqueBlock, error) {
	switch digest {
	case d.Genesis:
		return func() OpaqueBlock { return new(phase0.SignedBeaconBlock) }, nil
	case d.Altair:
		return func() OpaqueBlock { return new(altair.SignedBeaconBlock) }, nil
	case d.Bellatrix:
		return func() OpaqueBlock { return new(bellatrix.SignedBeaconBlock) }, nil
	case d.Capella:
		return func() OpaqueBlock { return new(capella.SignedBeaconBlock) }, nil
	case d.Deneb:
		return func() OpaqueBlock { return new(deneb.SignedBeaconBlock) }, nil
	default:
		return nil, fmt.Errorf("unrecognized fork digest: %s", digest)
	}
}

func (d *ForkDecoder) ForkDigest(epoch common.Epoch) common.ForkDigest {
	if epoch < d.Spec.ALTAIR_FORK_EPOCH {
		return d.Genesis
	} else if epoch < d.Spec.BELLATRIX_FORK_EPOCH {
		return d.Altair
	} else if epoch < d.Spec.CAPELLA_FORK_EPOCH {
		return d.Bellatrix
	} else {
		// only consider deneb if it's actually set equal or higher than capella, to ignore it if it's missing in a config.
		if d.Spec.DENEB_FORK_EPOCH >= d.Spec.CAPELLA_FORK_EPOCH && epoch >= d.Spec.DENEB_FORK_EPOCH {
			return d.Deneb
		}
		return d.Capella
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
	if tpre, ok := s.BeaconState.(*altair.BeaconStateView); ok && slot == common.Slot(spec.BELLATRIX_FORK_EPOCH)*spec.SLOTS_PER_EPOCH {
		post, err := bellatrix.UpgradeToBellatrix(spec, epc, tpre)
		if err != nil {
			return fmt.Errorf("failed to upgrade atalir to bellatrix state: %v", err)
		}
		s.BeaconState = post
	}
	if tpre, ok := s.BeaconState.(*bellatrix.BeaconStateView); ok && slot == common.Slot(spec.CAPELLA_FORK_EPOCH)*spec.SLOTS_PER_EPOCH {
		post, err := capella.UpgradeToCapella(spec, epc, tpre)
		if err != nil {
			return fmt.Errorf("failed to upgrade bellatrix to capella state: %v", err)
		}
		s.BeaconState = post
	}
	if tpre, ok := s.BeaconState.(*capella.BeaconStateView); ok && slot == common.Slot(spec.DENEB_FORK_EPOCH)*spec.SLOTS_PER_EPOCH {
		post, err := deneb.UpgradeToDeneb(spec, epc, tpre)
		if err != nil {
			return fmt.Errorf("failed to upgrade bellatrix to capella state: %v", err)
		}
		s.BeaconState = post
	}
	return nil
}

var _ common.UpgradeableBeaconState = (*StandardUpgradeableBeaconState)(nil)

func EnvelopeToSignedBeaconBlock(benv *common.BeaconBlockEnvelope) (common.SpecObj, error) {
	switch x := benv.Body.(type) {
	case *phase0.BeaconBlockBody:
		return &phase0.SignedBeaconBlock{
			Message: phase0.BeaconBlock{
				Slot:          benv.Slot,
				ProposerIndex: benv.ProposerIndex,
				ParentRoot:    benv.ParentRoot,
				StateRoot:     benv.StateRoot,
				Body:          *x,
			},
			Signature: benv.Signature,
		}, nil
	case *altair.BeaconBlockBody:
		return &altair.SignedBeaconBlock{
			Message: altair.BeaconBlock{
				Slot:          benv.Slot,
				ProposerIndex: benv.ProposerIndex,
				ParentRoot:    benv.ParentRoot,
				StateRoot:     benv.StateRoot,
				Body:          *x,
			},
			Signature: benv.Signature,
		}, nil
	case *bellatrix.BeaconBlockBody:
		return &bellatrix.SignedBeaconBlock{
			Message: bellatrix.BeaconBlock{
				Slot:          benv.Slot,
				ProposerIndex: benv.ProposerIndex,
				ParentRoot:    benv.ParentRoot,
				StateRoot:     benv.StateRoot,
				Body:          *x,
			},
			Signature: benv.Signature,
		}, nil
	case *capella.BeaconBlockBody:
		return &capella.SignedBeaconBlock{
			Message: capella.BeaconBlock{
				Slot:          benv.Slot,
				ProposerIndex: benv.ProposerIndex,
				ParentRoot:    benv.ParentRoot,
				StateRoot:     benv.StateRoot,
				Body:          *x,
			},
			Signature: benv.Signature,
		}, nil
	case *deneb.BeaconBlockBody:
		return &deneb.SignedBeaconBlock{
			Message: deneb.BeaconBlock{
				Slot:          benv.Slot,
				ProposerIndex: benv.ProposerIndex,
				ParentRoot:    benv.ParentRoot,
				StateRoot:     benv.StateRoot,
				Body:          *x,
			},
			Signature: benv.Signature,
		}, nil
	default:
		return nil, fmt.Errorf("cannot convert beacon block envelope to full signed block, unrecognized body type: %T", x)
	}
}
