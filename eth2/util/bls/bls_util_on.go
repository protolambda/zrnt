// +build !bls_off

package bls

import (
	hbls "github.com/herumi/bls-eth-go-binary/bls"
)

func init() {
	if err := hbls.Init(hbls.BLS12_381); err != nil {
		panic(err)
	}
	if err := hbls.SetETHmode(1); err != nil {
		panic(err)
	}
}

const BLS_ACTIVE = true

func Verify(pubkey *CachedPubkey, message [32]byte, signature BLSSignature) bool {
	parsedPubkey, err := pubkey.Pubkey()
	if err != nil {
		return false
	}
	var parsedSig hbls.Sign
	if err := parsedSig.Deserialize(signature[:]); err != nil {
		return false
	}

	return parsedSig.VerifyHash(parsedPubkey, message[:])
}

func parsePubkeys(pubkeys []*CachedPubkey) []hbls.PublicKey {
	pubs := make([]hbls.PublicKey, len(pubkeys), len(pubkeys))
	for i, p := range pubkeys {
		pub, err := p.Pubkey()
		if err != nil {
			return nil
		}
		pubs[i] = *pub
	}
	return pubs
}

func FastAggregateVerify(pubkeys []*CachedPubkey, message [32]byte, signature BLSSignature) bool {
	pubs := parsePubkeys(pubkeys)
	if len(pubs) == 0 { // also if parsePubkeys errors and returns nil
		return false
	}

	var parsedSig hbls.Sign
	if err := parsedSig.Deserialize(signature[:]); err != nil {
		return false
	}

	return parsedSig.FastAggregateVerify(pubs, message[:])
}
