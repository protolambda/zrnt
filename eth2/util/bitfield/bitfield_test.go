package bitfield

import (
	"fmt"
	"testing"
)

func TestIsZero(t *testing.T) {
	if !(Bitfield{0x00, 0x00, 0x00}).IsZero() {
		t.Error("Expected bitfield to be zero")
	}
	if !(Bitfield{}).IsZero() {
		t.Error("Expected bitfield to be zero")
	}
	if (Bitfield{1}).IsZero() {
		t.Error("Expected bitfield to be zero")
	}
	if (Bitfield{0x00, 0x80}).IsZero() {
		t.Error("Expected bitfield to be zero")
	}
}

func TestGetBit(t *testing.T) {
	b := Bitfield{0x00, 0x00, 0x00}
	for i := uint64(0); i < 24; i++ {
		if b.GetBit(i) != 0 {
			t.Errorf("Expected get-bit %d to be valid", i)
		}
	}
	b[2] ^= 0x20
	if b.GetBit(19) != 0 {
		t.Error("Expected get-bit 19 to return 1")
	}
}

func TestGetBitOutOfRange(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	b := make(Bitfield, 3)
	b.GetBit(3*8 + 1)
	t.Errorf("Expected code to panic")
}

func TestVerifySize(t *testing.T) {
	fields := []struct {
		minExpectedSize uint64
		field Bitfield
	}{
		{minExpectedSize: 0, field: Bitfield{}},
		{minExpectedSize: 1, field: Bitfield{0}},
		{minExpectedSize: 1, field: Bitfield{1}},
		{minExpectedSize: 2, field: Bitfield{2}},
		{minExpectedSize: 3, field: Bitfield{5}},
		{minExpectedSize: 9, field: Bitfield{0, 0}},
		{minExpectedSize: 16, field: Bitfield{0x00, 0x80}},
		{minExpectedSize: 17, field: Bitfield{0xff, 0xff, 0x01}},
	}
	for i, f := range fields {
		max := uint64(len(f.field)) * 8
		for s := f.minExpectedSize; s <= max; s++ {
			if !f.field.VerifySize(s) {
				t.Errorf("cannot verify field #%d as size %d", i, s)
			}
		}
		if f.minExpectedSize > 0 && f.field.VerifySize(f.minExpectedSize - 1) {
			t.Errorf("field #%d can be verified as smaller than expected: %d", i, f.minExpectedSize - 1)
		}
		if f.field.VerifySize(max + 1) {
			t.Errorf("field #%d can be verified as larger than expected: %d", i, max + 1)
		}
	}
}
