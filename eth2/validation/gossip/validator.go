package gossip

import (
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/chain"
	"time"
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

type GossipValidator struct {
	Spec  *beacon.Spec
	Chain chain.FullChain

	// Returns the slot after the given duration elapsed. The duration may be negative. It clips on genesis.
	SlotAfter func(delta time.Duration) beacon.Slot

	// Like BeaconState.GetDomain, but assuming only one canonical fork schedule is maintained.
	GetDomain func(dom beacon.BLSDomainType, epoch beacon.Epoch) (beacon.BLSDomain, error)
}
