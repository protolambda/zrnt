package main

import (
	"encoding/binary"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/ztyp/tree"
)

func CreateTestValidators(count uint64, balance common.Gwei) []phase0.KickstartValidatorData {
	out := make([]phase0.KickstartValidatorData, 0, count)
	for i := uint64(0); i < count; i++ {
		pubkey := common.BLSPubkey{0xaa}
		binary.LittleEndian.PutUint64(pubkey[1:], i)
		withdrawalCred := common.Root{0xbb}
		binary.LittleEndian.PutUint64(withdrawalCred[1:], i)
		out = append(out, phase0.KickstartValidatorData{
			Pubkey:                pubkey,
			WithdrawalCredentials: withdrawalCred,
			Balance:               balance,
		})
	}
	return out
}

func CreateTestState(spec *common.Spec, validatorCount uint64, balance common.Gwei) (*phase0.BeaconStateView, *common.EpochsContext) {
	out, epc, err := phase0.KickStartState(spec, common.Root{123}, 1564000000, CreateTestValidators(validatorCount, balance))
	if err != nil {
		panic(err)
	}
	return out, epc
}

func main() {
	// can load other testnet configurations as well
	spec := configs.Mainnet

	state, epc := CreateTestState(spec, 1000, spec.MAX_EFFECTIVE_BALANCE)
	count, err := epc.GetCommitteeCountPerSlot(0)
	if err != nil {
		panic(err)
	}

	for i := common.Slot(0); i < spec.SLOTS_PER_EPOCH*2; i++ {
		fmt.Printf("slot %d\n", i)
		for j := uint64(0); j < count; j++ {
			committee, err := epc.GetBeaconCommittee(i, common.CommitteeIndex(j))
			if err != nil {
				panic(err)
			}
			fmt.Printf("slot %d, committee %d: %v\n", i, j, committee)
		}
	}

	root := state.HashTreeRoot(tree.GetHashFn())
	fmt.Printf("state root: %s\n", root)
}
