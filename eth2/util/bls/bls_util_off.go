// +build bls_off

package bls

import . "github.com/protolambda/zrnt/eth2/core"

const BLS_ACTIVE = false

func BlsVerify(pubkey BLSPubkey, messageHash Root, signature BLSSignature, domain BLSDomain) bool {
	// BLS OFF: just allow it.
	return true
}

func BlsAggregatePubkeys(pubkeys []BLSPubkeyNode) BLSPubkeyNode {
	// BLS OFF: just return an empty key
	return BLSPubkeyNode{}
}

func BlsVerifyMultiple(pubkeys []BLSPubkeyNode, messageHashes []Root, signature BLSSignatureNode, domain BLSDomain) bool {
	// BLS OFF: just allow it.
	return true
}
