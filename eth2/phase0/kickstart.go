package phase0

import (
	"errors"
	"github.com/phoreproject/bls/g1pubs"
	"github.com/protolambda/zrnt/eth2/beacon/deposits"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

type KickstartValidatorData struct {
	Pubkey                BLSPubkey
	WithdrawalCredentials Root
	Balance               Gwei
}

// To build a genesis state without Eth 1.0 deposits, i.e. directly from a sequence of minimal validator data.
func KickStartState(eth1BlockHash Root, time Timestamp, validators []KickstartValidatorData) (*FullFeaturedState, error) {
	deps := make([]deposits.Deposit, len(validators), len(validators))

	for i := range validators {
		v := &validators[i]
		d := &deps[i]
		d.Data = deposits.DepositData{
			Pubkey:                v.Pubkey,
			WithdrawalCredentials: v.WithdrawalCredentials,
			Amount:                v.Balance,
			Signature:             BLSSignature{},
		}
	}

	state, err := GenesisFromEth1(eth1BlockHash, 0, deps, false)
	if err != nil {
		return nil, err
	}
	state.GenesisTime = time
	return state, nil
}

// To build a genesis state without Eth 1.0 deposits, i.e. directly from a sequence of minimal validator data.
func KickStartStateWithSignatures(eth1BlockHash Root, time Timestamp, validators []KickstartValidatorData, keys [][32]byte) (*FullFeaturedState, error) {
	deps := make([]deposits.Deposit, len(validators), len(validators))

	for i := range validators {
		v := &validators[i]
		d := &deps[i]
		d.Data = deposits.DepositData{
			Pubkey:                v.Pubkey,
			WithdrawalCredentials: v.WithdrawalCredentials,
			Amount:                v.Balance,
			Signature:             BLSSignature{},
		}
		root := ssz.HashTreeRoot(d.Data.Message(), deposits.DepositMessageSSZ)
		priv := g1pubs.DeserializeSecretKey(keys[i])
		dom := ComputeDomain(DOMAIN_DEPOSIT, Version{})
		sig := g1pubs.SignWithDomain(root, priv, dom)
		p := g1pubs.PrivToPub(priv).Serialize()
		if p != d.Data.Pubkey {
			return nil, errors.New("privkey invalid, expected different pubkey")
		}
		d.Data.Signature = sig.Serialize()
	}

	state, err := GenesisFromEth1(eth1BlockHash, 0, deps, false)
	if err != nil {
		return nil, err
	}
	state.GenesisTime = time
	return state, nil
}
