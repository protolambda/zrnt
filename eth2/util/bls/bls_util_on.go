// +build !bls_off

package bls

import (
	hbls "github.com/herumi/bls-eth-go-binary/bls"
)

func init() {
	if err := hbls.Init(hbls.BLS12_381); err != nil {
		panic(err)
	}
	if err := hbls.SetETHmode(3); err != nil { // draft 7
		panic(err)
	}
}

var G2_POINT_AT_INFINITY = BLSSignature{0: 0xc0}

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

func parsePubkeys(pubkeys []*CachedPubkey) ([]hbls.PublicKey, error) {
	pubs := make([]hbls.PublicKey, len(pubkeys), len(pubkeys))
	for i, p := range pubkeys {
		pub, err := p.Pubkey()
		if err != nil {
			return nil, err
		}
		pubs[i] = *pub
	}
	return pubs, nil
}

func Eth2FastAggregateVerify(pubkeys []*CachedPubkey, message [32]byte, signature BLSSignature) bool {
	pubs, err := parsePubkeys(pubkeys)
	if err != nil {
		return false
	}
	if len(pubs) == 0 {
		if signature == G2_POINT_AT_INFINITY {
			return true
		} else {
			// If it's not G2_POINT_AT_INFINITY,
			// then don't use Herumi BLS to verify something unnecessarily.
			// And the 0 length pubkeys would panic in Herumi BLS.
			return false
		}
	}

	var parsedSig hbls.Sign
	if err := parsedSig.Deserialize(signature[:]); err != nil {
		return false
	}

	return parsedSig.FastAggregateVerify(pubs, message[:])
}
