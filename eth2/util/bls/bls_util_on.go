// +build !bls_off

package bls

import (
	phbls "github.com/phoreproject/bls/g1pubs"
	. "github.com/protolambda/zrnt/eth2/core"
)

const BLS_ACTIVE = true

func BlsVerify(pubkey BLSPubkeyBytes, messageHash Root, signature BLSSignatureBytes, domain BLSDomain) bool {
	pub, err := phbls.DeserializePublicKey(pubkey)
	if err != nil {
		return false
	}
	sig, err := phbls.DeserializeSignature(signature)
	if err != nil {
		return false
	}
	return phbls.VerifyWithDomain(messageHash, pub, sig, domain)
}

func BlsAggregatePubkeys(pubkeys []BLSPubkeyBytes) BLSPubkeyBytes {
	agpub := phbls.AggregatePublicKeys(parsePubkeys(pubkeys))
	return agpub.Serialize()
}

func parsePubkeys(pubkeys []BLSPubkeyBytes) []*phbls.PublicKey {
	pubs := make([]*phbls.PublicKey, 0, len(pubkeys))
	for i := range pubkeys {
		p, err := phbls.DeserializePublicKey(pubkeys[i])
		if err != nil {
			return nil
		}
		pubs = append(pubs, p)
	}
	return pubs
}

func BlsVerifyMultiple(pubkeys []BLSPubkeyBytes, messageHashes []Root, signature BLSSignatureBytes, domain BLSDomain) bool {
	if len(pubkeys) != len(messageHashes) {
		return false
	}
	pubs := parsePubkeys(pubkeys)
	if len(pubs) == 0 {
		return false
	}

	sig, err := phbls.DeserializeSignature(signature)
	if err != nil {
		return false
	}

	msgs := make([][32]byte, 0, len(messageHashes))
	for i := range messageHashes {
		msgs = append(msgs, messageHashes[i])
	}

	return sig.VerifyAggregateWithDomain(pubs, msgs, domain)
}
