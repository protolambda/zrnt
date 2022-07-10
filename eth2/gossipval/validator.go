package gossipval

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type GossipValidatorCode uint

const (
	// Unlike the enum in gossipsub, this defaults (0 value) to rejection.
	REJECT GossipValidatorCode = iota
	IGNORE
	ACCEPT
)

func (gvr GossipValidatorCode) String() string {
	switch gvr {
	case REJECT:
		return "REJECT"
	case IGNORE:
		return "IGNORE"
	case ACCEPT:
		return "ACCEPT"
	default:
		return "UNKNOWN"
	}
}

type GossipValidatorResult struct {
	Result GossipValidatorCode
	Err    error
}

func (gve GossipValidatorResult) Error() string {
	return fmt.Sprintf("%s: %s", gve.Result.String(), gve.Err.Error())
}

func (gve GossipValidatorResult) Unwrap() error {
	return gve.Err
}

type Spec interface {
	Spec() *common.Spec
}

type SlotAfter interface {
	// Returns the slot after the given duration elapsed. The duration may be negative. It clips on genesis.
	SlotAfter(delta time.Duration) common.Slot
}

type GenesisValidatorsRoot interface {
	GenesisValidatorsRoot() common.Root
}

type DomainGetter interface {
	GetDomain(typ common.BLSDomainType, epoch common.Epoch) (common.BLSDomain, error)
}

type BadBlockValidator interface {
	// If votes for this block should be rejected.
	IsBadBlock(root common.Root) bool
}

type HeadInfo interface {
	HeadInfo(ctx context.Context) (beacon.ChainEntry, *common.EpochsContext, common.BeaconState, error)
}

type Chain interface {
	Chain() beacon.Chain
}

// RetrieveHeadInfo is a util to implement the HeadInfo interface
func RetrieveHeadInfo(ctx context.Context, ch beacon.Chain) (beacon.ChainEntry, *common.EpochsContext, common.BeaconState, error) {
	headRef, err := ch.Head()
	if err != nil {
		return nil, nil, nil, errors.New("could not fetch head ref for validation")
	}
	epc, err := headRef.EpochsContext(ctx)
	if err != nil {
		return nil, nil, nil, errors.New("could not fetch head EPC for validation")
	}
	state, err := headRef.State(ctx)
	if err != nil {
		return nil, nil, nil, errors.New("could not fetch head state for validation")
	}
	return headRef, epc, state, nil
}
