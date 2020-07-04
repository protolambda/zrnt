package bls

import (
	"encoding/hex"
	"errors"
	"fmt"
	hbls "github.com/herumi/bls-eth-go-binary/bls"
)

type BLSPubkey [48]byte

func (p BLSPubkey) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(p[:])), nil
}

func (p BLSPubkey) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *BLSPubkey) UnmarshalText(text []byte) error {
	if p == nil {
		return errors.New("cannot decode into nil BLSPubkey")
	}
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	if len(text) != 96 {
		return fmt.Errorf("unexpected length string '%s'", string(text))
	}
	_, err := hex.Decode(p[:], text)
	return err
}

type BLSSignature [96]byte

func (p BLSSignature) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(p[:])), nil
}

func (p BLSSignature) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *BLSSignature) UnmarshalText(text []byte) error {
	if p == nil {
		return errors.New("cannot decode into nil BLSSignature")
	}
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	if len(text) != 192 {
		return fmt.Errorf("unexpected length string '%s'", string(text))
	}
	_, err := hex.Decode(p[:], text)
	return err
}

type CachedPubkey struct {
	Compressed   BLSPubkey
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
