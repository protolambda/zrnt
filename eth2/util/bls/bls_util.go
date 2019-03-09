package bls

import "github.com/protolambda/go-beacon-transition/eth2/beacon"

func BlsVerify(pubkey beacon.BLSPubkey, messageHash beacon.Root, signature beacon.BLSSignature, domain beacon.BLSDomain) bool {
	// TODO BLS verify single
	// Temporary: just allow it.
	return true
}

func BlsAggregatePubkeys(pubkeys []beacon.BLSPubkey) beacon.BLSPubkey {
	// TODO aggregate pubkeys with BLS
	// Temporary: just return an empty key (TODO: or is XOR better temporarily?)
	return beacon.BLSPubkey{}
}

func BlsVerifyMultiple(pubkeys []beacon.BLSPubkey, messageHashes []beacon.Root, signature beacon.BLSSignature, domain beacon.BLSDomain) bool {
	// TODO BLS verify multiple
	// Temporary: just allow it.
	return true
}
