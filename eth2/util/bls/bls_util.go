package bls

func BlsVerify(pubkey BLSPubkey, messageHash [32]byte, signature BLSSignature, domain BLSDomain) bool {
	// TODO BLS verify single
	// Temporary: just allow it.
	return true
}

func BlsAggregatePubkeys(pubkeys []BLSPubkey) BLSPubkey {
	// TODO aggregate pubkeys with BLS
	// Temporary: just return an empty key (TODO: or is XOR better temporarily?)
	return BLSPubkey{}
}

func BlsVerifyMultiple(pubkeys []BLSPubkey, messageHashes [][32]byte, signature BLSSignature, domain BLSDomain) bool {
	// TODO BLS verify multiple
	// Temporary: just allow it.
	return true
}
