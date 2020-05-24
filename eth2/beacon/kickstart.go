package beacon

import (
	"errors"
	hbls "github.com/herumi/bls-eth-go-binary/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

type KickstartValidatorData struct {
	Pubkey                BLSPubkey
	WithdrawalCredentials Root
	Balance               Gwei
}

// To build a genesis state without Eth 1.0 deposits, i.e. directly from a sequence of minimal validator data.
func KickStartState(eth1BlockHash Root, time Timestamp, validators []KickstartValidatorData) (*BeaconStateView, *EpochsContext, error) {
	deps := make([]Deposit, len(validators), len(validators))

	for i := range validators {
		v := &validators[i]
		d := &deps[i]
		d.Data = DepositData{
			Pubkey:                v.Pubkey,
			WithdrawalCredentials: v.WithdrawalCredentials,
			Amount:                v.Balance,
			Signature:             BLSSignature{},
		}
	}

	state, epc, err := GenesisFromEth1(eth1BlockHash, 0, deps, true)
	if err != nil {
		return nil, nil, err
	}
	if err := state.SetGenesisTime(time); err != nil {
		return nil, nil, err
	}
	return state, epc, nil
}

// To build a genesis state without Eth 1.0 deposits, i.e. directly from a sequence of minimal validator data.
func KickStartStateWithSignatures(eth1BlockHash Root, time Timestamp, validators []KickstartValidatorData, keys [][32]byte) (*BeaconStateView, *EpochsContext, error) {
	deps := make([]Deposit, len(validators), len(validators))

	for i := range validators {
		v := &validators[i]
		d := &deps[i]
		d.Data = DepositData{
			Pubkey:                v.Pubkey,
			WithdrawalCredentials: v.WithdrawalCredentials,
			Amount:                v.Balance,
			Signature:             BLSSignature{},
		}
		root := ssz.HashTreeRoot(d.Data.ToMessage(), DepositMessageSSZ)
		var secKey hbls.SecretKey
		if err := secKey.Deserialize(keys[i][:]); err != nil {
			return nil, nil, err
		}
		dom := ComputeDomain(DOMAIN_DEPOSIT, GENESIS_FORK_VERSION, Root{})
		msg := ComputeSigningRoot(root, dom)
		sig := secKey.SignHash(msg[:])
		var p BLSPubkey
		copy(p[:], secKey.GetPublicKey().Serialize())
		if p != d.Data.Pubkey {
			return nil, nil, errors.New("privkey invalid, expected different pubkey")
		}
		copy(d.Data.Signature[:], sig.Serialize())
	}

	state, epc, err := GenesisFromEth1(eth1BlockHash, 0, deps, true)
	if err != nil {
		return nil, nil, err
	}
	if err := state.SetGenesisTime(time); err != nil {
		return nil, nil, err
	}
	return state, epc, nil
}
