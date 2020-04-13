// +build bls_off

package bls

const BLS_ACTIVE = false

func Verify(pubkey *CachedPubkey, message [32]byte, signature BLSSignature) bool {
	// TODO BLS verify single
	// Temporary: just allow it.
	return true
}

func FastAggregateVerify(pubkeys []*CachedPubkey, message [32]byte, signature BLSSignature) bool {
	// TODO BLS verify multiple
	// Temporary: just allow it.
	return true
}
