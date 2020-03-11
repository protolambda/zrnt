// +build !bls_off

package bls

import (
	hbls "github.com/herumi/bls-eth-go-binary/bls"
	. "github.com/protolambda/zrnt/eth2/core"
)

func init() {
	hbls.Init(hbls.BLS12_381)
	hbls.SetETHmode(1)
}

const BLS_ACTIVE = true

func Verify(pubkey BLSPubkey, message [32]byte, signature BLSSignature) bool {
	var parsedPubkey hbls.PublicKey
	if err := parsedPubkey.Deserialize(pubkey[:]); err != nil {
		return false
	}
	var parsedSig hbls.Sign
	if err := parsedSig.Deserialize(signature[:]); err != nil {
		return false
	}

	return parsedSig.VerifyHash(&parsedPubkey, message[:])
}

func parsePubkeys(pubkeys []BLSPubkey) []hbls.PublicKey {
	pubs := make([]hbls.PublicKey, len(pubkeys), len(pubkeys))
	for i := range pubkeys {
		if err := pubs[i].Deserialize(pubkeys[i][:]); err != nil {
			panic(err)
		}
	}
	return pubs
}

func FastAggregateVerify(pubkeys []BLSPubkey, message [32]byte, signature BLSSignature) bool {
	pubs := parsePubkeys(pubkeys)
	if len(pubs) == 0 {
		return false
	}

	var parsedSig hbls.Sign
	if err := parsedSig.Deserialize(signature[:]); err != nil {
		return false
	}

	return parsedSig.FastAggregateVerify(pubs, message[:])
}
