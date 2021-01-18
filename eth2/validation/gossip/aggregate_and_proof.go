package gossip

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon"
)

func (gv *GossipValidator) ValidateAggregateAndProof(ctx context.Context, att *beacon.SignedAggregateAndProof,
	hasSeen AttestationSeenFn, isBadBlock IsBadBlockFn) GossipValidatorResult {
	// TODO
	return GossipValidatorResult{ACCEPT, nil}
}
