package bls

import hbls "github.com/herumi/bls-eth-go-binary/bls"

type BLSPubkey [48]byte

type BLSSignature [96]byte

type CachedPubkey struct {
	Compressed BLSPubkey
	decompressed *hbls.PublicKey
}

func (c *CachedPubkey) Pubkey() (*hbls.PublicKey, error) {
	if c.decompressed == nil {
		var parsedPubkey hbls.PublicKey
		if err := parsedPubkey.Deserialize(c.Compressed[:]); err != nil {
			return nil, err
		}
		c.decompressed = &parsedPubkey
	}
	return c.decompressed, nil
}
