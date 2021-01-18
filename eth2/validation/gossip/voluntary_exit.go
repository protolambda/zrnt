package gossip

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon"
)

func (gv *GossipValidator) ValidateVoluntaryExit(ctx context.Context, volExit *beacon.VoluntaryExit) GossipValidatorResult {
	// TODO
	return GossipValidatorResult{ACCEPT, nil}
}
