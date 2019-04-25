package block_processing

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/hash"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessBlockRandao(state *BeaconState, block *BeaconBlock) error {
	propIndex := state.GetBeaconProposerIndex()
	proposer := state.ValidatorRegistry[propIndex]
	if !bls.BlsVerify(
		proposer.Pubkey,
		ssz.HashTreeRoot(state.Epoch()),
		block.Body.RandaoReveal,
		state.GetDomain(DOMAIN_RANDAO, state.Epoch()),
	) {
		return errors.New("randao invalid")
	}
	state.LatestRandaoMixes[state.Epoch()%LATEST_RANDAO_MIXES_LENGTH] = hash.XorRoots(
		state.GetRandaoMix(state.Epoch()),
		hash.HashRoot(block.Body.RandaoReveal[:]))
	return nil
}
