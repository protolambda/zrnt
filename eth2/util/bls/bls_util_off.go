// +build bls_off

package bls

import . "github.com/protolambda/zrnt/eth2/core"

const BLS_ACTIVE = false

func Verify(pubkey BLSPubkey, message [32]byte, signature BLSSignature) bool {
	// TODO BLS verify single
	// Temporary: just allow it.
	return true
}

func FastAggregateVerify(pubkeys []BLSPubkey, message [32]byte, signature BLSSignature) bool {
	// TODO BLS verify multiple
	// Temporary: just allow it.
	return true
}
