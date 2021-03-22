package benches

import (
	"encoding/binary"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/configs"
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

func CreateTestState(validatorCount uint64, balance common.Gwei) (*phase0.BeaconStateView, *common.EpochsContext) {
	out, epc, err := phase0.KickStartState(configs.Mainnet, common.Root{123}, 1564000000, CreateTestValidators(validatorCount, balance))
	if err != nil {
		panic(err)
	}
	return out, epc
}
