package core

import "github.com/protolambda/zrnt/eth2/util/hex"

type BLSPubkey [48]byte

func (v *BLSPubkey) String() string { return hex.EncodeHexStr((*v)[:]) }

type BLSSignature [96]byte

func (v *BLSSignature) String() string { return hex.EncodeHexStr((*v)[:]) }

// Mixed into a BLS domain to define its type
type BLSDomainType uint32

// BLS domain (64 bits): fork version (32 bits) concatenated with BLS domain type (32 bits)
type BLSDomain uint64
