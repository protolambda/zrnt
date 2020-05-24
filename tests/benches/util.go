package benches

import (
	"encoding/binary"
	. "github.com/protolambda/zrnt/eth2/beacon"
)

func CreateTestValidators(count uint64, balance Gwei) []KickstartValidatorData {
	out := make([]KickstartValidatorData, 0, count)
	for i := uint64(0); i < count; i++ {
		pubkey := BLSPubkey{0xaa}
		binary.LittleEndian.PutUint64(pubkey[1:], i)
		withdrawalCred := Root{0xbb}
		binary.LittleEndian.PutUint64(withdrawalCred[1:], i)
		out = append(out, KickstartValidatorData{
			Pubkey:                pubkey,
			WithdrawalCredentials: withdrawalCred,
			Balance:               balance,
		})
	}
	return out
}

func CreateTestState(validatorCount uint64, balance Gwei) (*BeaconStateView, *EpochsContext) {
	out, epc, err := KickStartState(Root{123}, 1564000000, CreateTestValidators(validatorCount, balance))
	if err != nil {
		panic(err)
	}
	return out, epc
}
