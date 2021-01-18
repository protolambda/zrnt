package gossip

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon"
)

func (gv *GossipValidator) ValidateAttesterSlashing(ctx context.Context, attSl *beacon.AttesterSlashing) GossipValidatorResult {
	// TODO
	return GossipValidatorResult{ACCEPT, nil}
}
