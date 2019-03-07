package bls

import "github.com/protolambda/go-beacon-transition/eth2/beacon"

func Bls_verify(pubkey beacon.BLSPubkey, message_hash beacon.Root, signature beacon.BLSSignature, domain beacon.BLSDomain) bool {
	// TODO BLS verify single
	// Temporary: just allow it.
	return true
}

func Bls_aggregate_pubkeys(pubkeys []beacon.BLSPubkey) beacon.BLSPubkey {
	// TODO aggregate pubkeys with BLS
	// Temporary: just return an empty key (TODO: or is XOR better temporarily?)
	return beacon.BLSPubkey{}
}

func Bls_verify_multiple(pubkeys []beacon.BLSPubkey, message_hashes []beacon.Root, signature beacon.BLSSignature, domain beacon.BLSDomain) bool {
	// TODO BLS verify multiple
	// Temporary: just allow it.
	return true
}
