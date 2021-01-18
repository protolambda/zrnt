package gossip

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon"
)

func (gv *GossipValidator) ValidateProposerSlashing(ctx context.Context, propSl *beacon.ProposerSlashing) GossipValidatorResult {
	// TODO
	return GossipValidatorResult{ACCEPT, nil}
}
