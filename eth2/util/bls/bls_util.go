package bls

import "go-beacon-transition/eth2"

func Bls_verify(pubkey eth2.BLSPubkey, message_hash eth2.Root, signature eth2.BLSSignature, domain eth2.BLSDomain) bool {
	// TODO BLS verify single
	return true
}

func Bls_aggregate_pubkeys(pubkeys []eth2.BLSPubkey) eth2.BLSPubkey {
	// TODO aggregate pubkeys with BLS
	return eth2.BLSPubkey{}
}

func Bls_verify_multiple(pubkeys []eth2.BLSPubkey, message_hashes []eth2.Root, signature eth2.BLSSignature, domain eth2.BLSDomain) bool {
	// TODO BLS verify multiple
	return false
}
