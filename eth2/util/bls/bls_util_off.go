// +build bls_off

package bls

import . "github.com/protolambda/zrnt/eth2/core"

const BLS_ACTIVE = false

func BlsVerify(pubkey BLSPubkeyBytes, messageHash Root, signature BLSSignatureBytes, domain BLSDomain) bool {
	// BLS OFF: just allow it.
	return true
}

func BlsAggregatePubkeys(pubkeys []BLSPubkey) BLSPubkey {
	// BLS OFF: just return an empty key
	return BLSPubkey{}
}

func BlsVerifyMultiple(pubkeys []BLSPubkey, messageHashes []Root, signature BLSSignature, domain BLSDomain) bool {
	// BLS OFF: just allow it.
	return true
}
