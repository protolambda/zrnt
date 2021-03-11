package common

import (
	"encoding/hex"
	"errors"
	"fmt"
)

type WithdrawalPrefix [1]byte

func (p WithdrawalPrefix) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(p[:])), nil
}

func (p WithdrawalPrefix) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *WithdrawalPrefix) UnmarshalText(text []byte) error {
	if p == nil {
		return errors.New("cannot decode into nil WithdrawalPrefix")
	}
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	if len(text) != 2 {
		return fmt.Errorf("unexpected length string '%s'", string(text))
	}
	_, err := hex.Decode(p[:], text)
	return err
}
